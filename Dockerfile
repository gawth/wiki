FROM golang

ADD . /wiki

WORKDIR /wiki

RUN mkdir wikidir
RUN mkdir secret
ENV WIKIDIR wikidir
ENV LOGFILE ""
ENV KEYLOCATION secret
ENV PORT 8990

RUN go get ./...
RUN go build ./... 

EXPOSE 8990
ENTRYPOINT /wiki/wiki
