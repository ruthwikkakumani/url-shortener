FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire workspace to have access to pkg/ and services/
COPY . .

# Build the auth service
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./services/auth-service/cmd/main.go

# Use a minimal alpine image for the final container
FROM alpine:latest

WORKDIR /root/

# Install CA certificates just in case your auth service makes external HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/auth-service .

# Inform Docker that the container listens on the specified port
EXPOSE 8080

# Command to run the executable
CMD ["./auth-service"]
