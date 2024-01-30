# Use the official Golang image to provide all the necessary build tools
FROM --platform=$BUILDPLATFORM ncr.sky.nhn.no/dockerhub/golang:1.21.0-alpine3.18 as builder

# Set the current working directory inside the container to /app
WORKDIR /app

# Install GCC and musl-dev to be able to compile Go code
# --no-cache option is used to keep the image size small by not caching the index locally
RUN apk add --update gcc musl-dev ca-certificates --no-cache

# Copy the dependency information
COPY go.mod go.sum ./

# Download all the dependencies that are specified in go.mod and go.sum
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app with CGO enabled and link statically
# -a flag to rebuild all the packages
# -ldflags to set linker options
RUN GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o main .

#-----------------------------------------------------------------------#
# Stage 2: Prepare the final Docker image with the compiled application

# Start from a scratch (empty) image
FROM scratch
ENV CONFIG_PATH="/config"
ENV ROLES_PATH="/app/roles.yaml"
ENV GIN_MODE="release"

# Set the working directory to /app
WORKDIR /app

# Copy the compiled application from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/roles.yaml .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy static files for the server
#COPY ./index.html .

# Expose port 8080 for the application
EXPOSE 8080
USER 1000

# Set the entry point of the container to the application executable
ENTRYPOINT ["/app/main"]
