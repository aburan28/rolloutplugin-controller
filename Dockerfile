FROM library/golang:1.24.1 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o rolloutplugin-controller cmd/root


# Start a new stage from scratch
FROM library/debian:stretch-slim

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/rolloutplugin-controller /app/rolloutplugin-controller

ENTRYPOINT ["/app/rolloutplugin-controller"]