FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auctionxi-api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /auctionxi-api .
EXPOSE 8080
CMD ["./auctionxi-api"]
