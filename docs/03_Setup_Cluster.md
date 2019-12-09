## Cluster of `EDGE` Nodes

`RWO` enables automatic set up of a cluster of `EDGE` nodes. These nodes can have various compute and storage specifications. Such a cluster is called `heterogenous` cluster of edge nodes.

### Features of RWO for heterogenous cluster

1. `RWO` supports scaling of  workloads on the `edge` node cluster.

2. `RWO` supports a `replicated` high available `glusterfs` storage which can be used via `docker volume plugin`

*NOTE: Good internet connectivity is recommended.*

### Set up Cluster of Nodes

1. Install `RWO` on atleast 3 Nodes. [*See Instructions Here*](01_Install.md)
2. Start `RWO` on one node and wait for few seconds (depends on the compute node's booting capability)
3. Start `RWO` on second node and wait for few seconds (depends on the compute node's booting capability)
4. Start `RWO` on the third node.

The Nodes will automatically form the cluster and you can [deploy workloads](04_Deploy_Workloads.md) on to them.
