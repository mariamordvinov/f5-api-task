# Use the official Go image
FROM golang:1.21.10

# Set the working directory
WORKDIR /app

# Copy Go modules manifests and source code
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the application
RUN go build -o main .

# Expose the port
EXPOSE 8080

# Run the application
CMD ["./main"]
