#!/bin/bash

set -xe

dd if=/dev/zero of=rootfs.ext4 bs=1M count=125
mkfs.ext4 rootfs.ext4
mkdir -p /tmp/my-rootfs
sudo mount rootfs.ext4 /tmp/my-rootfs

docker run -i --rm \
    -v /tmp/my-rootfs:/my-rootfs \
    alpine sh < setup.sh

sudo umount /tmp/my-rootfs

# rootfs available under `rootfs.ext4`