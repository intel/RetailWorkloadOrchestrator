## Deploy Application Workloads

`RWO` uses `docker swarm` for orchestrating workloads on the edge nodes.

It supports deploying an application/service stack using `docker-compose.yml` format. [More Info](https://docs.docker.com/compose/compose-file/).


<a name="startup-workload"></a>
### Workloads at Startup

`RWO` reads specific location `/opt/stacks` for docker-compose based service stack files. If we want to deploy some workload at startup then this option can be used.

Follow these steps before starting `RWO`. You may also stop `RWO`
to do this.

```
# cd /opt/stacks
# mkdir myinitapp
# cd myinitapp
# cp <some docker-compose file> docker-compose.yml
```
It is preferable to run swarm UI portals or management tools here. [Portainer](https://github.com/portainer/portainer-compose/blob/master/docker-stack.yml) and [Swarmpit](https://github.com/swarmpit/swarmpit/blob/master/docker-compose.yml) are two popular web based docker swarm management portals.


<a name="set-up-console"></a>

### SSH Service Console

`RWO` offers a ssh service which usually runs at 22 port.

At installation it is advisable to edit the `/opt/rwo/compose/docker-compose.yml` file to set `username` and `password` for ssh service.

`RWO` uses environment variables to create a user and its password for the ssh into the box.

```yaml
  console:
    image: edge/console-alpine:1.0
    environment:
      - HTTP_PROXY
      - HTTPS_PROXY
      - NO_PROXY
      - PATH=/usr/local/bin:/usr/app-local/bin:/opt/rwo/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin
      - USERNAME=intel-admin
      - PASSWORD=Intel123!
```



### Docker Swarm Stacks

The basic method which `RWO` offers is a ssh service. We can ssh into the node and use `docker swarm` to deploy and scale applications.

```
docker stack deploy -c mycompose.yml mystack

```

<a name="set-up-management-tools"></a>

### Swarm management Tools

One of these can be installed before `RWO` to manage `docker swarm` stacks. [See Here](#startup-workload)

Swarmpit offers a better UI  experience and Portainer offers a feature to manage the cluster remotely.

- [Portainer](https://github.com/portainer/portainer-compose/blob/master/docker-stack.yml)

Portainer supports two methods to manage `docker swarm`

1. `Portainer Web UI` runs on the cluster. We can connect to it through the IP of the manager node.



2. `Portainer Web UI` runs on cloud or another server outside the cluster and `portainer edge agent` runs on the cluster.[*More Info*](https://www.portainer.io/2019/07/portainer-edge-agent/)



- [Swarmpit](https://github.com/swarmpit/swarmpit/blob/master/docker-compose.yml)

With `swarmpit` it is possible to connect to the manager of the swarm and access the web portal for deploying/managing and scaling workloads.



#### Replicated Storage

`RWO` has a docker volume driver called
`edge/glusterfs-plugin{}`. It can be used to create persistent storage over the cluster.

**Create a replicated volume for cluster**

- using cli

```
docker volume create --driver edge/glusterfs-plugin:1.0 some_volume

```


- using docker-compose

We can add the following lines to the `docker-compose.yml` files of workloads/stacks to create a persistent replicated volume on the cluster. 

```yaml
volumes:
  portainer_data:
    driver: edge/glusterfs-plugin:1.0
    driver_opts: {}
```

#### Selective workloads deployment

docker node labels are used to define `constraints` while deploying the service stack.

`RWO` has a component called `dho`, `dynamic hardware orchestrator` which can update node labels with configurations fo the nodes. So it is possible to schedule the workload to the node with specific hardware capabilities.

Example.

```yaml
    deploy:
      mode: replicated
      replicas: 1
      placement:
        constraints: [node.role == manager]
```
