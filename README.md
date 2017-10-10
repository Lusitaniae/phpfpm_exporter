# Prometheus phpfpm_exporter

This repository provides code for a Prometheus metrics exporter
for [the php-fpm status page](http://nl1.php.net/manual/en/install.fpm.php). This exporter
provides several metrics listed below. It extracts these metrics from PHP-FPM by talking FastCGI to the UNIX socket.
The UNIX sockets can be specified by setting the `-- phpfpm.socket-paths` flag for each socket. 
The filename of the UNIX socket is added to the metrics as a label.

This exporter also provides a metric collector from php scripts. The scripts can be specified by setting the 
`--phpfpm.script-collector-paths` flag for each script.

# Usage:

``` 
./phpfpm_exporter --phpfpm.socket-paths "/var/run/php5-fpm_example.sock --phpfpm.script-collector-paths /etc/php/script.php"
```

# Metrics:

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

php_script_metric 1
php_script_metric_status 3
```
