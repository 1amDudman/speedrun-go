# Build stage
FROM golang:1.25.7-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o parser-app ./main.go

# Final stage
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/parser-app .

EXPOSE 8080

CMD ["./parser-app"]