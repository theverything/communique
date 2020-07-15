FROM golang:1.14-alpine AS builder

ENV GOOS linux
ENV GOARCH amd64

WORKDIR /build

RUN apk add --upgrade --no-cache git

COPY . ./

RUN go build -o ./communique ./

###################################################

FROM alpine

COPY --from=builder /build/communique /communique

EXPOSE 3000

ENTRYPOINT ["/communique"]
