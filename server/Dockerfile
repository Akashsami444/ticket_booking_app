# Step 1: Build stage
# Using the official Go image as the base
FROM golang:1.24.6-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code (including the proto folder)
# Note: Since your proto is in a sibling directory, we usually 
# copy the whole project root or adjust paths. 
# For now, let's assume we copy from the root context.
COPY . .

# Build the Go app
RUN go build -o /grpc-server ./server/main.go

# Step 2: Final lightweight image
FROM alpine:latest
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /grpc-server .

# Expose the port the server runs on
EXPOSE 50051

# Command to run the executable
CMD ["./grpc-server"]