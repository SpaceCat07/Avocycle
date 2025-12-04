# Build stage
FROM golang:1.25.1-alpine AS builder

# Install C compiler (CGO dependencies)
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build main application (CGO enabled)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .
# COPY --from=builder /app/seeder .

EXPOSE 2005

CMD ["./main"]
