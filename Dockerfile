FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o site-analyzer cmd/main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/site-analyzer .

COPY --from=builder /app/internal/web/templates /app/internal/web/templates

RUN adduser -D -u 1001 appuser
USER appuser

EXPOSE 8080

CMD ["./site-analyzer"]