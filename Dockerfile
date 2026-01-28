# Build-Stage
FROM golang:1.24-alpine AS build
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Force Go module mode
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org,direct

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# We use GOWORK=off and GO111MODULE=on to ensure strict module resolution
# Added debug checkpoints to diagnose resolution issues if they persist
RUN go env && ls -R /app && \
    GO111MODULE=on GOWORK=off CGO_ENABLED=1 GOOS=linux \
    go build -mod=readonly -o main .

# Deploy-Stage
FROM docker.io/alpine:latest

WORKDIR /app

# Install ca-certificates
RUN apk add --no-cache ca-certificates

# Set environment variable for runtime
ENV GO_ENV=production

# Copy the binary from the build stage
COPY --from=build /app/main .

# Expose the port your application runs on
EXPOSE 8090

# Command to run the application
CMD ["./main"]
