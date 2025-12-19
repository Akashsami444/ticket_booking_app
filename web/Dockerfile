# Build stage
FROM golang:1.24.6-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the web app
RUN go build -o /web-app ./web/main.go

# Final stage
FROM alpine:latest
WORKDIR /root/

# Copy the binary and the HTML file
COPY --from=builder /web-app .
# IMPORTANT: Copy the index.html so the binary can find it
COPY --from=builder /app/web/index.html .

EXPOSE 8888

CMD ["./web-app"]