# elastic-go-logging
ElasticSearch backend for go-logging

This package is a ElasticSearch backend log for <https://github.com/op/go-logging>.

**Warning**: this project is for ElasticSearch 2.x, if you want to use a different
ElasticSearch version you have to change the line `e "gopkg.in/olivere/elastic.v3"`
[here](https://github.com/cilium-team/elastic-go-logging/blob/master/elasticgologging.go#L9)
accordingly with the table shown in [https://github.com/olivere/elastic](https://github.com/olivere/elastic#releases).

## Usage

```go
package elasticgologging

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	el "github.com/cilium-team/elastic-go-logging"
	l "github.com/op/go-logging"
)

var (
	appName      = "my-super-app"
	stdoutFormat = l.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	log = l.MustGetLogger(appName)
)

func TestElasticSearch(t *testing.T) {
	level, err := l.LogLevel("DEBUG")
	if err != nil {
		fmt.Errorf("Error while parsing log level %v\n", err)
	}

	elasticBackend, err := el.NewElasticSearchBackendTo("127.0.0.1", "9200", appName, 3)
	if err != nil {
		fmt.Errorf("Error while getting an elastic instance: %v\n", err)
	}

	backend := l.NewLogBackend(os.Stderr, "", 0)
	oBF := l.NewBackendFormatter(backend, stdoutFormat)

	// Setting a both backends, elastic and to output to Stderr.
	backendLeveled := l.SetBackend(elasticBackend, oBF)

	backendLeveled.SetLevel(level, "")
	log.SetBackend(backendLeveled)

	// And we start logging
	log.Info("Hello Info")
	log.Warning("Hello Warning")
	log.Error("Hello Error")
	log.Debug("Hello Debug")

	// Wait until all logs were inserted into elastic search. We set a timeout
	// of 3 seconds, so 6 seconds should be sufficient.
	time.Sleep(6 * time.Second)

	indexName := appName + time.Now().Format(`-2006-01-02`)
	searchResult, err := elasticBackend.Client.Search().Index(indexName).Type("logs").Do()
	if err != nil {
		fmt.Errorf("Error while getting results from elastic instance: %v\n", err)
	}
	for _, item := range searchResult.Each(reflect.TypeOf(map[string]interface{}{})) {
		if i, ok := item.(map[string]interface{}); ok {
			fmt.Printf("Logged into elastic %v\n", i)
		} else {
			fmt.Printf("Something went wrong\n")
		}
	}
	fmt.Printf("Done\n")
}
```

```
$ go test -v example_test.go
23:02:54.633 main ▶ INFO 001 Hello Info
23:02:54.633 main ▶ WARN 002 Hello Warning
23:02:54.633 main ▶ ERRO 003 Hello Error
23:02:54.633 main ▶ DEBU 004 Hello Debug
Logged into elastic map[@timestamp:2015-11-10T23:02:54.633824258Z level:ERROR message:Hello Error node:NODE[127.0.0.1:9200]]
Logged into elastic map[@timestamp:2015-11-10T23:02:54.633799699Z level:WARNING message:Hello Warning node:NODE[127.0.0.1:9200]]
Logged into elastic map[@timestamp:2015-11-10T23:02:54.633837916Z level:DEBUG message:Hello Debug node:NODE[127.0.0.1:9200]]
Logged into elastic map[@timestamp:2015-11-10T23:02:54.633732715Z level:INFO message:Hello Info node:NODE[127.0.0.1:9200]]
```

## Installation

### Using *go get*
After knowing which ElasticSearch version you are running, download `elastic`,
`go-logging` and `elastic-go-logging` with *go get*:

```bash
$ go get "gopkg.in/olivere/elastic.v3"
$ go get "github.com/op/go-logging"
$ go get "github.com/cilium-team/elastic-go-logging"
```
Then you use it with following import path:
```go
import  "github.com/cilium-team/elastic-go-logging"
```
