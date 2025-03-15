# Virtual Machines IN Docker (`vind`)

`vind`, the name comes from <u>**V**</u>M <u>**IN**</u> <u>**D**</u>ocker, is a tool to create and manage a cluster of containers that look and work like virtual machines, on Docker (and Podman).

The container, or called **Machine** in `vind`, runs `systemd` as PID 1 and a SSH daemon that can be used to log into.
Such VM-like container behaves very much like a "normal" VM, it's even possible to run `dockerd` or `kubernetes` in it.

[![asciicast](https://asciinema.org/a/697321.svg)](https://asciinema.org/a/697321)

> Note: `vind` is a rebuild on top of weaveworks' [footloose](https://github.com/weaveworks/footloose), which was archived in year 2023. Kudos to the original developers!


## Install

`vind` binaries can be downloaded from this repo's [release page](https://github.com/brightzheng100/vind/releases).

### MacOS

On ARM Mx chip:

```sh
LATEST_VERSION=`curl -s "https://api.github.com/repos/brightzheng100/vind/releases/latest" | grep '"tag_name":' | cut -d '"' -f 4 | cut -c 2-`
curl -Lo vind.tar.gz https://github.com/brightzheng100/vind/releases/download/v${LATEST_VERSION}/vind_${LATEST_VERSION}_darwin_arm64.tar.gz
tar -xvf vind.tar.gz && chmod +x vind
sudo mv vind /usr/local/bin/
```

On Intel chip:

```sh
LATEST_VERSION=`curl -s "https://api.github.com/repos/brightzheng100/vind/releases/latest" | grep '"tag_name":' | cut -d '"' -f 4 | cut -c 2-`
curl -Lo vind.tar.gz https://github.com/brightzheng100/vind/releases/download/v${LATEST_VERSION}/vind_${LATEST_VERSION}_darwin_amd64.tar.gz
tar -xvf vind.tar.gz && chmod +x vind
sudo mv vind /usr/local/bin/
```

### Linux

On AMD64 / x86_64 CPU:

```sh
LATEST_VERSION=`curl -s "https://api.github.com/repos/brightzheng100/vind/releases/latest" | grep '"tag_name":' | cut -d '"' -f 4 | cut -c 2-`
curl -Lo vind.tar.gz https://github.com/brightzheng100/vind/releases/download/v${LATEST_VERSION}/vind_${LATEST_VERSION}_linux_amd64.tar.gz
tar -xvf vind.tar.gz && chmod +x vind
sudo mv vind /usr/local/bin/
```

On ARM64 CPU:

```sh
LATEST_VERSION=`curl -s "https://api.github.com/repos/brightzheng100/vind/releases/latest" | grep '"tag_name":' | cut -d '"' -f 4 | cut -c 2-`
curl -Lo vind.tar.gz https://github.com/brightzheng100/vind/releases/download/v${LATEST_VERSION}/vind_${LATEST_VERSION}_linux_arm64.tar.gz
tar -xvf vind.tar.gz && chmod +x vind
sudo mv vind /usr/local/bin/
```

### Windows

It should just work as the binaries are cross compiled.
But I personally haven't tried it yet. So please raise GitHub issues if there is any.

## Concepts

There are some simple concepts in `vind`:

- **`Machine`**: A Machine is a VM-like container that is created by the configured **`MachineSet`**'s specification.
- **`MachineSet`**: A MachineSet is a set of Machines that share the same configuration specification. One MachineSet can have 1 or more replicas, each of which represents a Machine. Each Machine in the MachineSet has its own index, starting from 0.
- **`Cluster`**: Cluster is the top level of objects in `vind`. A Cluster is a group of MachineSet(s), which has a unique name and authentication SSH key pair for the underlying Machines. The SSH key pair can be generated automatically if not exists, or you can generate it and assign to the cluster through the configuration YAML file.


## Config File & Lookup Strategy

There is a need to refer to the config file for `vind` actions, which is in YAML format.

There is a lookup sequence while looking for such a configuration:
1. Explicitly specified by `--config` or `-c` parameter while running the command.
2. Explicitly exported system variable namely `VIND_CONFIG`. For example, `export VIND_CONFIG=/path/to/file.yaml` parameter while running the command.
3. Current folder's `vind.yaml`, if any.


## Usage

```sh
$ vind -h
A tool to create containers that look and work like virtual machines, on Docker.

Usage:
  vind [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Manage cluster configuration
  cp          Copy files or folders between a machine and the host file system
  create      Create a cluster
  delete      Delete a cluster
  help        Help about any command
  show        Show all running machines or some specific machine(s) by the given machine name(s).
  ssh         SSH into a machine
  start       Start all cluster machines or specific machine(s) by given name(s)
  stop        stop all cluster machines or specific machine(s) by given name(s)
  version     Print vind version

Flags:
  -c, --config string   Cluster configuration file
  -h, --help            help for vind

Use "vind [command] --help" for more information about a command
```

### config

`vind` reads a description of the **`Cluster`** to create and manage its **`Machines`** from a YAML file, `vind.yaml` by default.

An alternate name can be specified on the command line with the `--config` or `-c` option, or through the `VIND_CONFIG` environment variable.

The `config` command helps with creating the initial config file:

```sh
$ vind config create --replicas 3
INFO[0000] Creating config file vind.yaml

$ cat vind.yaml
cluster:
  name: cluster
  privateKey: cluster-key
machineSets:
- name: test
  replicas: 3
  spec:
    backend: docker
    image: brightzheng100/vind-ubuntu:22.04
    name: node%d
    portMappings:
    - containerPort: 22
```

You may try `vind config create -h` to see what can be configured through the command, or simply update the YAML file manually if you want to further customize it.

### create

Create the `vind` cluster:

```sh
$ vind create
INFO[0000] Pulling image: brightzheng100/vind-ubuntu:22.04 ...
INFO[0005] Creating machine: cluster-test-node0 ...
INFO[0005] Starting machine test-node0...
INFO[0006] Creating machine: cluster-test-node1 ...
INFO[0006] Starting machine test-node1...
INFO[0006] Creating machine: cluster-test-node2 ...
INFO[0006] Starting machine test-node2...
```

At first time, it may take 1 minute or so to pull the Docker image and then create the machines.
The creation of the machines typically takes just a few seconds.

> Note: since we've created the `vind.yaml` by `vind config create --replicas 3` in above step, we need not to specify it in this step's command. The same applies to the rest of commands.

### show

You may use `show` command to display the cluster details.

```sh
$ vind show
```

Output:

```
CONTAINER NAME       MACHINE NAME   PORTS       IP           IMAGE                              CMD          STATE     BACKEND
cluster-test-node0   test-node0     35827->22   10.88.0.21   brightzheng100/vind-ubuntu:22.04   /sbin/init   Running   docker
cluster-test-node1   test-node1     34929->22   10.88.0.22   brightzheng100/vind-ubuntu:22.04   /sbin/init   Running   docker
cluster-test-node2   test-node2     34237->22   10.88.0.23   brightzheng100/vind-ubuntu:22.04   /sbin/init   Running   docker
```

Actually there are just some Docker containers.

Here, let's understand a bit on the naming, by given `cluster-test-node0` in our case: **{CLUSTER_NAME}**-**{MACHINE_SET}**-**{MACHINE_NAME_WITH_INDEX}**.
- `cluster` is really the Cluster name, which can be any sensible name specified in YAML file's `cluster.name`.
- `test` is the MachineSet's name.
- `node{n}` is the Machine's name with index. Typically, we need to specify the machine with a desired index pattern, like `node%d`, or `node-%d`.

Please note that there are some useful output formats, which can be specified by `--output` or `-o` parameter:

- `table`: the default tab-based table-like format, as shown above.
- `json`: the JSON format.
- `ansible`: the Ansible inventory format. Once exported as say `inventory.yaml`, you can play with `vind` Machines like `ansible -i inventory.yaml -m ping all`.
- `ssh`: the SSH config format. Once exported as say `ssh.config`, you can play with regular SSH command like `ssh -F ssh.config vind-node0`.


### ssh

SSH into a machine with `ssh [[USER@]<MACHINE_NAME>]`, where the `<MACHINE_NAME>` is the combination of MachineSet's name and Machine's name.

```sh
$ vind ssh test-node0
root@test-node0:~# ps fx
    PID TTY      STAT   TIME COMMAND
      1 ?        Ss     0:00 /sbin/init
     15 ?        Ss     0:00 /lib/systemd/systemd-journald
     30 ?        Ss     0:00 /lib/systemd/systemd-logind
     32 ?        Ss     0:00 sshd: /usr/sbin/sshd -D [listener] 0 of 10-100 startups
     63 ?        Ss     0:00  \_ sshd: root@pts/1
     81 pts/1    Ss     0:00      \_ -bash
     87 pts/1    R+     0:00          \_ ps fx
     66 ?        Ss     0:00 /lib/systemd/systemd --user
     67 ?        S      0:00  \_ (sd-pam)
```

> Note: 
> 1. The machine user name can be other user, instead of `root`, if that's prepared in the Docker image and is specified in the YAML file.
> 2. The `[[USER@]<MACHINE_NAME>]` is optional: when no machine is specified, it will automatically pick the first machine.

### stop

You can stop one, or some specific machines, or all if nothing is specified.

To stop `test-node1`:

```sh
$ vind stop test-node1
INFO[0000] Stopping machine: cluster-test-node1 ...
```

Or stop all machines in the cluster -- it will detect whether the machine is in `stopped` state:

```sh
$ vind stop
INFO[0000] Stopping machine: cluster-test-node0 ...
INFO[0000] Machine cluster-test-node1 is already stopped...
INFO[0000] Stopping machine: cluster-test-node2 ...
```

### start

You can start one, or some specific machines, or all if nothing is specified.

To start `test-node1`:

```sh
$ vind start test-node1
INFO[0000] Starting machine: test-node1 ...
```

Or start all machines in the cluster -- it will detect whether the machine is in `started` state:

```sh
$ vind start
INFO[0000] Starting machine: test-node0 ...
INFO[0000] Machine test-node1 is already started...
INFO[0000] Starting machine: test-node2 ...
```

### cp

Copying files / folders between the host and machine can be useful.

- Copy a file from the machine to host:

```sh
$ vind cp test-node1:/etc/resolv.conf .
$ ls -l resolv.conf
-rw-r--r--  1 brightzheng  staff  43 Jan  7 17:54 resolv.conf
```

- Copy a file from the host to the machine:

```sh
$ vind cp README.md test-node1:/root/

$ vind ssh test-node1
root@test-node1:~# ls -l
total 4
-rw-r--r--. 1 501 dialout 107 Jan  5 05:07 README.md
```

### delete

Once the VM job is done, the machines can be easily deleted too.

```sh
$ vind delete
INFO[0000] Machine test-node0 is started, stopping and deleting machine...
INFO[0000] Machine test-node1 is started, stopping and deleting machine...
INFO[0001] Machine test-node2 is started, stopping and deleting machine...

$ docker ps
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
```

## Images

I've created a series of Docker images, covering Ubuntu, CentOS, Debian, Fedora, Amazon Linux, by inheriting from original `footloose`'s legacy with necessary enhancements (e.g. multi-arch build). Each of which will act like the VM by following some industrial practices.

You may refer to [images/README.md](./images/README.md) for what have been prepared and how to customize.

## Use Cases

In the [demo/README.md](./demo/README.md), I've shared some interesting experiments as use cases that you may explore, on top of the `vind` fundamental capabilities.

For example:
- [General demo](./demo/README.md#general-demo): The automated demo to showcase the basic usage of `vind`.
- [Demo: Docker in `vind`](./demo/README.md#demo-docker-in-vind): About how to run Docker in `vind`'s Machine.
- [Demo: Kubernetes in `vind`](./demo/README.md#demo-kubernetes-in-vind): About how to build multi-node Kubernetes cluster from scratch with `vind`'s Machines.
- [Demo: Ansible](./demo/README.md#demo-ansible): About how to play with Ansible with `vind`'s Machines.
- And more to come -- don't forget to let me know if you've got some more interesting use cases, and PRs are always welcome.

## How About `podman`?

Under the hood, `vind` orchestrates the `docker` commands while having some logic on top.

Since `podman` is `docker` compatible, at least in the commands that `vind` uses, `podman` is also supported.

To make it work, what we need to do is to create a softlink from `docker` to `podman`.

For example, in my Mac:

```sh
$ which podman
/opt/homebrew/bin/podman

$ ln -s `which podman` /usr/local/bin/docker
$ ls -al /usr/local/bin/docker
lrwxr-xr-x ... /usr/local/bin/docker -> /opt/homebrew/bin/podman

$ which docker
/usr/local/bin/docker
```

That's it, and `vind` will be working friendly with `podman` as it will treat `podman` as Docker.


## Helpful Tips

### Run Docker into `vind` Machine?

Docker in Docker container is tricky but as promised, it's totally possible in `vind`.

What we need to do is to enable `privileged: true` like [./demo/docker-in-vind.yaml](./demo/docker-in-vind.yaml).

Then you're good to go to install Docker like you do in normal Linux, by following official doc [here](https://docs.docker.com/engine/install/ubuntu/).


### Auto Bind Mount Host

Even `vind` offers `cp` command to streamline the folders / files sync up between host and machines, it would be great if the host file system is automatically bind mounted into the `vind` machines.

This is achievable by defining a special bind mount like this, which simply says that the root file system, which is `/`, is bind mounted to `vind` machine's `/host`:

```yaml
    volumes:
    - type: bind
      source: /
      destination: /host
```

You may refer to [./demo/ubunt-2.yaml](./demo/ubuntu-2.yaml) for the usage.

Once you've done so, after `vind ssh`, the command will try to automatically redirect to where you're in the current host folder. For example:

```sh
$ pwd
/Users/brightzheng/development/go/projects/vind/demo

$ ls
README.md       cluster-key     cluster-key.pub demo.cast       demo.sh         ubuntu-1.yaml   ubuntu-2.yaml

$ vind create -c ubuntu-2.yaml

$ vind ssh normal-node0 -c ubuntu-2.yaml
INFO[0000] SSH into machine [normal-node0] with user [root]
INFO[0000] Trying to cd into: /host/Users/brightzheng/development/go/projects/vind/demo
root@normal-node0:/host/Users/brightzheng/development/go/projects/vind/demo# ls
README.md  cluster-key  cluster-key.pub  demo.cast  demo.sh  ubuntu-1.yaml  ubuntu-2.yaml
```

## Contributions

Your issues, PRs, feedback, and whatever makes sense to making `vind` better is always welcome!
