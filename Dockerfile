FROM quay.io/prometheus/busybox:latest

COPY phpfpm_exporter /bin/phpfpm_exporter

ENTRYPOINT ["/bin/phpfpm_exporter"]
EXPOSE     9253
