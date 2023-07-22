# Stage 1: Build Go binary
FROM golang:alpine as builder

WORKDIR /app

# Download necessary Go dependencies
COPY ../go.mod .
COPY ../go.sum .
RUN go mod download

# Copy all local source code to the container workspace
COPY .. .

# Build the Go binary
RUN GOARCH=arm64 go build -o weather ./cmd/weather/main.go

# Stage 2: Runtime image
FROM alpine:latest

# Copy creds.json from local to docker image
COPY ./creds.json /app/

# Copy Go binary from builder stage
COPY --from=builder /app/weather /app/

# Change working directory
WORKDIR /app

# Run the binary with the provided argument
CMD ["./weather", "flag"]
