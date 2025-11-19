FROM golang:1.24.3 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/auth ./cmd/auth-service

FROM alpine:3.19
RUN adduser -D -g '' appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /bin/auth ./auth
EXPOSE 8080
ENTRYPOINT ["./auth"]



