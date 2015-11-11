package elasticgologging

import (
	"fmt"
	"os"
	"time"

	l "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
	e "github.com/cilium-team/cilium/Godeps/_workspace/src/gopkg.in/olivere/elastic.v3"
)

const (
	defaultTimeoutSecond = 20
	defaultMinActions    = 20
	typeName             = "logs"
)

type elasticSearchBackend struct {
	Client     *e.Client
	appName    string
	nodeName   string
	logMsgChan chan map[string]string
	timeout    <-chan time.Time
}

func NewElasticSearchBackendFrom(ec *e.Client, appName, nodeName string, timeout int64) (*elasticSearchBackend, error) {
	if ec == nil {
		return nil, fmt.Errorf("elastic Client is nil")
	}
	if timeout == 0 {
		timeout = defaultTimeoutSecond
	}
	c := elasticSearchBackend{
		appName:    appName,
		Client:     ec,
		nodeName:   nodeName,
		logMsgChan: make(chan map[string]string, 5*defaultMinActions),
		timeout:    time.NewTicker(time.Duration(timeout) * time.Second).C,
	}
	go c.log()
	return &c, nil
}

func NewElasticSearchBackendTo(ip, port, appName string, timeout int64) (*elasticSearchBackend, error) {
	c, err := e.NewClient(
		e.SetURL("http://"+ip+":"+port),
		e.SetMaxRetries(10))
	if err != nil {
		return nil, err
	}
	nodeName := fmt.Sprintf("NODE[%s:%s]", ip, port)
	return NewElasticSearchBackendFrom(c, appName, nodeName, timeout)
}

func (eSB *elasticSearchBackend) log() {
	currIndexName := eSB.appName + time.Now().Format(`-2006-01-02`)
	bulkReq := eSB.Client.Bulk().Index(currIndexName)
	send := func() {
		if _, err := bulkReq.Do(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send bulk to elasticsearch: %v\n", err)
		}
		bulkReq = eSB.Client.Bulk().Index(currIndexName)
		currIndexName = eSB.appName + time.Now().Format(`-2006-01-02`)
	}
	for {
		select {
		case msg := <-eSB.logMsgChan:
			bulkReq.Add(e.NewBulkIndexRequest().Index(currIndexName).Type(typeName).Doc(msg))
			if bulkReq.NumberOfActions() > defaultMinActions {
				send()
			}
		case <-eSB.timeout:
			if bulkReq.NumberOfActions() > 0 {
				send()
			}
		}
	}
}

func (eSB *elasticSearchBackend) Log(level l.Level, calldepth int, rec *l.Record) error {
	eSB.logMsgChan <- map[string]string{
		"level":      level.String(),
		"message":    rec.Message(),
		"@timestamp": rec.Time.Format(time.RFC3339Nano),
		"node":       eSB.nodeName,
	}
	return nil
}
