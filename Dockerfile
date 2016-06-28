FROM golang:1.6.0
MAINTAINER Greg Aker <me@gregaker.net>

ADD . /go/src/github.com/gaker/snmp_send
WORKDIR /go/src/github.com/gaker/snmp_send

RUN go get github.com/tools/godep \ 
    && godep restore 

CMD go run main.go
