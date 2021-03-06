FROM golang

ADD . /wiki

WORKDIR /wiki

ENV WIKIDIR /usr/share/wiki

RUN mkdir secret
ENV KEYLOCATION secret

ENV LOGFILE ""
ENV PORT 8990

ENV CGO_ENABLED 1

RUN go get ./...
RUN go build ./... 

EXPOSE 8990
ENTRYPOINT /wiki/wiki
