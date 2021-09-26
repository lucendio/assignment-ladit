FROM docker.io/golang:1.17.1 as test-stage

WORKDIR /go/src
COPY ./src ./
RUN go get -d
RUN go test -race ./...


FROM docker.io/golang:1.17.1 as build-stage

WORKDIR /go/src
COPY ./src ./
RUN go get -d
RUN GOOS=linux GOARCH=amd64 go build -v -o /go/artifact.bin ./*.go


FROM docker.io/fedora:34

ENV CNTROOT=/opt/ctnroot

RUN dnf update -y \
    && dnf clean -y all

ENV \
    HOME=/opt/ctnroot/app \
    PATH=/opt/ctnroot/app/bin:/opt/ctnroot/bin:${PATH}

WORKDIR ${HOME}
COPY --from=build-stage /go/artifact.bin /opt/ctnroot/app/bin/blocksvc
RUN chmod u+x /opt/ctnroot/app/bin/blocksvc

RUN groupadd --gid 2002 ctnrgroup \
    && useradd --uid 1001 --system --gid 2002 --home-dir ${HOME} \
            --shell /sbin/nologin --comment "ctnr user" \
            ctnruser \
    && chown -R 1001:2002 ${CNTROOT}

USER 1001

ENTRYPOINT [ "blocksvc" ]
CMD [ ]

EXPOSE 3000
