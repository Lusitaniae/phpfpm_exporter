// Copyright 2017 Kumina, https://kumina.nl/
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tomasen/fcgi_client"
)

var (
	phpfpmUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("php", "fpm", "up"),
		"Whether scraping PHP-FPM's metrics was successful.",
		[]string{"socket_path"}, nil)

	phpfpmAcceptedConnections = prometheus.NewDesc(
		prometheus.BuildFQName("php", "fpm", "accepted_connections_total"),
		"Number of request accepted by the pool.",
		[]string{"socket_path"}, nil)

	phpfpmStartTime = prometheus.NewDesc(
		prometheus.BuildFQName("php", "fpm", "start_time_seconds"),
		"Unix time when FPM has started or reloaded.",
		[]string{"socket_path"}, nil)

	phpfpmGauges = map[string]*prometheus.Desc{
		"listen queue": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "listen_queue"),
			"Number of request in the queue of pending connections.",
			[]string{"socket_path"}, nil),
		"max listen queue": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "max_listen_queue"),
			"Maximum number of requests in the queue of pending connections since FPM has started.",
			[]string{"socket_path"}, nil),
		"listen queue len": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "listen_queue_length"),
			"The size of the socket queue of pending connections.",
			[]string{"socket_path"}, nil),
		"idle processes": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "idle_processes"),
			"Number of idle processes.",
			[]string{"socket_path"}, nil),
		"active processes": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "active_processes"),
			"Number of active processes.",
			[]string{"socket_path"}, nil),
		"max active processes": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "max_active_processes"),
			"Maximum number of active processes since FPM has started.",
			[]string{"socket_path"}, nil),
		"max children reached": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "max_children_reached"),
			"Number of times, the process limit has been reached.",
			[]string{"socket_path"}, nil),
		"slow requests": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "slow_requests"),
			"Enable php-fpm slow-log before you consider this. If this value is non-zero you may have slow php processes.",
			[]string{"socket_path"}, nil),
	}
)

// Converts the output of Dovecot's EXPORT command to metrics.
func CollectFromReader(reader io.Reader, socketPath string, ch chan<- prometheus.Metric) error {
	scanner := bufio.NewScanner(reader)
	re := regexp.MustCompile("^(.*): +(.*)$")

	// Scrape the interesting values:
	for scanner.Scan() {
		fields := re.FindStringSubmatch(scanner.Text())
		if fields == nil {
			return fmt.Errorf("Failed to parse %s", scanner.Text())
		}

		if gauge, ok := phpfpmGauges[fields[1]]; ok {
			f, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return err
			}
			ch <- prometheus.MustNewConstMetric(
				gauge,
				prometheus.GaugeValue,
				f,
				socketPath)
		} else if fields[1] == "accepted conn" {
			f, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return err
			}
			ch <- prometheus.MustNewConstMetric(
				phpfpmAcceptedConnections,
				prometheus.CounterValue,
				f,
				socketPath)
		} else if fields[1] == "start time" {
			location, err := time.LoadLocation("Local")
			if err != nil {
				return err
			}
			since, err := time.ParseInLocation("02/Jan/2006:15:04:05 -0700", fields[2], location)
			if err != nil {
				return err
			}
			f := float64(since.Unix())
			ch <- prometheus.MustNewConstMetric(
				phpfpmStartTime,
				prometheus.GaugeValue,
				f,
				socketPath)
		}
	}
	return nil
}

func CollectFromSocket(path string, scriptName string, ch chan<- prometheus.Metric) error {

	env := make(map[string]string)
	env["SCRIPT_FILENAME"] = scriptName
	env["SCRIPT_NAME"] = scriptName
	env["REQUEST_METHODGET"] = "GET"

	fcgi, err := fcgiclient.Dial("unix", path)
	if err != nil {
		return err
	}

	resp, err := fcgi.Get(env)
	if err != nil {
		return err
	}

	return CollectFromReader(resp.Body, path, ch)
}

type PhpfpmExporter struct {
	socketPaths []string
	scriptName  string
}

func NewPhpfpmExporter(socketPaths []string, scriptName string) (*PhpfpmExporter, error) {
	return &PhpfpmExporter{
		socketPaths: socketPaths,
		scriptName:  scriptName,
	}, nil
}

func (e *PhpfpmExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- phpfpmUpDesc
	ch <- phpfpmAcceptedConnections
	ch <- phpfpmStartTime
	for _, desc := range phpfpmGauges {
		ch <- desc
	}
}

func (e *PhpfpmExporter) Collect(ch chan<- prometheus.Metric) {
	for _, socketPath := range e.socketPaths {
		err := CollectFromSocket(socketPath, e.scriptName, ch)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				phpfpmUpDesc,
				prometheus.GaugeValue,
				1.0,
				socketPath)
		} else {
			log.Printf("Failed to scrape socket: %s", err)
			ch <- prometheus.MustNewConstMetric(
				phpfpmUpDesc,
				prometheus.GaugeValue,
				0.0,
				socketPath)
		}
	}
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9253", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		socketPaths   = flag.String("phpfpm.socket-paths", "", "Paths of the PHP-FPM sockets.")
		scriptName = flag.String("phpfpm.scriptname", "/server-status-fpm.php", "Scriptname for fcgi socket communication.")
	)
	flag.Parse()

	exporter, err := NewPhpfpmExporter(strings.Split(*socketPaths, ","), *scriptName)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>PHP-FPM Exporter</title></head>
			<body>
			<h1>PHP-FPM Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
