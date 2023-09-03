# Use the official Golang image to create a build artifact.
FROM golang:1.17 as builder

# Create and change to the app directory.
WORKDIR /app

# Initialize go module and download dependencies
RUN go mod init memodawg-frontend && \
    go mod tidy

COPY main.go .
COPY templates/ templates/
COPY static/ static/
RUN CGO_ENABLED=0 go build -o main

# Second build stage to create the runtime image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the first stage
COPY --from=builder /app/main /app/
COPY --from=builder /app/templates/ /app/templates/
COPY --from=builder /app/static/ /app/static/

CMD ["/app/main"]
