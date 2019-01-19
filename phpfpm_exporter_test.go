package main

import (
	"bufio"
	"io"
	"net"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

const (
	testSocket   = "/tmp/phpfpmtest.sock"
	phpfpmStatus = `pool:                 www
process manager:      ondemand
start time:           04/Jan/2019:17:59:31 +0000
start since:          1277299
accepted conn:        51602
listen queue:         0
max listen queue:     14
listen queue len:     128
idle processes:       0
active processes:     1
total processes:      1
max active processes: 42
max children reached: 0
slow requests:        0
`
)

type haproxy struct {
	*httptest.Server
	response []byte
}

func expectMetrics(t *testing.T, c prometheus.Collector, fixture string) {
	exp, err := os.Open(path.Join("test", fixture))
	if err != nil {
		t.Fatalf("Error opening fixture file %q: %v", fixture, err)
	}
	if err := testutil.CollectAndCompare(c, exp); err != nil {
		t.Fatal("Unexpected metrics returned:", err)
	}
}

func newHaproxyUnix(file, statsPayload string) (io.Closer, error) {
	if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	l, err := net.Listen("unix", file)
	if err != nil {
		return nil, err
	}
  					
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
      go func(c net.Conn) {
				defer c.Close()
				r := bufio.NewScanner(c)
				for r.Scan() {
          
          //l, err := r.ReadString('\n')
					//if err != nil {
					//	return
					//}
					switch r.Text() {
					case "show stat\n":
						c.Write([]byte(statsPayload))
						return
					default:
						// invalid command
						return
					}
				}
			}(c)
		}
	}()
	return l, nil
}

func TestUnixDomain(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not on windows")
		return
	}
	srv, err := newHaproxyUnix(testSocket, phpfpmStatus)
	if err != nil {
		t.Fatalf("can't start test server: %v", err)
	}
	defer srv.Close()

  e, err := NewPhpfpmExporter([]string{testSocket}, "/status")
	if err != nil {
		t.Fatal(err)
	}

  expectMetrics(t, e, "unix_domain.metrics")
}
