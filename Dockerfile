# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy everything first so go mod tidy can resolve all transitive dependencies
COPY . .

# Tidy generates go.sum, downloads deps, then builds a static binary
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o seriestracker ./main.go

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
# Minimal final image — no Go toolchain, ~15MB total
FROM alpine:3.19

WORKDIR /app

# CA certs needed for HTTPS calls to Cloudinary
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/seriestracker .

EXPOSE 8080

CMD ["./seriestracker"]