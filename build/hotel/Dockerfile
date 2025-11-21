FROM golang:1.23 AS builder
WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/hotel ./cmd/hotel-service

FROM alpine:3.19
RUN adduser -D -g '' appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /bin/hotel ./hotel
EXPOSE 8081
ENTRYPOINT ["./hotel"]
