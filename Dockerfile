# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build argument to select which binary to build
ARG CMD=api
RUN go build -o main ./cmd/${CMD}

# Stage 2: Runtime
FROM alpine:3.20

WORKDIR /app

# Install ca-certificates for HTTPS (if needed)
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
