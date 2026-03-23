FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o blog-api ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/blog-api .

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./blog-api"]
