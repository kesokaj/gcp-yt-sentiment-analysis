FROM golang:1.25.0-alpine AS builder

# Set necessary environment variables for the build
ENV CGO_ENABLED=0
ENV GOOS=linux

# Set the working directory inside the container
WORKDIR /build

# Copy the application source code
COPY main.go .
COPY pkgs/ ./pkgs/
COPY web/ ./web/

# Download packages
RUN go mod init app && go mod tidy

# Build the application
# -ldflags="-s -w": reduces the binary size by omitting debug info and symbol table
RUN go build -o server -ldflags="-s -w" .

# --- Final Stage ---
FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/server .
COPY --from=builder /build/web ./web

ENTRYPOINT ["./server"]
