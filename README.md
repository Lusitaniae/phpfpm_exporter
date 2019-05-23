# PHP-FPM Exporter for Prometheus [![Build Status][buildstatus]][circleci]

[![GitHub release](https://img.shields.io/github/release/Lusitaniae/phpfpm_exporter.svg)][release]
![GitHub Downloads](https://img.shields.io/github/downloads/Lusitaniae/phpfpm_exporter/total.svg)
[![Docker Repository on Quay](https://quay.io/repository/Lusitaniae/phpfpm-exporter/status)][quay]
[![Docker Pulls](https://img.shields.io/docker/pulls/lusotycoon/phpfpm-exporter.svg?maxAge=604800)][hub]

Prometheus Exporter for the PHP-FPM status page.

Metrics are scrapped via unix socket and made available on port 9253.

This exporter also provides a way for embedding the output of arbitrary
PHP scripts into its metrics page, analogous to the node exporter's
`textfile` collector. Scripts that are specified with the
`--phpfpm.script-collector-paths` flag will be run through PHP-FPM. Any
metrics printed by the PHP script will be merged into the metrics
provided by this exported. An example use case includes printing metrics
for PHP's `opcache`.

# Usage:

Run the exporter
```
./phpfpm_exporter --phpfpm.socket-paths /var/run/phpfpm.sock
```

Include additional metrics from a PHP script

E.g. export OPcache metrics (using `contrib/php_opcache_exporter.php`)

Bear in mind these metrics are global, all FPM pools share the same cache.
```
./phpfpm_exporter --phpfpm.socket-paths /var/run/phpfpm.sock \
--phpfpm.script-collector-paths /usr/local/bin/php_exporter/phpfpm_opcache_exporter.php

```

Run with Docker
```
SOCK="/run/php/php7.2-fpm.sock"; \
docker run -d -p 9253:9253 -v $SOCK:$SOCK  \
lusotycoon/phpfpm-exporter \
--phpfpm.socket-paths=$SOCK
```

Help on flags

    ./phpfpm_exporter -h
    usage: phpfpm_exporter [<flags>]

    Flags:
      -h, --help     Show context-sensitive help (also try --help-long and --help-man).
          --web.listen-address=":9253"
                     Address to listen on for web interface and telemetry.
          --web.telemetry-path="/metrics"
                     Path under which to expose metrics.
          --phpfpm.socket-paths=PHPFPM.SOCKET-PATHS ...
                     Path(s) of the PHP-FPM sockets.
          --phpfpm.socket-directories=PHPFPM.SOCKET-DIRECTORIES ...
                     Path(s) of the directory where PHP-FPM sockets are located.
          --phpfpm.status-path="/status"
                     Path which has been configured in PHP-FPM to show status page.
          --version  Print version information.
          --phpfpm.script-collector-paths=PHPFPM.SCRIPT-COLLECTOR-PATHS ...
                     Paths of the PHP file whose output needs to be collected.

When using `--phpfpm.socket-directories`  make sure to use dedicated directories for PHP-FPM sockets to avoid timeouts.

# Metrics emitted by PHP-FPM:

```
php_fpm_accepted_connections_total{socket_path="/var/run/phpfpm.sock"} 300940
php_fpm_active_processes{socket_path="/var/run/phpfpm.sock"} 1
php_fpm_idle_processes{socket_path="/var/run/phpfpm.sock"} 5
php_fpm_listen_queue{socket_path="/var/run/phpfpm.sock"} 0
php_fpm_listen_queue_length{socket_path="/var/run/phpfpm.sock"} 0
php_fpm_max_active_processes{socket_path="/var/run/phpfpm.sock"} 10
php_fpm_max_children_reached{socket_path="/var/run/phpfpm.sock"} 3
php_fpm_max_listen_queue{socket_path="/var/run/phpfpm.sock"} 0
php_fpm_slow_requests{socket_path="/var/run/phpfpm.sock"} 0
php_fpm_start_time_seconds{socket_path="/var/run/phpfpm.sock"} 1.49277445e+09
php_fpm_total_processes{socket_path="/var/run/phpfpm.sock"} 3
php_fpm_up{socket_path="/var/run/phpfpm.sock"} 1
```
# Requirements

The FPM status page must be enabled in every pool you'd like to monitor by defining `pm.status_path = /status`.

# Grafana Dashboards
There's multiple grafana dashboards available for this exporter, find them at the urls below or in ```contrib/```.

[Basic:](https://grafana.com/dashboards/5579) for analyzing a single fpm pool in detail.

[Multi Pool:](https://grafana.com/dashboards/5714) for analyzing a cluster of fpm pools.

Basic:
![basic](https://grafana.com/api/dashboards/5579/images/3536/image)

Multi Pool:
![multi pool](https://grafana.com/api/dashboards/5714/images/3608/image)

[buildstatus]: https://circleci.com/gh/Lusitaniae/phpfpm_exporter/tree/master.svg?style=shield
[circleci]: https://circleci.com/gh/Lusitaniae/phpfpm_exporter
[quay]: https://quay.io/repository/Lusitaniae/phpfpm-exporter
[hub]: https://hub.docker.com/r/lusotycoon/phpfpm-exporter/
[release]: https://github.com/Lusitaniae/phpfpm_exporter/releases/latest
