FROM golang:1.25.0-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache ca-certificates

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY server.go ./

# Build the application
RUN go build -o oidc-mock server.go

# Expose port
EXPOSE 9999

# Run the server
CMD ["/app/oidc-mock"]
