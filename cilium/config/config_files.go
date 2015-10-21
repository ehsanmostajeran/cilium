package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	upl "github.com/cilium-team/cilium/cilium/utils/plugins/loadbalancer"
	up "github.com/cilium-team/cilium/cilium/utils/profile"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/yaml"
)

var (
	log = logging.MustGetLogger("cilium")
)

func createOwners(conn ucdb.Db, pf up.ProfileFile) (err error) {
	log.Debug("")
	for _, profile := range pf.PolicySource {
		if _, err = conn.PutUser(profile.Owner); err != nil {
			return err
		}
	}
	return nil
}

func storePolicies(conn ucdb.Db, pf up.ProfileFile, basePath string) error {
	log.Debug("")
	for _, profile := range pf.PolicySource {
		for i := range profile.Policies {
			profile.Policies[i].ReadOVSConfigFiles(basePath)
		}
		if err := conn.PutPolicy(profile); err != nil {
			return err
		}
	}
	return nil
}

func StoreInDB(filename string, flushConfig bool) error {
	log.Debug("")
	var (
		conn ucdb.Db
		err  error
	)
	if flushConfig {
		if err = ucdb.FlushConfig(""); err != nil {
			return err
		}
	}
	if conn, err = ucdb.NewConn(); err != nil {
		return err
	}
	defer conn.Close()

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, _ := ioutil.ReadDir(filename)
		if len(files) == 0 {
			log.Info("Empty directory")
			return nil
		}
		for _, f := range files {
			if err := storeFileInDB(conn, filepath.Join(filename, f.Name())); err != nil {
				log.Error("Error: %v", err)
			}
		}
	case mode.IsRegular():
		return storeFileInDB(conn, filename)
	default:
		return fmt.Errorf("Unknown filetype")
	}
	return nil
}

func storeFileInDB(conn ucdb.Db, filename string) error {
	log.Info("Reading file %v", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	pf := up.ProfileFile{}
	dnsConfig := uc.DNSClient{}
	haproxyClient := upl.HAProxyClient{}

	if bytes.HasPrefix(data, []byte("#DNSCONFIG")) {
		err = yaml.Unmarshal(data, &dnsConfig)
		if err != nil {
			return err
		}
		if err = conn.PutDNSConfig(dnsConfig); err != nil {
			return err
		}
	} else if bytes.HasPrefix(data, []byte("#HAPROXYCONFIG")) {
		err = yaml.Unmarshal(data, &haproxyClient)
		if err != nil {
			return err
		}
		if err = conn.PutHAProxyConfig(haproxyClient); err != nil {
			return err
		}
	} else {
		err = yaml.Unmarshal(data, &pf)
		if err != nil {
			return err
		}
		for _, profile := range pf.PolicySource {
			if _, err = conn.PutUser(profile.Owner); err != nil {
				return err
			}
		}
		baseDir, _ := filepath.Split(filename)
		if err = storePolicies(conn, pf, baseDir); err != nil {
			return err
		}
	}
	return nil
}
