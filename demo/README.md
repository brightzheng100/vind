# Demo

## General demo

[![asciicast](https://asciinema.org/a/697321.svg)](https://asciinema.org/a/697321)

I used [`asciinema`](https://docs.asciinema.org/) to record the demo, where the demo script is powered by []`demo magic`](https://github.com/paxtonhare/demo-magic) to make the demo process automated and reproducible.

Check out the demo script [here](./demo.sh)


## Demo: Docker in `vind`

To run Docker in `vind`, the only thing we need to do is to make sure the MachineSet specifies `privileged: true`, like [docker-in-vind.yaml](./docker-in-vind.yaml):

```sh
cluster:
  name: cluster
  privateKey: cluster-key
machineSets:
- name: ubuntu
  replicas: 1
  spec:
    image: brightzheng100/vind-ubuntu:22.04
    name: node%d
    networks:
    - my-network
    portMappings:
    - containerPort: 22
    privileged: true      # <-- this does the trick
```

Once you've created the `vind` Machine, and you can `vind ssh` into it and do whatever you typically do in a VM for Docker.


## Demo: Kubernetes in `vind`

Kubernetes in `vind` is much more complicated than one may typically expect.

There are a few tweaks we have to do and the best place to learn is `kind` project's [entrypoint](https://github.com/kubernetes-sigs/kind/blob/main/images/base/files/usr/local/bin/entrypoint) and [provision.go](https://github.com/kubernetes-sigs/kind/blob/main/pkg/cluster/internal/providers/docker/provision.go#L135-L205).

In this demo, I also reuse the work `kind` has done for the tweaks.

### 1. Rebuild the image with `kind`'s `entrypoint` shell script as the entrypoint.

Refer to [Dockerfile](./k8s-in-vind/Dockerfile):

```Dockerfile
...
# Create the /kind folder to facilitate kind's hacking/patching process
RUN mkdir /kind

# The entrypoint is copied from kind project
# Here: https://github.com/kubernetes-sigs/kind/blob/main/images/base/files/usr/local/bin/entrypoint
COPY --chmod=0755 entrypoint /usr/local/bin/entrypoint

# Comment out the default ENTRYPOINT and configure it in vind's YAML
# ENTRYPOINT [ "/usr/local/bin/entrypoint", "/bin/bash" ]
```

Then build it -- make sure we're in `./demo/k8s-in-vind` folder:

By Docker:

```sh
docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile --push -t brightzheng100/vind-ubuntu:k8s .
```

Or, by Podman:

```sh
# 0. If you already have the manifest (for example, by doing more than 1 time)
podman manifest rm brightzheng100/vind-ubuntu:k8s-manifest

# 1. Build them with a manifest specified
podman build --platform linux/amd64,linux/arm64 --file Dockerfile --manifest brightzheng100/vind-ubuntu:k8s-manifest .

# 2. Push manifest with the targeted image tag
podman manifest push brightzheng100/vind-ubuntu:k8s-manifest brightzheng100/vind-ubuntu:k8s
```

### 2. Create `vind` cluster

As usual, create the `vind` config file -- refer to [k8s-in-vind.yaml](./k8s-in-vind/k8s-in-vind.yaml):

```yaml
cluster:
  name: cluster
  privateKey: cluster-key
machineSets:
- name: k8s
  replicas: 3
  spec:
    image: brightzheng100/vind-ubuntu:k8s
    name: node%d
    networks:
    - my-network
    portMappings:
    - containerPort: 22
    privileged: true
    cmd: /usr/local/bin/entrypoint /bin/bash
```

Then, create the cluster:

```sh
# Expose VIND_CONFIG and point to the right vind YAML file
$ export VIND_CONFIG=`PWD`/k8s-in-vind/k8s-in-vind.yaml

# Create the vind cluster
$ vind create

# Where we have 3 machines
$ vind show
CONTAINER NAME      MACHINE NAME   PORTS       IP           IMAGE                            CMD                                    STATE
cluster-k8s-node0   k8s-node0      44107->22   10.89.0.29   brightzheng100/vind-ubuntu:k8s   /usr/local/bin/entrypoint,/sbin/init   Running
cluster-k8s-node1   k8s-node1      46607->22   10.89.0.30   brightzheng100/vind-ubuntu:k8s   /usr/local/bin/entrypoint,/sbin/init   Running
cluster-k8s-node2   k8s-node2      33375->22   10.89.0.31   brightzheng100/vind-ubuntu:k8s   /usr/local/bin/entrypoint,/sbin/init   Running
```

### 3. Bootstrap it 

In `node0`:

```sh
vind ssh ubuntu@k8s-node0
```

Reuse the boostrapping scripts I built with `kubeadm`:

```sh
ubuntu@k8s-node0:~$ sudo apt-get update
ubuntu@k8s-node0:~$ sudo apt-get install git -y
ubuntu@k8s-node0:~$ git clone https://github.com/brightzheng100/instana-handson-labs.git
ubuntu@k8s-node0:~$ cd instana-handson-labs/scripts
ubuntu@k8s-node0:~/instana-handson-labs/scripts$ ./bootstrap-k8s.sh
```

Then the `kubeadm` bootstrapped Kubernetes should be ready in just a few minutes:

```sh
ubuntu@k8s-node0:~/instana-handson-labs/scripts$ kubectl get pod -A -w
NAMESPACE            NAME                                       READY   STATUS    RESTARTS   AGE
kube-system          calico-kube-controllers-5b9b456c66-tsgbc   1/1     Running   0          67s
kube-system          calico-node-crjmb                          1/1     Running   0          67s
kube-system          coredns-55cb58b774-b9twc                   1/1     Running   0          5m36s
kube-system          coredns-55cb58b774-jhnr4                   1/1     Running   0          5m35s
kube-system          etcd-k8s-node0                             1/1     Running   0          5m52s
kube-system          kube-apiserver-k8s-node0                   1/1     Running   0          5m52s
kube-system          kube-controller-manager-k8s-node0          1/1     Running   0          5m52s
kube-system          kube-proxy-47zm2                           1/1     Running   0          5m36s
kube-system          kube-scheduler-k8s-node0                   1/1     Running   0          5m52s
local-path-storage   local-path-provisioner-8ffbb88cb-7dwwv     1/1     Running   0          57s
```

### 4. Join other nodes into the cluster

At the end of the bootstrap command `./bootstrap-k8s.sh`, there should have printed out the join command like:

```sh
kubeadm join 10.89.0.29:6443 --token ypmsa4.1u4ldos4vusk657o \
        --discovery-token-ca-cert-hash sha256:7467a28f6aa2e3ebbc0cbe055a83f5908620c0019fcbee733c1bbed784b8d617
```

Copy and keep it first.

Then, `vind ssh` into other nodes to join the cluster:

#### Join `node1` into the cluster

SSH into machine `node1` first:

```sh
$ vind ssh ubuntu@k8s-node1
```

```sh
ubuntu@k8s-node1:~$ sudo apt-get update
ubuntu@k8s-node1:~$ sudo apt-get install git -y
ubuntu@k8s-node1:~$ git clone https://github.com/brightzheng100/instana-handson-labs.git
ubuntu@k8s-node1:~$ cd instana-handson-labs/scripts
ubuntu@k8s-node1:~/instana-handson-labs/scripts$ ./prepare-join-k8s.sh
```

Now let's slight update the generated `kubeadm join` command with `sudo` (as it needs root permission to run) and `--ignore-preflight-errors=all` flag:

```sh
$ sudo kubeadm join 10.89.0.29:6443 --token ypmsa4.1u4ldos4vusk657o \
        --discovery-token-ca-cert-hash sha256:7467a28f6aa2e3ebbc0cbe055a83f5908620c0019fcbee733c1bbed784b8d617 \
        --ignore-preflight-errors=all
```

Just wait for 1 minute or so, we can check in `node0` (NOT current `node1` as we haven't prepared the kubeconfig yet).

```sh
ubuntu@k8s-node0:~$ kubectl get node
NAME        STATUS   ROLES           AGE     VERSION
k8s-node0   Ready    control-plane   45m     v1.30.9
k8s-node1   Ready    <none>          4m26s   v1.30.9
```

Yes, our cluster has two nodes!

> Note: it may take slightly longer time to turn the status into `Ready`. That's fine and just wait for a while.

#### Join `node2` into the cluster

Similarly, SSH into machine `node2`:

```sh
$ vind ssh ubuntu@k8s-node2
```

```sh
ubuntu@k8s-node2:~$ sudo apt-get update
ubuntu@k8s-node2:~$ sudo apt-get install git -y
ubuntu@k8s-node2:~$ git clone https://github.com/brightzheng100/instana-handson-labs.git
ubuntu@k8s-node2:~$ cd instana-handson-labs/scripts
ubuntu@k8s-node2:~/instana-handson-labs/scripts$ ./prepare-join-k8s.sh
```

Copy and run above updated `kubeadm join` command:

```sh
$ sudo kubeadm join 10.89.0.29:6443 --token ypmsa4.1u4ldos4vusk657o \
        --discovery-token-ca-cert-hash sha256:7467a28f6aa2e3ebbc0cbe055a83f5908620c0019fcbee733c1bbed784b8d617 \
        --ignore-preflight-errors=all
```

Just wait for 1 minute or so, we can check in `node0` (NOT current `node2` as we haven't prepared the kubeconfig yet).

```sh
ubuntu@k8s-node0:~$ kubectl get node
NAME        STATUS   ROLES           AGE   VERSION
k8s-node0   Ready    control-plane   51m   v1.30.9
k8s-node1   Ready    <none>          11m   v1.30.9
k8s-node2   Ready    <none>          91s   v1.30.9
```

Yes, our cluster has three nodes!

> Note: it may take slightly longer time to turn the status into `Ready`. That's fine and just wait for a while.

And very soon, the DaemonSets will automatically roll out the pods to all three nodes:

```sh
ubuntu@k8s-node0:~$ kubectl get pod -A
NAMESPACE            NAME                                       READY   STATUS    RESTARTS   AGE
kube-system          calico-kube-controllers-5b9b456c66-tsgbc   1/1     Running   0          47m
kube-system          calico-node-9gq4l                          1/1     Running   0          2m19s
kube-system          calico-node-crjmb                          1/1     Running   0          47m
kube-system          calico-node-w2x5k                          1/1     Running   0          12m
kube-system          coredns-55cb58b774-b9twc                   1/1     Running   0          52m
kube-system          coredns-55cb58b774-jhnr4                   1/1     Running   0          52m
kube-system          etcd-k8s-node0                             1/1     Running   0          52m
kube-system          kube-apiserver-k8s-node0                   1/1     Running   0          52m
kube-system          kube-controller-manager-k8s-node0          1/1     Running   0          52m
kube-system          kube-proxy-47zm2                           1/1     Running   0          52m
kube-system          kube-proxy-f8g4m                           1/1     Running   0          12m
kube-system          kube-proxy-txqnw                           1/1     Running   0          2m19s
kube-system          kube-scheduler-k8s-node0                   1/1     Running   0          52m
local-path-storage   local-path-provisioner-8ffbb88cb-7dwwv     1/1     Running   0          47m
```
