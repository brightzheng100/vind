#!/bin/bash

repo_namespace=${REPO_NAMESPACE:-brightzheng100}
image_action=${IMAGE_ACTION:-push}  # push or load


# == Ubuntu ==

for distro in "25.04" "24.10" "24.04" "22.04" "20.04" "18.04" "16.04"; do

pushd ubuntu

    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.${distro}.non-root --${image_action} -t ${repo_namespace}/vind-ubuntu:${distro} .
    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.${distro}.root --${image_action} -t ${repo_namespace}/vind-ubuntu-root:${distro} .

popd

done


# == Amazon Linux ==

pushd amazonlinux

    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.2 --${image_action} -t ${repo_namespace}/vind-amazonlinux:2 .

popd


# == CentOS ==

pushd centos

    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.7 --${image_action} -t ${repo_namespace}/vind-centos:7 .
    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.8 --${image_action} -t ${repo_namespace}/vind-centos:8 .

popd


# == Debian ==

pushd debian

    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.buster --${image_action} -t ${repo_namespace}/vind-debian:buster .
    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.bullseye --${image_action} -t ${repo_namespace}/vind-debian:bullseye .
    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.bookworm --${image_action} -t ${repo_namespace}/vind-debian:bookworm .

popd


# == Fedora ==

pushd fedora

    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.40 --${image_action} -t ${repo_namespace}/vind-fedora:40 .
    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.41 --${image_action} -t ${repo_namespace}/vind-fedora:41 .
    docker buildx build --platform linux/amd64,linux/arm64 --file Dockerfile.42 --${image_action} -t ${repo_namespace}/vind-fedora:42 .

popd
