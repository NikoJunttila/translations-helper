# Build-Stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
WORKDIR /app

# Force Go module mode
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org,direct

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Generate Go code from templates
RUN go tool templ generate

# Build the application
# Use TARGETOS and TARGETARCH for cross-compilation
ARG TARGETOS TARGETARCH
RUN GO111MODULE=on GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
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

# Set version from build arg
ARG VERSION
ENV VERSION=${VERSION}

# Command to run the application
CMD ["./main"]
