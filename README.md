# cilium: Operational Constraints and Policies for Container Clusters

![cilium-architecture](./docs/kibana.png)

Cilium is an open source project to define and enforce operational constraints
and policies for container clusters. Policies are defined in YAML and associated
with containers using labels. The enforcement of the policies is enforced by
either cilium itself or plumbing plugins depending on the policy specified.

This initial version is bound to Docker using powerstrip but the architecture
allows for integration with Kubernetes and Mesosphere as well.

TOC
===
  * [How to run cilium](#how-to-run-cilium)
    * [Checklist](#checklist)
      * [Requirements for each node](#requirements-for-each-node)
      * [Re-check all dependencies on each node](#re-check-all-dependencies-on-each-node)
      * [Preparing node](#preparing-node)
      * [Infecting a node](#infecting-a-node)
      * [Starting services](#starting-services)
  * [Compose demo](#compose-demo)
    * [Getting docker-compose and app-policy](#getting-docker-compose-and-app-policy)
  * [Kibana](#kibana)
  * [F.A.Q.](#faq)
    * [Why am I getting json: cannot unmarshal number into Go value of type []types.Container?](#why-am-i-getting-json-cannot-unmarshal-number-into-go-value-of-type-typescontainer)
    * [Why am I getting a Error: dial unix /var/run/docker.sock: permission denied?](#why-am-i-getting-a-error-dial-unix-varrundockersock-permission-denied)
    * [Why do I have to open port all of those ports in my firewall?](#why-do-i-have-to-open-port-all-of-those-ports-in-my-firewall)
  * [License](#license)

# How to run cilium

There are 2 configurations to install cilium and you should follow only one of
those steps depending your final goal:
- [infecting a node](#checklist)
- [development](./docs/CONTRIBUTING.md#installation-for-developers)

## Checklist

If you want to test `cilium` in your cluster make sure you have the following
requirements:

### Requirements for each node

- Docker (>=1.8.0)
- Open vSwitch (>=2.3.2) (for each node)
- ~5 GB free disk (for each node)
- Compose (>=1.3.0) (for demo)

### Re-check all dependencies on each node

Make sure you have all requirements checked.

* `curl -Ssl https://raw.githubusercontent.com/cilium-team/cilium/master/entrypoint.sh | bash -s check`

### Preparing node

You can skip this step but we don't recommend it to skip it since it will allow
a faster deployment of cilium's components on the specific node. This step will
pull all docker images that we'll be using on the next steps.

* `curl -Ssl https://raw.githubusercontent.com/cilium-team/cilium/master/entrypoint.sh | bash -s prepare`

### Infecting a node

We will now infect a node with `cilium` components (more info about these
components [here](docs/CONTRIBUTING.md#what-is-a-policy-file)) that will allow us to
deploy your applications into your `cilium` cluster.

![cilium-architecture](docs/node-architecture.png)

*You will need to open the ports 80, 2371, 2373, 5000, 8080, 8300, 8301
(tcp/udp), 8302 (tcp/udp), 8400, 8500, 9200 and 9300 on your firewall to receive
tcp traffic and 53, 4789, 54328 for udp traffic.
([Why?](#why-do-i-have-to-open-port-all-of-those-ports-in-my-firewall))*

On the first node of a cluster that will be infected, you have to provide the
following environment variables:

* `IP` - reachable node's IP from all remaining nodes. __Don't__ use
`127.0.0.1`. For example, `192.168.50.37` is fine.
* `NET_IP` - network address of the reachable node's IP. For example,
`192.168.50.0/24` if the node is on a network with the netmask `255.255.255.0`.

In our example we have:

* `# curl -Ssl https://raw.githubusercontent.com/cilium-team/cilium/master/entrypoint.sh | NET_IP=192.168.50.0/24 IP=192.168.50.37 bash -s infect`

* `curl -Ssl https://raw.githubusercontent.com/cilium-team/cilium/master/entrypoint.sh | NET_IP=<node's network address> IP=<node's IP address> bash -s infect`

If you only have one node in your cluster that's ok, you can go to [starting services](#starting-services)
to continue.

The second and remaining nodes will only require the following environment
variables:

* `IP` - reachable node's IP from all remaining nodes. __Don't__ use
`127.0.0.1`. For example `192.168.50.38` is fine.
* `MASTER_IP` - network address of __one__ of the infected nodes. For example,
the first node we infected had the IP `192.168.50.37` so we will use this
one as `MASTER_IP`.

In our example we have:

* `# curl -Ssl https://raw.githubusercontent.com/cilium-team/cilium/master/entrypoint.sh | MASTER_IP=192.168.50.38 IP=192.168.50.37 bash -s infect`

* `curl -Ssl https://raw.githubusercontent.com/cilium-team/cilium/master/entrypoint.sh | MASTER_IP=<An already infected node's IP address> IP=<node's IP address> bash -s infect`

At this point you should have 9 containers on each node. For example, in one of
the nodes we have:

```bash
$ DOCKER_HOST=127.0.0.1:2375 docker ps -a  --format 'table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}'
CONTAINER ID        IMAGE                              STATUS              NAMES
d141780d59b2        cilium/docker-collector:latest     Up 6 minutes        node2/cilium-docker-collector
cf3e329c9a24        cilium/cilium                      Up 6 minutes        node2/cilium-swarm-event-handler
f4ca2699a644        swarm:0.4.0                        Up 6 minutes        node2/cilium-swarm-master
5983cec8cc1b        swarm:0.4.0                        Up 6 minutes        node2/cilium-swarm-agent
24af1ac8e4a3        cilium/powerstrip:latest           Up 7 minutes        node2/cilium-powerstrip-pre-daemon
3befe78a69cc        cilium/powerstrip:latest           Up 7 minutes        node2/cilium-powerstrip-pre-swarm
a80915511fe2        cilium/cilium                      Up 7 minutes        node2/cilium
6cf49abe8237        elasticsearch:1.7.1                Up 7 minutes        node2/cilium-elastic
30343151fde0        progrium/consul                    Up 7 minutes        node2/cilium-consul
```

### Starting services

Make sure you have infected all nodes that you want before going through this
step.

Now that we have our cluster ready to receive some containers , we can run some
services such as a load balancer and a DNS.

* `curl -Ssl https://raw.githubusercontent.com/cilium-team/cilium/master/entrypoint.sh | bash -s start-services`

Now you should have 2 more containers running in your cluster:
```
$ DOCKER_HOST=127.0.0.1:2375 docker ps -a  --format 'table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}'
CONTAINER ID        IMAGE                                    STATUS              NAMES
bd820c3fa3fc        cilium/docker-dns-rest:1.0-rr-with-del   Up 6 minutes        localhost/cilium-dns
753a75b39dff        tnolet/haproxy-rest                      Up 6 minutes        node1/cilium-loadbalancer
...
```

Congratulations you have setup a `cilium` cluster! Go to the [compose-example](#compose-demo)
step to complete the demo.

# Compose demo

To continue this demo please make sure you've [docker-compose](https://docs.docker.com/compose/install/)
(>=1.3.0) installed in the node were you'll execute the demo.

## Getting docker-compose and app-policy

Get the following files:
```bash
$ curl -Ssl -o docker-compose.yml https://raw.githubusercontent.com/cilium-team/cilium/master/examples/compose/docker-compose.yml
$ curl -Ssl -o app-policy.yml https://raw.githubusercontent.com/cilium-team/cilium/master/examples/compose/app-policy.yml
```

Inside that directory you should have:

* `docker-compose.yml` - is a normal compose file without any modifications.
* `app-policy.yml` - is cilium's policies that will be enforced accordingly
the given coverage. To understand how this enforced go to
[what options do you have available?](./docs/CONTRIBUTING.md#what-options-do-you-have-available).

We have to store `app-policy.yml` in our distributed database. For so, run:

```
docker run --rm \
    -e ELASTIC_IP=<any-nodes-reachable-IP> \
    -v $PWD/app-policy.yml:/opt/cilium/policies/ \
    cilium/cilium -f /opt/cilium/policies
```

*Don't forget it should be the full path or it won't work, that's why we have
written the $PWD*

Now run `DOCKER_HOST=127.0.0.1:2375 docker-compose up -d`, this will start
both services (`redis` and `web`).
```bash
$ DOCKER_HOST=127.0.0.1:2375 docker ps -a
CONTAINER ID        IMAGE                    COMMAND                  CREATED             STATUS              PORTS               NAMES
168640ceb0ae        cilium/compose-example   "python app.py"          2 minutes ago       Up 2 minutes                            node1/compose_web_1
2d5b35eb1dd5        redis                    "/entrypoint.sh redis"   3 minutes ago       Up 3 minutes                            node1/compose_redis_1
...
```
*If redis is running in your physical machine and you have a firewall enabled,
you must open port 6379*

To test `web` container, open a new terminal and find out the tcp ports the
load-balancer has open.
```bash
$ DOCKER_HOST=127.0.0.1:2375 docker ps -a --filter=name=cilium-loadbalancer
CONTAINER ID        IMAGE                 COMMAND             CREATED             STATUS              PORTS                                                                                             NAMES
6d96555a4395        tnolet/haproxy-rest   "/haproxy-rest"     17 minutes ago      Up 17 minutes       192.168.50.5:1988->1988/tcp, 192.168.50.5:5000->5000/tcp, 80/tcp, 192.168.50.5:10001->10001/tcp   node1/cilium-loadbalancer
```

I our case we can see there's the `192.168.50.5:5000`, in your case should be
the same IP of one of the cluster's nodes with the port 5000.
The load balancer round-robin the requests through `web` containers available.
Since we only have 1 container running service `web`, all requests go to that
container.

```bash
$ curl 192.168.50.5:5000
Hello World! I have been seen 1 times.
```

Next we are going to scale up `web` to 3 containers. Go to the terminal were
you have the `docker-compose.yml` file and run:

```bash
$ DOCKER_HOST=localhost:2375 docker-compose scale web=3
```
You have successfully scaled the `web` service on to 3 containers.

To see it, run:
```bash
$ DOCKER_HOST=localhost:2375 docker-compose logs web
```
And leave this terminal open.

On the other terminal run *don't forget in our example we have 192.168.50.5 and
your IP might be different*:

```bash
$ curl 192.168.50.5:5000
Hello World! I have been seen 2 times.
$ curl 192.168.50.5:5000
Hello World! I have been seen 3 times.
```

and you can see under the first terminal that the requests are being round-robin
through the 3 web containers:

```bash
web_1 | 1.1.0.251 - - [08/Jul/2015 02:42:06] "GET / HTTP/1.1" 200 -
web_2 | 1.1.0.251 - - [08/Jul/2015 02:42:07] "GET / HTTP/1.1" 200 -
web_3 | 1.1.0.251 - - [08/Jul/2015 02:42:08] "GET / HTTP/1.1" 200 -
```

# Kibana

All GETs previously performed were counted by `cilium-docker-collector`
container.

You can see those statistics them by running:

* `curl -Ssl https://raw.githubusercontent.com/cilium-team/cilium/master/entrypoint.sh | IP=<node's IP address> bash -s start-kibana`

* Open your browser in [kibana's dashboard](http://127.0.0.1:5601/#/dashboard/Cilium-dashboard?_g=(refreshInterval:(display:'30%20seconds',pause:!f,section:1,value:30000),time:(from:now-15m,mode:quick,to:now))&_a=(filters:!(),query:(query_string:(analyze_wildcard:!t,query:'*')),title:Cilium-dashboard))

*If the webpage is not available make sure the IP is the same IP of the node
where you are running Kibana and the port 5601 is open that node's firewall*

* You'll see pretty graphs with network traffic logged.

You can also see some ElasticSearch cluster statistics under [http://127.0.0.1:9200/_plugin/marvel/](http://localhost:9200/_plugin/marvel/)

# F.A.Q.

## Why am I getting `json: cannot unmarshal number into Go value of type []types.Container`?

This is a bug that occurs when powerstrip tries to communicate with a local
swarm that hasn't the `Role: replica`. Has a workaround you can perform the same
request on the `Role: primary`. To find out, run:

```bash
$ DOCKER_HOST=192.168.50.1:2375 docker info | grep Primary
Primary: 192.168.50.5:2373
```

then you can run your requests as `DOCKER_HOST=192.168.50.3:2375 docker ps -a`

## Why am I getting a `Error: dial unix /var/run/docker.sock: permission denied`?

If you don't have reading and writing permissions to that particular file, you
need to change them so your user can read and write on that file.

## Why do I have to open port all of those ports in my firewall?

- Powerstrip works by sending docker requests to powerstrip-adapters, since
cilium is a powerstrip adapter it will have a server listening on port 8080 but
it is on your local machine. Since powerstrip and cilium runs inside a docker
container, powerstrip needs to contact cilium somehow and that's through the IP
address that you are prompt when you run `infect`.
- Port 53 and 80 are for cilium-dns (one is for dns requests/reply) the last one
is for configuration.
- Ports 2371 and 2373 are for cilium to contact swarm containers.
- Port 4789 is for the VXLAN tunnel.
- Port 5000 is for the compose example.
- Port 9200 is for kibana to contact elasticsearch.
- Port 9300 is for elasticsearch, the distributed database, to exchange
information between nodes.
- Ports 8300, 8301, 8302, 8400 and 8500 are for consul.
- Port 54328 is also for elasticsearch to receive multicast pings from the
remaining elasticsearch nodes.

# License

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License.  You may obtain a copy of the
License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied.  See the License for the
specific language governing permissions and limitations under the License.
