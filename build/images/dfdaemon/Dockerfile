FROM golang:1.15.10-alpine as builder

WORKDIR /go/src/d7y.io/dragonfly/v2

RUN apk --no-cache add bash make gcc libc-dev git

COPY . /go/src/d7y.io/dragonfly/v2

ARG GOPROXY

RUN make build-dfget && make install-dfget

FROM alpine:3.12

ENV PATH=/opt/dragonfly/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

COPY --from=builder /opt/dragonfly/bin/dfget /opt/dragonfly/bin/dfget

EXPOSE 65001

ENTRYPOINT ["/opt/dragonfly/bin/dfget", "daemon"]

