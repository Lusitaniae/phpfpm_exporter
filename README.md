# PHP-FPM Exporter for Prometheus [![Build Status][buildstatus]][circleci]

[![Docker Repository on Quay](https://quay.io/repository/Lusitaniae/phpfpm-exporter/status)][quay]
[![Docker Pulls](https://img.shields.io/docker/pulls/lusotycoon/phpfpm-exporter.svg?maxAge=604800)][hub]

This repository provides code for a Prometheus metrics exporter
for [the PHP-FPM status page](http://nl1.php.net/manual/en/install.fpm.php).
This exporter provides several metrics listed below. It extracts these
metrics from PHP-FPM by talking to FastCGI over its UNIX socket. The UNIX
sockets can be specified by adding a `--phpfpm.socket-paths` flag for
each socket to be monitored. The filename of the socket is added to each
metric as a label.

This exporter also provides a way for embedding the output of arbitrary
PHP scripts into its metrics page, analogous to the node exporter's
`textfile` collector. Scripts that are specified with the
`--phpfpm.script-collector-paths` flag will be run through PHP-FPM. Any
metrics printed by the PHP script will be merged into the metrics
provided by this exported. An example use case includes printing metrics
for PHP's `opcache`.

# Usage:

```
./phpfpm_exporter --phpfpm.socket-paths /var/run/php5-fpm_example.sock --phpfpm.script-collector-paths /some/php/script.php
```

# Metrics emitted by PHP-FPM:

```
php_fpm_accepted_connections_total{socket_path="/var/run/php5-fpm_example.sock"} 300940
php_fpm_active_processes{socket_path="/var/run/php5-fpm_example.sock"} 1
php_fpm_idle_processes{socket_path="/var/run/php5-fpm_example.sock"} 5
php_fpm_listen_queue{socket_path="/var/run/php5-fpm_example.sock"} 0
php_fpm_listen_queue_length{socket_path="/var/run/php5-fpm_example.sock"} 0
php_fpm_max_active_processes{socket_path="/var/run/php5-fpm_example.sock"} 10
php_fpm_max_children_reached{socket_path="/var/run/php5-fpm_example.sock"} 3
php_fpm_max_listen_queue{socket_path="/var/run/php5-fpm_example.sock"} 0
php_fpm_slow_requests{socket_path="/var/run/php5-fpm_example.sock"} 0
php_fpm_start_time_seconds{socket_path="/var/run/php5-fpm_example.sock"} 1.49277445e+09
php_fpm_up{socket_path="/var/run/php5-fpm_example.sock"} 1
```

[buildstatus]: https://circleci.com/gh/Lusitaniae/phpfpm_exporter/tree/master.svg?style=shield
[circleci]: https://circleci.com/gh/Lusitaniae/phpfpm_exporter
[quay]: https://quay.io/repository/Lusitaniae/phpfpm-exporter
[hub]: https://hub.docker.com/r/lusotycoon/phpfpm-exporter/
