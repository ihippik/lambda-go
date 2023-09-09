FROM golang:1.21-alpine AS builder

WORKDIR /tmp/server-app

COPY go.mod ./
COPY go.sum ./
COPY main.go ./
RUN go mod tidy
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o ./out/server-app .

FROM alpine:3.18.3

RUN apk add ca-certificates

COPY --from=builder /tmp/server-app/out/server-app /app/server-app

EXPOSE 8080

CMD ["/app/server-app"]