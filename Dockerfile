FROM golang
MAINTAINER Douézan-Grard Guillaume - Quorums

ADD . /go/src/github.com/quorumsco/oauth2

WORKDIR /go/src/github.com/quorumsco/oauth2

RUN \
  go get && \
  go build

EXPOSE 8080

ENTRYPOINT ["./oauth2"]
