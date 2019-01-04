# Build the Go Binary.

FROM golang:1.10.3 as build
ENV CGO_ENABLED 0
ARG packagename
ARG packageprefix
RUN mkdir -p /go/src/github.com/ardanlabs/service
COPY . /go/src/github.com/ardanlabs/service
WORKDIR /go/src/github.com/ardanlabs/service/cmd/${packageprefix}${packagename}
RUN go build -ldflags "-s -w -X main.build=$(git rev-parse HEAD)" -a -tags netgo


# Run the Go Binary in Alpine.

FROM alpine:3.7
ARG BUILD_DATE
ARG VCS_REF
ARG packagename
ARG packageprefix
EXPOSE 3001
EXPOSE 4001
COPY --from=build /go/src/github.com/ardanlabs/service/cmd/${packageprefix}${packagename}/${packagename} /bin/package
ENTRYPOINT /bin/package

LABEL org.opencontainers.image.created=$BUILD_DATE \
      org.opencontainers.image.title=${packagename} \
      org.opencontainers.image.authors="William Kennedy <bill@ardanlabs.com>" \
      org.opencontainers.image.source="https://github.com/ardanlabs/service/cmd/${packageprefix}${packagename}" \
      org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.vendor="Ardan Labs"