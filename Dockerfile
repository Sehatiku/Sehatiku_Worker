FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
ENV TZ=Asia/Jakarta
WORKDIR /app
COPY --from=builder /app/worker .
CMD ["./worker"]
