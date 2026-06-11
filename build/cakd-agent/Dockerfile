FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/cakd-agent/ cmd/cakd-agent/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux go build -o cakd-agent ./cmd/cakd-agent

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/cakd-agent .

RUN mkdir -p /etc/cakd

EXPOSE 8080
ENTRYPOINT ["./cakd-agent"]
