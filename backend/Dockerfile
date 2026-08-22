FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o social-network .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/social-network ./social-network
COPY --from=builder /app/pkg/db/migrations/sqlite ./pkg/db/migrations/sqlite
RUN mkdir -p uploads
EXPOSE 8080

CMD ["./social-network"]