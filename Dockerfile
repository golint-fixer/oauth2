FROM golang
MAINTAINER Douézan-Grard Guillaume - Quorums

RUN go get github.com/quorumsco/oauth2

ADD . /go/src/github.com/quorumsco/oauth2

WORKDIR /go/src/github.com/quorumsco/oauth2

RUN \
  go get -u && \
  go build

EXPOSE 8080

ENTRYPOINT ["./oauth2"]
