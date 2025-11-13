FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0
WORKDIR /api

COPY . .
RUN go mod download

RUN go build -o api ./cmd/api

FROM scratch
WORKDIR /api

COPY --from=builder /api/api /api/api

COPY .env .env


CMD ["/api/api"]
