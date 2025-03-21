FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
WORKDIR /app/cmd/veiculo
RUN go build -o /veiculo main.go

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /veiculo /app/veiculo

CMD ["/app/servidor" ]
