FROM golang:alpine as builder

# RUN apk update && apk upgrade --no-cache && apk add --no-cache ca-certificates
RUN apk update && apk upgrade --no-cache

WORKDIR /src
COPY . .
RUN go build -trimpath -gcflags=all="-B" -ldflags="-s -w -buildid=" -o ovai ./cmd/ovai/main.go

FROM prantlf/healthchk as healthchk

# FROM gcr.io/distroless/static-debian12
FROM busybox:stable
LABEL maintainer="Ferdinand Prantl <prantlf@gmail.com>"

# RUN apt-get update -y && apt-get upgrade -y && \
#   apt-get install -y --no-install-recommends ca-certificates && \
#   apt-get clean && apt-get autoremove && \
#   rm -rf /var/cache/apt/archives/* && rm -rf /var/lib/apt/lists/*

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/ovai /
COPY --from=healthchk /healthchk /

WORKDIR /
EXPOSE 22434
ENTRYPOINT ["/ovai"]

ARG DEBUG=ovai,ovai:srv
ENV DEBUG=${DEBUG}
ENV PORT=22434
ENV NETWORK=
ENV OLLAMA_ORIGIN=

# HEALTHCHECK --interval=5m \
#   CMD /healthchk http://localhost:22434/api/ping || exit 1
