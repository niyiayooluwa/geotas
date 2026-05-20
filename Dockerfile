# Step 1: Build the Go binary
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

# Copy dependency manifests first for efficient caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire workspace structure
COPY . .

# Compile the binary from the specific cmd path
# Explicitly statically link it for maximum Alpine compatibility
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# Step 2: Minimal runtime image
FROM alpine:latest  

# Add ca-certificates in case your Go backend hits external HTTPS endpoints
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Optional: If your Go app serves openapi.yaml or executes raw migration SQL files at startup,
# uncomment the line below to ensure they are available in the container.
# COPY --from=builder /app/openapi.yaml ./openapi.yaml

# Hugging Face Spaces expects traffic on port 7860
EXPOSE 7860
ENV PORT=7860

CMD ["./main"]