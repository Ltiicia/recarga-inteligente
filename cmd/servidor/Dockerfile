FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY cmd/servidor/ ./ 
COPY . .
RUN go build -o /servidor 

FROM alpine:latest

WORKDIR /

COPY --from=builder /servidor /servidor

EXPOSE 5000

CMD ["./servidor"]
