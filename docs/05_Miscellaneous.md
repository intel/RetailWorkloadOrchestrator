## Miscellaneous

### Open Ports

This is a snapshot of ports used by running `RWO` service.


```
$ netstat -lntup

Active Internet connections (only servers)

Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
tcp        0      0 127.0.0.1:7373          0.0.0.0:*               LISTEN      2744/serf
tcp        0      0 0.0.0.0:2222            0.0.0.0:*               LISTEN      656/sshd
tcp        0      0 127.0.0.53:53           0.0.0.0:*               LISTEN      530/systemd-resolve
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      3337/sshd
tcp        0      0 0.0.0.0:49152           0.0.0.0:*               LISTEN      3178/glusterfsd
tcp        0      0 0.0.0.0:49153           0.0.0.0:*               LISTEN      4468/glusterfsd
tcp        0      0 0.0.0.0:24007           0.0.0.0:*               LISTEN      2582/glusterd
tcp        0      0 127.0.0.1:5000          0.0.0.0:*               LISTEN      1620/docker-proxy
tcp6       0      0 :::2377                 :::*                    LISTEN      1462/dockerd
tcp6       0      0 :::7945                 :::*                    LISTEN      2744/serf
tcp6       0      0 :::7946                 :::*                    LISTEN      1462/dockerd
tcp6       0      0 :::2222                 :::*                    LISTEN      656/sshd
tcp6       0      0 :::22                   :::*                    LISTEN      3337/sshd
tcp6       0      0 :::9000                 :::*                    LISTEN      1462/dockerd
udp        0      0 127.0.0.53:53           0.0.0.0:*                           530/systemd-resolve
udp        0      0 192.168.2.100:68        0.0.0.0:*                           477/systemd-network
udp        0      0 0.0.0.0:4789            0.0.0.0:*                           -
udp        0      0 0.0.0.0:5353            0.0.0.0:*                           2744/serf
udp6       0      0 :::7945                 :::*                                2744/serf
udp6       0      0 :::7946                 :::*                                1462/dockerd
udp6       0      0 :::5353                 :::*                                2744/serf
```

###
