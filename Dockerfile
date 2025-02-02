# Use a Go base image
FROM golang:1.20-alpine

# Set the current working directory in the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o websocket-server .

# Expose the WebSocket server port
EXPOSE 8080

# Command to run the Go application
CMD ["./websocket-server"]
