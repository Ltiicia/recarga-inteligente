FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
WORKDIR /app/cmd/ponto-de-recarga
RUN go build -o /ponto-de-recarga main.go

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /ponto-de-recarga /app/ponto-de-recarga

CMD ["/app/servidor"]
