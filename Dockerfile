FROM golang:1.21 AS build-env

ADD . /lambda-go
WORKDIR /lambda-go

RUN go build -buildmode=pie -trimpath -ldflags='-s -w -buildid' -o app ./cmd/lambda-go

FROM cgr.dev/chainguard/busybox:latest-glibc

WORKDIR /run/app

COPY --from=build-env /lambda-go/app /app/

CMD /app/app