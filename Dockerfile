FROM golang:1.24.2 AS builder

WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the project
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux go build -v -a -o bin/my-device-plugin cmd/main.go

FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/bin/my-device-plugin .

ENTRYPOINT ["./my-device-plugin"]