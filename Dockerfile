FROM golang

RUN groupadd -r wikigrp -g 1024 && useradd -r -u 1024 -g wikigrp wikiusr

# Had to add this as Go needs it.  For some reason the above doesn't create it correctly
RUN mkdir /home/wikiusr
RUN chown wikiusr /home/wikiusr && chgrp wikigrp /home/wikiusr

ADD . /wiki
RUN chown wikiusr /wiki && chgrp wikigrp /wiki
RUN chmod 775 /wiki

WORKDIR /wiki

ENV WIKIDIR /usr/share/wiki

USER wikiusr

RUN mkdir secret
ENV KEYLOCATION secret


ENV LOGFILE ""
ENV PORT 8990

ENV CGO_ENABLED 1

RUN go get ./...
RUN go build ./... 

EXPOSE 8990
ENTRYPOINT /wiki/wiki
