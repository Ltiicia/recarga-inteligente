FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/servidor
RUN go build -o /servidor main.go

WORKDIR /app/cmd/veiculo
RUN go build -o /veiculo main.go

WORKDIR /app/cmd/ponto-de-recarga
RUN go build -o /ponto-de-recarga main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /servidor /app/servidor
COPY --from=builder /veiculo /app/veiculo
COPY --from=builder /ponto-de-recarga /app/ponto-de-recarga
COPY --from=builder /app/internal/dataJson /app/internal/dataJson

EXPOSE 5000