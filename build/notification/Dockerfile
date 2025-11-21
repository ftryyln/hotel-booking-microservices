FROM golang:1.23 AS builder
WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/notification ./cmd/notification-service

FROM alpine:3.19
RUN adduser -D -g '' appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /bin/notification ./notification
EXPOSE 8085
ENTRYPOINT ["./notification"]
