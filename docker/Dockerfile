FROM golang:1.24 AS builder

WORKDIR /app

COPY ./go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o docker-mgr ./cmd/docker-mgr

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/docker-mgr .

COPY --from=builder /app/config/config.yaml ./config/config.yaml

EXPOSE 8082

ENTRYPOINT ["./docker-mgr"]