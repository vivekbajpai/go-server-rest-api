# Build stage
FROM golang:1.24-bullseye AS build
WORKDIR /src

# Copy go.mod & go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /app/server ./

# Runtime stage
FROM gcr.io/distroless/static:nonroot

# Create app directory
WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/server /app/server

# Expose port
EXPOSE 8080

# Use non-root user provided by distroless
USER nonroot:nonroot

# Entrypoint
ENTRYPOINT ["/app/server"]
