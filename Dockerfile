FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o webhook cmd/main.go

FROM scratch

COPY --from=builder /app/webhook /webhook

EXPOSE 8443
ENTRYPOINT ["/webhook"]