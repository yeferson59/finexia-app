# ---- Build stage ----
FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache ca-certificates upx

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/api \
    ./cmd/api

RUN upx --best --lzma /app/api

# ---- Final stage ----
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

COPY --from=builder /app/api .

USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/app/api"]
