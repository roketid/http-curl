
# Start from golang base image
FROM golang:alpine as builder

# ENV GO111MODULE=on

# Add Maintainer info
LABEL maintainer="Kata.ai"

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git tzdata build-base brotli-dev

# Set environment variables
ENV APPDIR /app

# Set the current working directory inside the container
RUN mkdir -p ${APPDIR}
WORKDIR ${APPDIR}

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./main.go

ARG ENV
ENV ENV ${ENV:-production}

# Start a new stage from scratch
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata bash vim brotli-libs curl

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /app/main .

# Expose port to the outside world
EXPOSE 8080

#Command to run the executable
CMD ["./main"]
