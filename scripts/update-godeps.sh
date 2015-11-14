#!/usr/bin/env bash
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cd "${dir}/.."

deps=(\
"bitbucket.org/ww/goautoneg" \
"github.com/ant0ine/go-json-rest/rest" \
"github.com/beorn7/perks/quantile" \
"github.com/cilium-team/elastic-go-logging" \
"github.com/cilium-team/mergo" \
"github.com/cilium-team/yaml" \
"github.com/davecgh/go-spew/spew" \
"github.com/deckarep/golang-set" \
"github.com/fsouza/go-dockerclient" \
"github.com/ghodss/yaml" \
"github.com/golang/glog" \
"github.com/golang/protobuf/proto" \
"github.com/google/gofuzz" \
"github.com/juju/ratelimit" \
"github.com/matttproud/golang_protobuf_extensions/pbutil" \
"github.com/op/go-logging" \
"github.com/pborman/uuid" \
"github.com/prometheus/client_golang/prometheus" \
"github.com/prometheus/client_model/go" \
"github.com/prometheus/procfs" \
"github.com/samalba/dockerclient" \
"github.com/spf13/pflag" \
"golang.org/x/net/context" \
"gopkg.in/olivere/elastic.v3" \
"gopkg.in/yaml.v2" \
"speter.net/go/exp/math/dec/inf" \
)

special_deps=(\
"github.com/docker/docker/..." \
"github.com/docker/libcontainer/..." \
"github.com/prometheus/common/..." \
"golang.org/x/crypto/..." \
"k8s.io/kubernetes/..." \
)

echo "Pulling necessary images from DockerHub..."
for dep in "${deps[@]}"; do
    echo "Updating: ${dep}"
    go get -u "${dep}"
    godep update "${dep}"
done

for dep in "${special_deps[@]}"; do
    echo "Updating: ${dep::-4}"
    go get -u "${dep::-4}"
    godep update "${dep}"
done

godep save -r ./...

exit 0
