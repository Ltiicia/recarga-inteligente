FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
WORKDIR /app/cmd/servidor
RUN go build -o /servidor main.go

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /servidor /app/servidor

EXPOSE 5000

CMD ["/app/servidor"]
