FROM golang:1.24.2-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
# CGO_ENABLED=0 is important for a static build, which is necessary for a scratch/distroless image
# -o /app/lmt specifies the output file name
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lmt ./cmd/server/main.go

# ---- Runtime Stage ----
# Use a minimal image for the runtime environment
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/lmt .

# Copy the example configuration file
# The user will need to mount a real configuration file
COPY lmt.conf.example /app/lmt.conf.example

# Command to run the executable
# The user will likely need to pass arguments or a config file path
ENTRYPOINT ["/app/lmt"]
