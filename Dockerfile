FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/api

FROM alpine:3.23
RUN adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /server /app/server
COPY docs /app/docs
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/server"]
