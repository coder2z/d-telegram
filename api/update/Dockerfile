FROM golang:1.19-alpine AS builder
RUN apk update && mkdir /app
WORKDIR /app
ENV GOPROXY https://goproxy.cn
ENV GO111MODULE on
COPY . /app
RUN go mod download
RUN CGO_ENABLED=0 go build -o tgUpdateApp ./api/update

FROM alpine:3.10

RUN apk update \
  && apk upgrade \
  && apk add --no-cache \
  ca-certificates tzdata bash \
  && update-ca-certificates 2>/dev/null || true \
  && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
  && mkdir /app
WORKDIR /app
COPY --from=builder /app/tgUpdateApp /app/tgUpdateApp

CMD ["/tgUpdateApp"]