FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY cmd/ponto-de-recarga/ ./ 
COPY . .
RUN go build -o /ponto-de-recarga

FROM alpine:latest

WORKDIR /

COPY --from=builder /ponto-de-recarga /ponto-de-recarga

ENV SERVER_ADDRESS=servidor
ENV SERVER_PORT=5000

CMD ["./ponto-de-recarga"]
