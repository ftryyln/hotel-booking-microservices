package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/ftryyln/hotel-booking-microservices/pkg/config"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
)

const (
	// ModeWhitelist keeps existing explicit routes only.
	ModeWhitelist = "whitelist"
	// ModeProxyAll enables fallback routing for every service.
	ModeProxyAll = "proxy_all"
)

type proxyEngine struct {
	mode             string
	routes           []*route
	log              *zap.Logger
	timeout          time.Duration
	retries          int
	metrics          *gatewayMetrics
	transport        http.RoundTripper
	healthInterval   time.Duration
	healthClient     *http.Client
	readyCh          chan struct{}
	readyOnce        sync.Once
	upstreams        map[string]*upstreamTarget
	jwtSecret        string
	circuitWindow    time.Duration
	circuitThreshold float64
	circuitCooldown  time.Duration
}

type route struct {
	name         string
	prefix       string
	stripPrefix  bool
	rewrite      string
	requireAuth  bool
	authStrategy string
	methods      map[string]struct{}
	target       *upstreamTarget
	proxy        *httputil.ReverseProxy
}

type upstreamTarget struct {
	name   string
	url    *url.URL
	health string

	mu     sync.RWMutex
	status upstreamStatus
}

type upstreamStatus struct {
	Healthy           bool      `json:"healthy"`
	LastChecked       time.Time `json:"last_checked"`
	LastError         string    `json:"last_error,omitempty"`
	CircuitOpenUntil  time.Time `json:"circuit_open_until,omitempty"`
	RequestsInWindow  int       `json:"requests_in_window"`
	FailuresInWindow  int       `json:"failures_in_window"`
	WindowStartedAt   time.Time `json:"window_started_at"`
	UnhealthySince    time.Time `json:"unhealthy_since,omitempty"`
	ConsecutiveErrors int       `json:"consecutive_errors"`
}

type routeFile struct {
	Routes   []routeDefinition   `yaml:"routes"`
	Fallback *fallbackDefinition `yaml:"fallback"`
}

type routeDefinition struct {
	Name         string   `yaml:"name"`
	Prefix       string   `yaml:"prefix"`
	Upstream     string   `yaml:"upstream"`
	StripPrefix  bool     `yaml:"strip_prefix"`
	Rewrite      string   `yaml:"rewrite"`
	RequireAuth  bool     `yaml:"require_auth"`
	AuthStrategy string   `yaml:"auth_strategy"`
	HealthPath   string   `yaml:"health_path"`
	Methods      []string `yaml:"methods"`
}

type fallbackDefinition struct {
	BasePath   string                   `yaml:"base_path"`
	StripBase  bool                     `yaml:"strip_base"`
	HealthPath string                   `yaml:"health_path"`
	Mapping    map[string]fallbackRoute `yaml:"mapping"`
}

type fallbackRoute struct {
	Upstream     string `yaml:"upstream"`
	StripPrefix  bool   `yaml:"strip_prefix"`
	RequireAuth  bool   `yaml:"require_auth"`
	AuthStrategy string `yaml:"auth_strategy"`
	HealthPath   string `yaml:"health_path"`
}

func NewProxyEngine(cfg config.Config, log *zap.Logger) (*proxyEngine, error) {
	engine := &proxyEngine{
		mode:           cfg.GatewayMode,
		log:            log,
		timeout:        cfg.UpstreamTimeout,
		retries:        cfg.UpstreamRetries,
		metrics:        newGatewayMetrics(),
		healthInterval: cfg.HealthInterval,
		healthClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		readyCh:          make(chan struct{}),
		upstreams:        map[string]*upstreamTarget{},
		jwtSecret:        cfg.JWTSecret,
		circuitWindow:    cfg.CircuitWindow,
		circuitThreshold: cfg.CircuitThreshold,
		circuitCooldown:  cfg.CircuitCooldown,
	}

	engine.transport = &retryTransport{
		base: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          128,
			MaxIdleConnsPerHost:   32,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: cfg.UpstreamTimeout,
		},
		retries: cfg.UpstreamRetries,
	}

	if engine.mode == "" {
		engine.mode = ModeWhitelist
	}

	definitions, err := loadRouteDefinitions(cfg.RoutesFile)
	if err != nil {
		if engine.mode == ModeProxyAll {
			return nil, fmt.Errorf("proxy mode requires routes configuration: %w", err)
		}
		log.Warn("unable to load routes file, proxy_all disabled", zap.Error(err))
		definitions = nil
	}

	if err := engine.buildRoutes(definitions); err != nil {
		return nil, err
	}

	return engine, nil
}

func (p *proxyEngine) buildRoutes(defs []routeDefinition) error {
	if len(defs) == 0 {
		p.routes = nil
		return nil
	}

	for _, def := range defs {
		if def.Upstream == "" || def.Prefix == "" {
			continue
		}

		targetURL, err := url.Parse(def.Upstream)
		if err != nil {
			return fmt.Errorf("invalid upstream for prefix %s: %w", def.Prefix, err)
		}

		key := targetURL.String()
		up, ok := p.upstreams[key]
		if !ok {
			healthPath := def.HealthPath
			if healthPath == "" {
				healthPath = "/healthz"
			}
			up = &upstreamTarget{
				name:   targetURL.Host,
				url:    targetURL,
				health: healthPath,
				status: upstreamStatus{
					Healthy:         true,
					WindowStartedAt: time.Now(),
				},
			}
			p.upstreams[key] = up
		}

		route := &route{
			name:         def.Name,
			prefix:       normalizePrefix(def.Prefix),
			stripPrefix:  def.StripPrefix,
			rewrite:      def.Rewrite,
			requireAuth:  def.RequireAuth,
			authStrategy: normalizeAuthStrategy(def.AuthStrategy),
			target:       up,
		}

		if len(def.Methods) > 0 {
			route.methods = map[string]struct{}{}
			for _, m := range def.Methods {
				route.methods[strings.ToUpper(m)] = struct{}{}
			}
		}

		if route.name == "" {
			route.name = route.prefix
		}

		route.proxy = p.newReverseProxy(up.url)
		route.proxy.ErrorHandler = p.makeErrorHandler(route)

		p.routes = append(p.routes, route)
	}

	sort.SliceStable(p.routes, func(i, j int) bool {
		return len(p.routes[i].prefix) > len(p.routes[j].prefix)
	})
	return nil
}

func (p *proxyEngine) newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = p.transport
	return proxy
}

func (p *proxyEngine) makeErrorHandler(route *route) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		writeProxyError(w, pkgErrors.New("bad_gateway", err.Error()))
		p.metrics.Observe(route.name, http.StatusBadGateway, 0)
		if route.target.recordResult(false, time.Now(), p.circuitWindow, p.circuitThreshold, p.circuitCooldown) {
			go p.checkUpstream(route.target)
		}
		p.log.Warn("proxy upstream error",
			zap.String("route", route.prefix),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
	}
}

func normalizeAuthStrategy(strategy string) string {
	switch strings.ToLower(strategy) {
	case "validate":
		return "validate"
	default:
		return "forward"
	}
}

func normalizePrefix(prefix string) string {
	if prefix == "" {
		return "/"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if len(prefix) > 1 {
		prefix = strings.TrimRight(prefix, "/")
	}
	return prefix
}

func loadRouteDefinitions(path string) ([]routeDefinition, error) {
	if path == "" {
		return nil, errors.New("routes file path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file routeFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	defs := append([]routeDefinition{}, file.Routes...)
	if !nullFallback(file.Fallback) {
		base := strings.TrimRight(file.Fallback.BasePath, "/")
		if base == "" {
			base = "/api"
		}
		for key, mapping := range file.Fallback.Mapping {
			prefix := fmt.Sprintf("%s/%s", base, key)
			strip := mapping.StripPrefix
			if file.Fallback.StripBase {
				strip = true
			}
			defs = append(defs, routeDefinition{
				Name:         fmt.Sprintf("fallback-%s", key),
				Prefix:       prefix,
				Upstream:     mapping.Upstream,
				StripPrefix:  strip,
				RequireAuth:  mapping.RequireAuth,
				AuthStrategy: mapping.AuthStrategy,
				HealthPath:   firstNonEmpty(mapping.HealthPath, file.Fallback.HealthPath, "/healthz"),
			})
		}
	}
	return defs, nil
}

func nullFallback(fb *fallbackDefinition) bool {
	return fb == nil || fb.BasePath == "" || len(fb.Mapping) == 0
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func (p *proxyEngine) Start(ctx context.Context) {
	if len(p.routes) == 0 {
		p.readyOnce.Do(func() {
			close(p.readyCh)
		})
		return
	}

	go func() {
		p.runHealthChecks()
		p.readyOnce.Do(func() { close(p.readyCh) })

		if p.healthInterval <= 0 {
			p.healthInterval = 10 * time.Second
		}
		ticker := time.NewTicker(p.healthInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.runHealthChecks()
			}
		}
	}()
}

func (p *proxyEngine) WaitUntilReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.readyCh:
		return nil
	}
}

func (p *proxyEngine) runHealthChecks() {
	for _, upstream := range p.upstreams {
		p.checkUpstream(upstream)
	}
}

func (p *proxyEngine) checkUpstream(target *upstreamTarget) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target.healthURL(), nil)
	if err != nil {
		return
	}

	resp, err := p.healthClient.Do(req)
	if err != nil {
		target.markUnhealthy(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		target.markUnhealthy(fmt.Errorf("health check status %d", resp.StatusCode))
		return
	}
	target.markHealthy()
}

func (u *upstreamTarget) healthURL() string {
	healthPath := u.health
	if !strings.HasPrefix(healthPath, "/") {
		healthPath = "/" + healthPath
	}
	return fmt.Sprintf("%s://%s%s", u.url.Scheme, u.url.Host, healthPath)
}

func (u *upstreamTarget) markHealthy() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.status.Healthy = true
	u.status.LastChecked = time.Now()
	u.status.LastError = ""
	u.status.UnhealthySince = time.Time{}
	u.status.ConsecutiveErrors = 0
	if time.Now().After(u.status.CircuitOpenUntil) {
		u.status.CircuitOpenUntil = time.Time{}
	}
}

func (u *upstreamTarget) markUnhealthy(err error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.status.Healthy = false
	u.status.LastChecked = time.Now()
	u.status.LastError = err.Error()
	if u.status.UnhealthySince.IsZero() {
		u.status.UnhealthySince = time.Now()
	}
	u.status.ConsecutiveErrors++
}

func (u *upstreamTarget) isAvailable(now time.Time) (bool, string) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if !u.status.CircuitOpenUntil.IsZero() && now.Before(u.status.CircuitOpenUntil) {
		return false, "circuit_open"
	}
	if !u.status.Healthy {
		return false, u.status.LastError
	}
	return true, ""
}

func (u *upstreamTarget) snapshot() upstreamStatus {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.status
}

func (u *upstreamTarget) recordResult(success bool, now time.Time, window time.Duration, threshold float64, cooldown time.Duration) bool {
	if window <= 0 {
		window = 30 * time.Second
	}
	if threshold <= 0 {
		threshold = 0.5
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if now.Sub(u.status.WindowStartedAt) > window {
		u.status.WindowStartedAt = now
		u.status.RequestsInWindow = 0
		u.status.FailuresInWindow = 0
	}

	u.status.RequestsInWindow++
	if !success {
		u.status.FailuresInWindow++
		u.status.ConsecutiveErrors++
	} else {
		u.status.ConsecutiveErrors = 0
	}

	if !success && u.status.RequestsInWindow >= 3 {
		ratio := float64(u.status.FailuresInWindow) / float64(u.status.RequestsInWindow)
		if ratio >= threshold {
			u.status.CircuitOpenUntil = now.Add(cooldown)
			u.status.LastError = "circuit opened due to error ratio"
			return true
		}
	}
	return false
}

func (p *proxyEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.mode != ModeProxyAll {
		writeProxyError(w, pkgErrors.New("not_found", "proxy mode disabled (whitelist)"))
		return
	}

	route := p.matchRoute(r.URL.Path)
	if route == nil {
		writeProxyError(w, pkgErrors.New("not_found", "no upstream mapping"))
		return
	}

	if len(route.methods) > 0 {
		if _, ok := route.methods[r.Method]; !ok {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	now := time.Now()
	if ok, reason := route.target.isAvailable(now); !ok {
		writeProxyError(w, pkgErrors.New("service_unavailable", fmt.Sprintf("upstream unavailable: %s", reason)))
		p.metrics.Observe(route.name, http.StatusServiceUnavailable, 0)
		return
	}

	if err := p.ensureAuth(r, route); err.Code != "" {
		writeProxyError(w, err)
		p.metrics.Observe(route.name, pkgErrors.StatusCode(err), 0)
		return
	}

	p.forward(route, w, r)
}

func (p *proxyEngine) ensureAuth(r *http.Request, route *route) pkgErrors.APIError {
	if !route.requireAuth {
		return pkgErrors.APIError{}
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return pkgErrors.New("unauthorized", "missing Authorization header")
	}

	if route.authStrategy != "validate" {
		return pkgErrors.APIError{}
	}
	tokenString := extractBearer(authHeader)
	if tokenString == "" {
		return pkgErrors.New("unauthorized", "missing bearer token")
	}

	if p.jwtSecret == "" {
		return pkgErrors.New("unauthorized", "jwt secret not configured")
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(p.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return pkgErrors.New("unauthorized", "invalid token")
	}
	return pkgErrors.APIError{}
}

func extractBearer(header string) string {
	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}

func (p *proxyEngine) matchRoute(path string) *route {
	for _, route := range p.routes {
		if route.matches(path) {
			return route
		}
	}
	return nil
}

func (r *route) matches(path string) bool {
	if r.prefix == "/" {
		return true
	}
	if strings.HasPrefix(path, r.prefix) {
		if len(path) == len(r.prefix) {
			return true
		}
		if r.prefix == "/" {
			return true
		}
		if strings.HasPrefix(path, r.prefix+"/") {
			return true
		}
	}
	return false
}

func (p *proxyEngine) forward(route *route, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), p.timeout)
	defer cancel()

	req := r.Clone(ctx)
	req.URL.Path = route.rewritePath(r.URL.Path)
	req.URL.RawPath = req.URL.Path

	rec := &proxyResponseWriter{ResponseWriter: w, status: http.StatusOK}

	route.proxy.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	p.metrics.Observe(route.name, rec.status, elapsed)
	if route.target.recordResult(rec.status < http.StatusInternalServerError, time.Now(), p.circuitWindow, p.circuitThreshold, p.circuitCooldown) {
		go p.checkUpstream(route.target)
	}

	p.log.Info("proxy request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_ip", remoteIP(r)),
		zap.String("user_agent", r.UserAgent()),
		zap.String("route", route.prefix),
		zap.String("upstream", route.target.url.Host),
		zap.Int("status", rec.status),
		zap.Float64("latency_ms", float64(elapsed.Milliseconds())),
	)
}

func (r *route) rewritePath(path string) string {
	if r.prefix == "/" {
		return path
	}

	if r.rewrite != "" {
		suffix := strings.TrimPrefix(path, r.prefix)
		return cleanPath(r.rewrite + suffix)
	}

	if r.stripPrefix && strings.HasPrefix(path, r.prefix) {
		newPath := strings.TrimPrefix(path, r.prefix)
		if newPath == "" {
			return "/"
		}
		if !strings.HasPrefix(newPath, "/") {
			newPath = "/" + newPath
		}
		return newPath
	}
	return path
}

func cleanPath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

func remoteIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeProxyError(w http.ResponseWriter, err pkgErrors.APIError) {
	status := pkgErrors.StatusCode(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(err)
}

func (p *proxyEngine) Metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	for _, line := range p.metrics.Format() {
		_, _ = w.Write([]byte(line + "\n"))
	}
}

func (p *proxyEngine) DebugRoutes(w http.ResponseWriter, _ *http.Request) {
	type debugRoute struct {
		Name        string         `json:"name"`
		Prefix      string         `json:"prefix"`
		Upstream    string         `json:"upstream"`
		RequireAuth bool           `json:"require_auth"`
		Auth        string         `json:"auth_strategy"`
		Status      upstreamStatus `json:"status"`
	}

	var payload []debugRoute
	for _, route := range p.routes {
		payload = append(payload, debugRoute{
			Name:        route.name,
			Prefix:      route.prefix,
			Upstream:    route.target.url.String(),
			RequireAuth: route.requireAuth,
			Auth:        route.authStrategy,
			Status:      route.target.snapshot(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func (p *proxyEngine) Healthz(w http.ResponseWriter, _ *http.Request) {
	type health struct {
		Upstream string         `json:"upstream"`
		Status   upstreamStatus `json:"status"`
	}
	payload := make([]health, 0, len(p.upstreams))
	healthy := true
	for _, up := range p.upstreams {
		state := up.snapshot()
		if !state.Healthy {
			healthy = false
		}
		payload = append(payload, health{
			Upstream: up.url.String(),
			Status:   state,
		})
	}
	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type proxyResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *proxyResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

type retryTransport struct {
	base    http.RoundTripper
	retries int
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	attempts := 1
	if strings.EqualFold(req.Method, http.MethodGet) && t.retries > 0 {
		attempts += t.retries
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		cloned := cloneRequest(req)
		resp, err := t.base.RoundTrip(cloned)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		time.Sleep(backoffDelay(i))
	}
	return nil, lastErr
}

func cloneRequest(r *http.Request) *http.Request {
	return r.Clone(r.Context())
}

func backoffDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 50 * time.Millisecond
	}
	delay := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
	if delay > time.Second {
		return time.Second
	}
	return delay
}
