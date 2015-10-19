package main

import (
	"flag"
	"fmt"
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
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	dfsouza "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/samalba/dockerclient"
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
		`%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
	)
	logsDateFormat    = `-2006-01-02`
	logNameTimeFormat = time.RFC3339
)

const (
	dockerDaemonPreBaseAddr     = "/docker/daemon/cilium-adapter"
	dockerSwarmPreBaseAddr      = "/docker/swarm/cilium-adapter"
	kubernetesMasterPreBaseAddr = "/kubernetes/master/cilium-adapter"
)

func init() {

	flag.StringVar(&logLevel, "l", "info", "Set log level, valid options are (debug|info|warning|error|fatal|panic)")
	flag.StringVar(&filename, "f", "", "Configuration file or directory containing configuration files that will be written in the distributed database (Accepted formats: ProfileFile, DNSConfig and HA-ProxyConfig)")
	flag.BoolVar(&deleteDB, "D", false, "Deletes all information inside database")
	flag.BoolVar(&flushConfig, "F", false, "Clear configuration but keep state in database")
	flag.BoolVar(&events, "e", true, "Listens for docker events so it can automatically clean IPs and configurations used by stopped and deleted containers.")
	flag.BoolVar(&listOnlyForEvents, "o", false, "Listen mode only. It only listens for events from a particular docker daemon.")
	flag.IntVar(&port, "P", 8080, "Cilium's listening port.")
	flag.Parse()

	if len(filename) == 0 {
		setupRunnables()
	}

	setupLOG()

	log.Debug("Loglevel = %+v", logLevel)
	log.Debug("events = %+v", events)
	log.Debug("listOnlyForEvents = %+v", listOnlyForEvents)
	log.Debug("HOST_IP = %+v", os.Getenv("HOST_IP"))
}

func setupRunnables() {
	log.Debug("Registering runnables")
	if err := upr.Register(uprd.Name, uprd.DockerRunnable{}); err != nil {
		log.Fatal("Failed while registering a runnable", err)
	}
	if err := upr.Register(upri.Name, upri.IntentRunnable{}); err != nil {
		log.Fatal("Failed while registering a runnable", err)
	}
	if err := upr.Register(uprk.Name, uprk.KubernetesRunnable{}); err != nil {
		log.Fatal("Failed while registering a runnable", err)
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
		fo, err := os.Create(os.TempDir() + "/cilium-" + logTimename + ".log")
		fileBackend := logging.NewLogBackend(fo, "", 0)

		db, err := ucdb.NewElasticConn()
		if err != nil {
			log.Error("Error while getting a DB instance: %v", err)
		}
		hn, err := os.Hostname()
		if err != nil {
			log.Debug("Error while getting the hostname: %v", err)
		}

		elasticBackend, err := logging.NewElasticSearchBackendFrom(db.Client, "cilium-log", hn)
		if err != nil {
			log.Error("Error while getting the new logrus hook: %v", err)
		}

		fBF := logging.NewBackendFormatter(fileBackend, fileFormat)

		backend := logging.NewLogBackend(os.Stderr, "", 0)
		oBF := logging.NewBackendFormatter(backend, fileFormat)

		backendLeveled := logging.SetBackend(fBF, elasticBackend, oBF)
		backendLeveled.SetLevel(level, "")
		log.SetBackend(backendLeveled)
	}
}

func main() {
	if deleteDB {
		if err := ucdb.InitDb(""); err != nil {
			log.Error("Error: %s", err)
			os.Exit(-1)
		} else {
			log.Info("Database deleted with success")
		}
	}
	if len(filename) != 0 {
		if err := c.StoreInDB(filename, flushConfig); err != nil {
			log.Error("Error: %s", err)
			os.Exit(-1)
		}
		log.Info("File successfuly stored")
		os.Exit(0)
	} else {
		if deleteDB || flushConfig {
			os.Exit(0)
		}
	}

	dbConn, err := ucdb.NewConn()
	if err != nil {
		log.Error("%+v", err)
	}
	if events {
		dockerclient, err := uc.NewDockerClientSamalba()
		if err != nil {
			log.Error("Error: %s", err)
			return
		}

		log.Info("Trying to get docker client info")
		wait := 1 * time.Second
		retries := 10
		for {
			log.Info("Attempt %d...", 11-retries)
			if _, err := dockerclient.Info(); err == nil {
				break
			}
			if retries < 0 {
				log.Error("Unable to monitor for events on the given docker client")
				return
			}
			time.Sleep(wait)
			wait += wait
			retries--
		}
		log.Info("Connection successful")

		dockerclient.StartMonitorEvents(listenForEvents, nil, dbConn)

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
	if err := updateEndpoints(dockerclient, dbConn); err != nil {
		log.Error("Error while updating state from other nodes", err)
	}

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
	var (
		powerStripReq m.PowerstripRequest
		hook          h.Hook
		response      interface{}
		err           error
	)
	content, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		log.Error("ReadAll: %+v", err.Error())
		rest.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if err = m.DecodeRequest(content, &powerStripReq); err != nil {
		log.Error("DecodeRequest: %+v", err.Error())
		rest.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if hook, err = h.GetHook(powerStripReq.Type); err != nil {
		log.Error("GetHook: %+v", err.Error())
		rest.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if response, err = hook.ProcessRequest(baseAddr, powerStripReq.ClientRequest.Request, content); err != nil {
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

func listenForEvents(event *d.Event, ec chan error, args ...interface{}) {
	if event != nil {
		go func(event d.Event) {
			dbConn := args[0].(ucdb.Db)
			log.Debug("Msg received listen only %s", event)
			switch event.Status {
			case "create":
				log.Info("Adding endpoint for %s", event.Id)
				u.AddEndpoint(dbConn, event.Id)
			case "start":
			case "stop":
			case "destroy":
				fallthrough
			case "die":
				log.Info("Removing endpoint for %s", event.Id)
				if containerIPs, err := dbConn.GetEndpoint(event.Id); err == nil {
					for _, ip := range containerIPs.IPs {
						dbConn.DeleteIP(ip)
					}
					u.RemoveLocalEndpoint(dbConn, event.Id)
					dbConn.DeleteEndpoint(event.Id)
				}
				/*if haProxyClient, err := dbConn.GetHAProxyConfig(); err == nil {
					haProxyClient.DeleteBackend(event.Id)
				}*/
				u.RemoveEndpoint(event.Id)
			}
		}(*event)
	}
}

func updateEndpoints(dClient uc.Docker, dbConn ucdb.Db) error {
	err := uc.WaitForDockerReady(dClient, 10)
	if err != nil {
		return err
	}
	allContainers, err := dClient.ListContainers(dfsouza.ListContainersOptions{All: true})
	if err != nil {
		return err
	}
	for _, container := range allContainers {
		event := d.Event{
			Id:     container.ID,
			Status: "create",
			From:   "self",
			Time:   time.Now().Unix(),
		}
		listenForEvents(&event, nil, dbConn)
	}
	return nil
}
