FROM golang:1.23.5-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /safestore

FROM postgres:alpine

ENV POSTGRES_DB=safestore
ENV POSTGRES_USER=safeuser
ENV POSTGRES_PASSWORD=safepassword

# Copy the built Go application
COPY --from=builder /safestore /safestore

# Install necessary packages to run Go application
RUN apk add --no-cache ca-certificates

# Copy the entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 5432 3478

ENTRYPOINT ["/entrypoint.sh"]
