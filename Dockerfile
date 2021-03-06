FROM golang:1.18rc1-alpine3.15 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /usr/src/app

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/boomer ./boomer_fasthttp.go


FROM busybox

WORKDIR /app
COPY data.csv .
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /app/boomer /app/boomer

ENTRYPOINT ["/app/boomer"]