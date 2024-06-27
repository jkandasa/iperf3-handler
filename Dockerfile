# --- BUILDER ---
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.21-alpine3.20 AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app

# "tzdata" to print the build date with timezone
RUN apk add --no-cache tzdata git

ARG GOPROXY
# download deps before gobuild
RUN go mod download -x
ARG TARGETOS
ARG TARGETARCH
RUN scripts/build_container_binary.sh

# --- RUNNER ---
FROM alpine:3.20
LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ENV APP_HOME="/app"

EXPOSE 8080

# install timzone utils and iperf3
RUN apk --no-cache add tzdata iperf3

# create a user and give permission for the locations
RUN mkdir -p ${APP_HOME}

# copy application bin file
COPY --from=builder /app/iperf3-handler ${APP_HOME}/iperf3-handler

RUN chmod +x ${APP_HOME}/iperf3-handler

WORKDIR ${APP_HOME}

ENTRYPOINT [ "/app/iperf3-handler" ]
