FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /pgraph-server ./cmd/server

FROM scratch

COPY --from=builder /pgraph-server /pgraph-server

EXPOSE 8080

ENTRYPOINT ["/pgraph-server"]
