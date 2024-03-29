// +build linux freebsd

package fileutils

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient/external/github.com/Sirupsen/logrus"
)

// GetTotalUsedFds Returns the number of used File Descriptors by
// reading it via /proc filesystem.
func GetTotalUsedFds() int {
	if fds, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d/fd", os.Getpid())); err != nil {
		logrus.Errorf("Error opening /proc/%d/fd: %s", os.Getpid(), err)
	} else {
		return len(fds)
	}
	return -1
}
