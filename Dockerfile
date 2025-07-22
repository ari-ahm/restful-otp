FROM golang:1.23.11-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/restful-otp ./cmd/api

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/restful-otp .

EXPOSE 8080

CMD ["./restful-otp"]