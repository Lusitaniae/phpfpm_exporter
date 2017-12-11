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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	client_model "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/version"
	"github.com/tomasen/fcgi_client"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	phpfpmSocketPathLabel = "socket_path"
	phpfpmScriptPathLabel = "script_path"

	phpfpmUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("php", "fpm", "up"),
		"Whether scraping PHP-FPM's metrics was successful.",
		[]string{phpfpmSocketPathLabel}, nil)

	phpfpmAcceptedConnections = prometheus.NewDesc(
		prometheus.BuildFQName("php", "fpm", "accepted_connections_total"),
		"Number of request accepted by the pool.",
		[]string{phpfpmSocketPathLabel}, nil)

	phpfpmStartTime = prometheus.NewDesc(
		prometheus.BuildFQName("php", "fpm", "start_time_seconds"),
		"Unix time when FPM has started or reloaded.",
		[]string{phpfpmSocketPathLabel}, nil)

	phpfpmGauges = map[string]*prometheus.Desc{
		"listen queue": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "listen_queue"),
			"Number of request in the queue of pending connections.",
			[]string{phpfpmSocketPathLabel}, nil),
		"max listen queue": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "max_listen_queue"),
			"Maximum number of requests in the queue of pending connections since FPM has started.",
			[]string{phpfpmSocketPathLabel}, nil),
		"listen queue len": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "listen_queue_length"),
			"The size of the socket queue of pending connections.",
			[]string{phpfpmSocketPathLabel}, nil),
		"idle processes": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "idle_processes"),
			"Number of idle processes.",
			[]string{phpfpmSocketPathLabel}, nil),
		"active processes": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "active_processes"),
			"Number of active processes.",
			[]string{phpfpmSocketPathLabel}, nil),
		"max active processes": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "max_active_processes"),
			"Maximum number of active processes since FPM has started.",
			[]string{phpfpmSocketPathLabel}, nil),
		"max children reached": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "max_children_reached"),
			"Number of times, the process limit has been reached.",
			[]string{phpfpmSocketPathLabel}, nil),
		"slow requests": prometheus.NewDesc(
			prometheus.BuildFQName("php", "fpm", "slow_requests"),
			"Enable php-fpm slow-log before you consider this. If this value is non-zero you may have slow php processes.",
			[]string{phpfpmSocketPathLabel}, nil),
	}
)

func CollectStatusFromReader(reader io.Reader, socketPath string, ch chan<- prometheus.Metric) error {
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

func CollectStatusFromSocket(path string, statusPath string, ch chan<- prometheus.Metric) error {

	env := make(map[string]string)
	env["SCRIPT_FILENAME"] = statusPath
	env["SCRIPT_NAME"] = statusPath
	env["REQUEST_METHOD"] = "GET"

	fcgi, err := fcgiclient.Dial("unix", path)
	if err != nil {
		return err
	}
	defer fcgi.Close()

	resp, err := fcgi.Get(env)
	if err != nil {
		return err
	}

	return CollectStatusFromReader(resp.Body, path, ch)
}

func CollectMetricsFromScript(socketPaths []string, scriptPaths []string) ([]*client_model.MetricFamily, error) {
	var result []*client_model.MetricFamily

	for _, socketPath := range socketPaths {

		for _, scriptPath := range scriptPaths {
			fcgi, err := fcgiclient.Dial("unix", socketPath)
			if err != nil {
				return result, err
			}
			defer fcgi.Close()

			env := make(map[string]string)
			env["DOCUMENT_ROOT"] = path.Dir(scriptPath)
			env["SCRIPT_FILENAME"] = scriptPath
			env["SCRIPT_NAME"] = path.Base(scriptPath)
			env["REQUEST_METHOD"] = "GET"

			resp, err := fcgi.Get(env)
			if err != nil {
				return result, err
			}

			var parser expfmt.TextParser
			metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
			if err != nil {
				return result, err
			}

			for _, metricFamily := range metricFamilies {
				for _, metric := range metricFamily.Metric {
					socketPathCopy := socketPath
					scriptPathCopy := scriptPath
					metric.Label = append(
						metric.Label,
						&client_model.LabelPair{
							Name:  &phpfpmSocketPathLabel,
							Value: &socketPathCopy,
						},
						&client_model.LabelPair{
							Name:  &phpfpmScriptPathLabel,
							Value: &scriptPathCopy,
						})
				}
				result = append(result, metricFamily)
			}
		}
	}
	return result, nil
}

type PhpfpmExporter struct {
	socketPaths []string
	statusPath  string
}

func NewPhpfpmExporter(socketPaths []string, statusPath string) (*PhpfpmExporter, error) {
	return &PhpfpmExporter{
		socketPaths: socketPaths,
		statusPath:  statusPath,
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
		err := CollectStatusFromSocket(socketPath, e.statusPath, ch)
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
		listenAddress        = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9253").String()
		metricsPath          = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		socketPaths          = kingpin.Flag("phpfpm.socket-paths", "Paths of the PHP-FPM sockets.").Strings()
		statusPath           = kingpin.Flag("phpfpm.status-path", "Path which has been configured in PHP-FPM to show status page.").Default("/status").String()
		scriptCollectorPaths = kingpin.Flag("phpfpm.script-collector-paths", "Paths of the PHP file whose output needs to be collected.").Strings()
		showVersion          = kingpin.Flag("version", "Print version information.").Bool()
	)

	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	if *showVersion {
		fmt.Println(version.Print("phpfpm_exporter"))
		os.Exit(0)
	}

	exporter, err := NewPhpfpmExporter(*socketPaths, *statusPath)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(exporter)

	if len(*scriptCollectorPaths) != 0 {
		prometheus.DefaultGatherer = prometheus.Gatherers{
			prometheus.DefaultGatherer,
			prometheus.GathererFunc(func() ([]*client_model.MetricFamily, error) {
				return CollectMetricsFromScript(*socketPaths, *scriptCollectorPaths)
			}),
		}
	}

	log.Println("Starting phpfpm_exporter", version.Info())
	log.Println("Build context", version.BuildContext())
	log.Printf("Starting Server: %s", *listenAddress)

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
