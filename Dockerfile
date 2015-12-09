FROM golang:1.5.2
MAINTAINER Craig Peterson <peterson.craig@gmail.com>

ADD . /go/src/github.com/captncraig/caddycustom
ENV GO15VENDOREXPERIMENT=1
RUN go install github.com/captncraig/caddycustom

EXPOSE 80
EXPOSE 443

RUN echo "http://localhost:80" > /etc/Caddyfile

RUN mkdir /caddy
WORKDIR /caddy

ENTRYPOINT ["/go/bin/caddycustom"]
CMD ["--conf", "/etc/Caddyfile", "--agree"]