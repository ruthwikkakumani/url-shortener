FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire workspace to have access to pkg/ and services/
COPY . .

# Build the url service
RUN CGO_ENABLED=0 GOOS=linux go build -o redirect-service ./services/redirect-service/cmd/main.go

# Use a minimal alpine image for the final container
FROM alpine:latest

WORKDIR /root/

# Install CA certificates
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/redirect-service .

# Inform Docker that the container listens on port 8083
EXPOSE 8083

# Command to run the executable
CMD ["./redirect-service"]
