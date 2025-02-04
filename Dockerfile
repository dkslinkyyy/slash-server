# Start from Golang base image
FROM golang:1.23.4

WORKDIR /app

# Copy go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Set environment variables (optional)
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080
ENV SERVER_WEBSOCKET_PATH=/ws

# Build the application
RUN go build -o websocket-server ./cmd/server/main.go

# Run the application
CMD ["./websocket-server"]
