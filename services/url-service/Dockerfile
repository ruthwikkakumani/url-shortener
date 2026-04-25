FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire workspace to have access to pkg/ and services/
COPY . .

# Build the url service
RUN CGO_ENABLED=0 GOOS=linux go build -o url-service ./services/url-service/cmd/main.go

# Use a minimal alpine image for the final container
FROM alpine:latest

WORKDIR /root/

# Install CA certificates
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/url-service .

# Inform Docker that the container listens on port 8082
EXPOSE 8082

# Command to run the executable
CMD ["./url-service"]
