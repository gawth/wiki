FROM golang

ADD . /wiki

WORKDIR /wiki

RUN mkdir wikidir
ENV WIKIDIR wikidir
ENV LOGFILE ""

RUN go get ./...
RUN go build ./... 

EXPOSE 8990
ENTRYPOINT /wiki/wiki
