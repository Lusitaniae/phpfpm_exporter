ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="https://github.com/Lusitaniae"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/phpfpm_exporter /bin/phpfpm_exporter

EXPOSE      9117
USER        nobody
ENTRYPOINT  [ "/bin/phpfpm_exporter" ]