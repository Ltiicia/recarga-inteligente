FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY cmd/veiculo/ ./ 
COPY . .
RUN go build -o /veiculo

FROM alpine:latest

WORKDIR /

COPY --from=builder /veiculo /veiculo

ENV SERVER_ADDRESS=servidor
ENV SERVER_PORT=5000

CMD ["./veiculo"]
