package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	c "github.com/cilium-team/cilium/cilium/config"
	h "github.com/cilium-team/cilium/cilium/hook"
	m "github.com/cilium-team/cilium/cilium/messages"
	u "github.com/cilium-team/cilium/cilium/utils"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"
	uprd "github.com/cilium-team/cilium/cilium/utils/profile/runnables/docker"
	upri "github.com/cilium-team/cilium/cilium/utils/profile/runnables/intent"
	uprk "github.com/cilium-team/cilium/cilium/utils/profile/runnables/kubernetes"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	dtypes "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/docker/engine-api/types"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
)

var (
	logLevel          string
	filename          string
	events            bool
	listOnlyForEvents bool
	deleteDB          bool
	flushConfig       bool
	port              int
	log               = logging.MustGetLogger("cilium")
	wg                sync.WaitGroup
	stdoutFormat      = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	fileFormat = logging.MustStringFormatter(
		`%{time:` + time.RFC3339Nano + `} ` + os.Getenv("HOSTNAME") + ` %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
	)
	logsDateFormat    = `-2006-01-02`
	logNameTimeFormat = time.RFC3339
	containersInCache = u.NewSet()
	refreshNetConfig  = 60 //seconds
)

const (
	dockerDaemonPreBaseAddr     = "/docker/daemon/cilium-adapter"
	dockerSwarmPreBaseAddr      = "/docker/swarm/cilium-adapter"
	kubernetesMasterPreBaseAddr = "/kubernetes/master/cilium-adapter"
)

func init() {
	mainFlagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	mainFlagSet.StringVar(&logLevel, "l", "info", "Set log level, valid options are (debug|info|warning|error|fatal|panic)")
	mainFlagSet.StringVar(&filename, "f", "", "Configuration file or directory containing configuration files that will be written in the distributed database (Accepted formats: ProfileFile, DNSConfig and HA-ProxyConfig)")
	mainFlagSet.BoolVar(&deleteDB, "D", false, "Deletes all information inside database")
	mainFlagSet.BoolVar(&flushConfig, "F", false, "Clear configuration but keep state in database")
	mainFlagSet.BoolVar(&events, "e", true, "Listens for docker events so it can automatically clean IPs and configurations used by stopped and deleted containers.")
	mainFlagSet.BoolVar(&listOnlyForEvents, "o", false, "Listen mode only. It only listens for events from a particular docker daemon.")
	mainFlagSet.IntVar(&port, "P", 8080, "Cilium's listening port.")
	mainFlagSet.Parse(os.Args[1:])

	if len(filename) == 0 {
		setupRunnables()
	}

	setupLOG()

	log.Debug("logLevel: %+v", logLevel)
	log.Debug("filename: %+v", filename)
	log.Debug("deleteDB: %+v", deleteDB)
	log.Debug("flushConfig: %+v", flushConfig)
	log.Debug("events: %+v", events)
	log.Debug("listOnlyForEvents: %+v", listOnlyForEvents)
	log.Debug("port: %+v", port)
	log.Debug("HOST_IP = %+v", os.Getenv("HOST_IP"))
	log.Debug("DOCKER_CERT_PATH = %+v", os.Getenv("DOCKER_CERT_PATH"))
	log.Debug("DOCKER_HOST = %+v", os.Getenv("DOCKER_HOST"))
	log.Debug("ELASTIC_PORT = %+v", os.Getenv("ELASTIC_PORT"))
	log.Debug("ELASTIC_IP = %+v", os.Getenv("ELASTIC_IP"))
	log.Debug("PIPEWORK = %+v", os.Getenv("PIPEWORK"))
}

func setupRunnables() {
	log.Debug("Registering runnables")
	// Order matters, we want intent to be the last one so it can perform
	// actions based on all merged configurations and policies.
	if err := upr.Register(uprd.Name, uprd.DockerRunnable{}); err != nil {
		log.Fatal("Failed while registering a runnable: ", err)
	}
	if err := upr.Register(uprk.Name, uprk.KubernetesRunnable{}); err != nil {
		log.Fatal("Failed while registering a runnable: ", err)
	}
	if err := upr.Register(upri.Name, upri.IntentRunnable{}); err != nil {
		log.Fatal("Failed while registering a runnable: ", err)
	}
}

func setupLOG() {
	level, err := logging.LogLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}

	if len(filename) != 0 || deleteDB {
		backend := logging.NewLogBackend(os.Stderr, "", 0)
		oBF := logging.NewBackendFormatter(backend, fileFormat)
		backendLeveled := logging.SetBackend(oBF)
		backendLeveled.SetLevel(level, "")
		log.SetBackend(backendLeveled)
	} else {
		logTimename := time.Now().Format(logNameTimeFormat)
		ciliumLogsDir := os.TempDir() + string(os.PathSeparator) + "cilium-logs"
		if err := os.MkdirAll(ciliumLogsDir, 0755); err != nil {
			log.Error("Error while creating directory: %v", err)
		}
		fo, err := os.Create(ciliumLogsDir + string(os.PathSeparator) + "cilium-log-" + logTimename + ".log")
		if err != nil {
			log.Error("Error while creating log file: %v", err)
		}
		fileBackend := logging.NewLogBackend(fo, "", 0)

		fBF := logging.NewBackendFormatter(fileBackend, fileFormat)

		backend := logging.NewLogBackend(os.Stderr, "", 0)
		oBF := logging.NewBackendFormatter(backend, fileFormat)

		backendLeveled := logging.SetBackend(fBF, oBF)
		backendLeveled.SetLevel(level, "")
		log.SetBackend(backendLeveled)
	}
}

func databaseOperations(delDB, flushCfg bool, fname string) (bool, error) {
	exit := delDB || flushCfg || len(fname) != 0

	if delDB {
		if err := ucdb.InitDb(""); err != nil {
			return exit, err
		}
		log.Info("Database deleted with success")
	}
	if flushCfg {
		if err := ucdb.FlushConfig(""); err != nil {
			return exit, err
		}
		log.Info("Database successfuly cleaned")
	}
	if len(fname) != 0 {
		if err := c.StoreInDB(filename); err != nil {
			return exit, err
		}
		log.Info("File successfuly stored")
	}
	return exit, nil
}

func main() {
	if exit, err := databaseOperations(deleteDB, flushConfig, filename); err != nil {
		log.Error("Error: %+v", err)
		os.Exit(-1)
	} else if exit {
		os.Exit(0)
	}

	dbConn, err := ucdb.NewConn()
	if err != nil {
		log.Error("%+v", err)
	}
	if events {
		dockerclient, err := uc.NewDockerClient()
		if err != nil {
			log.Error("Error: %s", err)
			return
		}

		log.Info("Trying to get docker client info")
		wait := 1 * time.Second
		retries := 5
		for {
			log.Info("Attempt %d...", 6-retries)
			if _, err := dockerclient.Info(); err == nil {
				break
			}
			if retries < 0 {
				log.Error("Unable to monitor for events on the given docker client: ", err)
				return
			}
			time.Sleep(wait)
			wait += wait
			retries--
		}
		log.Info("Connection successful")

		eo := dtypes.EventsOptions{Since: strconv.FormatInt(time.Now().Unix(), 10)}
		r, err := dockerclient.Events(eo)
		if err != nil {
			log.Error("Error %s...", err)
		}
		go listenForEvents(r, dbConn)

		if listOnlyForEvents {
			wg.Add(1)
			log.Info("cilium events only has started")
			wg.Wait()
			log.Info("cilium has exited")
			os.Exit(0)
		}
	}

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"POST", dockerDaemonPreBaseAddr, DockerDaemonRequestsHandler},
		&rest.Route{"POST", dockerSwarmPreBaseAddr, DockerSwarmRequestsHandler},
		&rest.Route{"POST", kubernetesMasterPreBaseAddr, KubernetesMasterRequestHandler},
	)
	if err != nil {
		log.Fatalf("%s", err)
	}
	dockerclient, err := uc.NewDockerClient()
	if err != nil {
		log.Error("%s", err)
	}
	api.SetApp(router)
	log.Info("cilium has started")
	log.Info("Updating state based on the other nodes")

	go func() {
		for {
			timeToProcess1 := time.Now()
			if err := updateEndpoints(dockerclient, dbConn); err != nil {
				log.Error("Error while updating state from other nodes", err)
			}
			timeToProcess2 := time.Now()
			time.Sleep(time.Second*time.Duration(refreshNetConfig) - timeToProcess2.Sub(timeToProcess1))
		}
	}()

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), api.MakeHandler()))

}

func DockerDaemonRequestsHandler(w rest.ResponseWriter, req *rest.Request) {
	RequestsHandler(dockerDaemonPreBaseAddr, w, req)
}

func DockerSwarmRequestsHandler(w rest.ResponseWriter, req *rest.Request) {
	RequestsHandler(dockerSwarmPreBaseAddr, w, req)
}

func KubernetesMasterRequestHandler(w rest.ResponseWriter, req *rest.Request) {
	RequestsHandler(kubernetesMasterPreBaseAddr, w, req)
}

func RequestsHandler(baseAddr string, w rest.ResponseWriter, req *rest.Request) {
	log.Debug("Request received")
	content, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		log.Error("ReadAll: %+v", err.Error())
		rest.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	var powerStripReq m.PowerstripRequest
	if err = m.DecodeRequest(content, &powerStripReq); err != nil {
		log.Error("DecodeRequest: %+v", err.Error())
		rest.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	log.Debug("Request: %+v", powerStripReq)
	hook, err := h.GetHook(powerStripReq.Type)
	if err != nil {
		log.Error("GetHook: %+v", err.Error())
		rest.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	response, err := hook.ProcessRequest(baseAddr, powerStripReq.ClientRequest.Request, content)
	if err != nil {
		log.Warning("ProcessRequest: %+v", err.Error())
		rest.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	log.Debug("Response: %+v", response)
	if err = w.WriteJson(&response); err != nil {
		log.Error("Error WriteJson: ", err.Error())
		rest.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

type Event struct {
	Id     string
	Status string
	From   string
	Time   int64
}

func ParseEvent(eventStr string) Event {
	log.Info("Event received: %+v", eventStr)
	var e Event
	if err := json.Unmarshal([]byte(eventStr), &e); err != nil {
		log.Error("Error while unmarshalling event %+v", e)
	}
	return e
}

func listenForEvents(reader io.ReadCloser, dbConn ucdb.Db) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		e := ParseEvent(scanner.Text())
		go processEvent(e, dbConn)
	}
	if err := scanner.Err(); err != nil {
		log.Error("Error while reading events: %+v", err)
	}
}

func processEvent(event Event, dbConn ucdb.Db) {
	log.Debug("Msg received listen only %s", event)
	switch event.Status {
	case "create":
		maxAttemps := 3
		if containersInCache.Add(event.Id) < maxAttemps {
			log.Info("Adding endpoint for %s", event.Id)
			if err := u.AddEndpoint(dbConn, event.Id); err != nil {
				if attemps := containersInCache.IncFail(event.Id); attemps >= maxAttemps {
					containersInCache.Set(event.Id, u.Failed)
				}
			} else {
				containersInCache.Set(event.Id, u.Configured)
			}
		}
	case "start":
	case "stop":
		fallthrough
	case "destroy":
		fallthrough
	case "die":
		if containersInCache.Remove(event.Id) {
			log.Info("Removing endpoint for %s", event.Id)
			// Only local will be allowed to remove entries
			if event.From == "node:"+os.Getenv("HOSTNAME") ||
				event.From == "self" {
				if containerIPs, err := dbConn.GetEndpoint(event.Id); err == nil {
					for _, ip := range containerIPs.IPs {
						dbConn.DeleteIP(ip)
					}
					u.RemoveLocalEndpoint(dbConn, event.Id)
					dbConn.DeleteEndpoint(event.Id)
				}
			}
			/*if haProxyClient, err := dbConn.GetHAProxyConfig(); err == nil {
				haProxyClient.DeleteBackend(event.Id)
			}*/
			u.RemoveEndpoint(event.Id)
		}
	}
}

func updateEndpoints(dClient uc.Docker, dbConn ucdb.Db) error {
	err := uc.WaitForDockerReady(dClient, 10)
	if err != nil {
		return err
	}
	allContainers, err := dClient.ContainerList(dtypes.ContainerListOptions{All: false})
	if err != nil {
		return err
	}
	for _, container := range allContainers {
		event := Event{
			Id:     container.ID,
			Status: "create",
			From:   "self",
			Time:   time.Now().Unix(),
		}
		go processEvent(event, dbConn)
	}
	return nil
}
