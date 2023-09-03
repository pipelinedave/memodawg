# First build stage to build the Go application
FROM golang:1.16 AS builder

WORKDIR /app

# Initialize go module and download dependencies
RUN go mod init memodawg-api && \
    go mod tidy

COPY main.go .
RUN CGO_ENABLED=0 go build -o main

# Second build stage to create the runtime image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the first stage
COPY --from=builder /app/main /app/

CMD ["/app/main"]
