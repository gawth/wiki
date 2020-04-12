FROM golang

ADD . /go/src/github.com/gawth/wiki

WORKDIR /go/src/github.com/gawth/wiki

RUN mkdir wikidir
ENV WIKIDIR wikidir
ENV LOGFILE ""

RUN go get ./...
RUN go build ./... 

ENTRYPOINT /go/src/github.com/gawth/wiki/wiki

EXPOSE 8080
