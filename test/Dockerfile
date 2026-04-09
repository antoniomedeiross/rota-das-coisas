FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o teste_concorrencia .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/teste_concorrencia .

CMD ["./teste_concorrencia"]