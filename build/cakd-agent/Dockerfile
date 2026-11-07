FROM golang:1.25-alpine AS builder
ENV GOTOOLCHAIN=auto

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/cakd-agent/ cmd/cakd-agent/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux go build -o cakd-agent ./cmd/cakd-agent

FROM alpine:3.21
RUN apk --no-cache add ca-certificates wget && adduser -D -u 65532 cakd

WORKDIR /home/cakd/
COPY --from=builder /app/cakd-agent .

RUN mkdir -p /etc/cakd && chown -R cakd:cakd /home/cakd /etc/cakd

USER cakd
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:8080/healthz || exit 1
ENTRYPOINT ["./cakd-agent"]
