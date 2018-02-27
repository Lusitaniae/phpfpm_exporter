# PHP-FPM Exporter for Prometheus [![Build Status][buildstatus]][circleci]

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

Running exporter
```
./phpfpm_exporter --phpfpm.socket-paths /var/run/phpfpm.sock --phpfpm.script-collector-paths /path/script.php
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
php_fpm_up{socket_path="/var/run/phpfpm.sock"} 1
```

[buildstatus]: https://circleci.com/gh/Lusitaniae/phpfpm_exporter/tree/master.svg?style=shield
[circleci]: https://circleci.com/gh/Lusitaniae/phpfpm_exporter
[quay]: https://quay.io/repository/Lusitaniae/phpfpm-exporter
[hub]: https://hub.docker.com/r/lusotycoon/phpfpm-exporter/
