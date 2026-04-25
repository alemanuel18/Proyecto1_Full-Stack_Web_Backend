# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o seriestracker ./main.go

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
# Use a minimal image — no Go toolchain needed at runtime
FROM alpine:3.19

WORKDIR /app

# Add CA certs (needed for HTTPS calls to Cloudinary)
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/seriestracker .

EXPOSE 8080

CMD ["./seriestracker"]