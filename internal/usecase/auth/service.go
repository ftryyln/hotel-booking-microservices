package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
	"github.com/ftryyln/hotel-booking-microservices/pkg/valueobject"
)

// Service coordinates registration/login use cases.
type Service struct {
	repo   domain.UserRepository
	issuer domain.TokenIssuer
}

func NewService(repo domain.UserRepository, issuer domain.TokenIssuer) *Service {
	return &Service{repo: repo, issuer: issuer}
}

var allowedRoles = map[string]struct{}{
	"customer": {},
	"admin":    {},
}

// Register creates new user and issues tokens.
func (s *Service) Register(ctx context.Context, req dto.RegisterRequest) (dto.AuthResponse, error) {
	email, err := valueobject.NormalizeEmail(req.Email)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	if req.Password == "" {
		return dto.AuthResponse{}, errors.New("bad_request", "password required")
	}
	role, err := valueobject.ParseRole(req.Role)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return dto.AuthResponse{}, errors.New("conflict", "email already used")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	user := domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hash),
		Role:      string(role),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return dto.AuthResponse{}, err
	}

	return s.issueTokens(ctx, user)
}

// Login verifies credentials.
func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (dto.AuthResponse, error) {
	email, err := valueobject.NormalizeEmail(req.Email)
	if err != nil {
		return dto.AuthResponse{}, errors.New("unauthorized", "invalid credentials")
	}
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return dto.AuthResponse{}, errors.New("unauthorized", "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return dto.AuthResponse{}, errors.New("unauthorized", "invalid credentials")
	}

	return s.issueTokens(ctx, user)
}

// Me returns profile info.
func (s *Service) Me(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, errors.New("not_found", "user not found")
	}
	return user, nil
}

// List returns all users (admin use).
func (s *Service) List(ctx context.Context, opts query.Options) ([]domain.User, error) {
	users, err := s.repo.List(ctx, opts.Normalize(50))
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Get fetches a user by id.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, errors.New("not_found", "user not found")
	}
	return user, nil
}

func (s *Service) issueTokens(ctx context.Context, user domain.User) (dto.AuthResponse, error) {
	access, refresh, err := s.issuer.Generate(ctx, user)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	return dto.AuthResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
