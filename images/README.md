## Images

Here are some example images that I've built, which are out-of-the-box and available in my Docker Hub namespace.

### Prepare

Assume you have installed Docker, like `brew install docker` in Mac OS.

Now let's create and use a builder:

```sh
docker buildx create --use
```

### Build

This is how I built the images:

```sh
./build.sh
```

If you want to host the images in your namespace, do this:

```sh
export REPO_NAMESPACE=<YOUR REPO NAMESPACE, e.g. brightzheng100, or ghcr.io/YOURNAME>
./build.sh
```

Or, you can build you own always by using `docker buildx build`.
Refer to the [`build.sh](./build.sh) for how.


Please note that when using `podman` to build, the commands are actually quite different.

Let's take building `brightzheng100/vind-ubuntu:22.04` for example:

```sh
# 1. Build them with a manifest specified
podman build --platform linux/amd64,linux/arm64 --file Dockerfile.22.04.non-root --manifest brightzheng100/vind-ubuntu:ubuntu-manifest .
# 2. Push manifest with the targeted image tag
podman manifest push brightzheng100/vind-ubuntu:ubuntu-manifest brightzheng100/vind-ubuntu:22.04
```

### List

#### Ubuntu

- brightzheng100/vind-ubuntu:`version` -- with a "normal" `ubuntu` user builtin -- where the `version` can be:
  - 25.04
  - 24.10
  - 24.04
  - 22.04
  - 20.04
  - 18.04
- brightzheng100/vind-ubuntu-root:`version`, where the `version` can be:
  - 25.04
  - 24.10
  - 24.04
  - 22.04
  - 20.04
  - 18.04

#### Fedora

- brightzheng100/vind-fedora:`version`, where the `version` can be:
  - 42
  - 41
  - 40

#### Debian

- brightzheng100/vind-debian:`version`, where the `version` can be:
  - bookworm
  - bullseye
  - buster

#### CentOS

- brightzheng100/vind-centos:`version`, where the `version` can be:
  - 8
  - 7

#### Amazon Linux

- brightzheng100/vind-amazonlinux:`version`, where the `version` can be:
  - 2
