set -eu

# [ -e hello-vmlinux.bin ] || wget https://s3.amazonaws.com/spec.ccfc.min/img/hello/kernel/hello-vmlinux.bin
# [ -e hello-rootfs.ext4 ] || wget -O hello-rootfs.ext4 https://raw.githubusercontent.com/firecracker-microvm/firecracker-demo/ec271b1e5ffc55bd0bf0632d5260e96ed54b5c0c/xenial.rootfs.ext4
# [ -e hello-id_rsa ] || wget -O hello-id_rsa https://raw.githubusercontent.com/firecracker-microvm/firecracker-demo/ec271b1e5ffc55bd0bf0632d5260e96ed54b5c0c/xenial.rootfs.id_rsa

# TAP_DEV="fc-88-tap0"


# # set up the kernel boot args
# MASK_LONG="255.255.255.252"
# MASK_SHORT="/30"
# FC_IP="169.254.0.21"
# TAP_IP="169.254.0.22"
# # FC_MAC="02:FC:00:00:00:05"

# KERNEL_BOOT_ARGS="ro console=ttyS0 noapic reboot=k panic=1 pci=off nomodules random.trust_cpu=on"
# KERNEL_BOOT_ARGS="${KERNEL_BOOT_ARGS} ip=${FC_IP}::${TAP_IP}:${MASK_LONG}::eth0:off"

# ip link del "$TAP_DEV" 2> /dev/null || true
# ip tuntap add dev "$TAP_DEV" mode tap
# sysctl -w net.ipv4.conf.${TAP_DEV}.proxy_arp=1 > /dev/null
# sysctl -w net.ipv6.conf.${TAP_DEV}.disable_ipv6=1 > /dev/null
# ip addr add "${TAP_IP}${MASK_SHORT}" dev "$TAP_DEV"
# ip link set dev "$TAP_DEV" up

./firecracker  --no-api --config-file config.json