#!/bin/bash

# build_daemon_rootfs.sh
# Script to build a rootfs with the daemon binary inside 
# and configure it to launch automatically at boot

set -xe

# Fixed configuration for daemon
DAEMON_PATH="/home/bluequbit/Dev/faas/cmd/daemon/daemon"
ROOTFS_FILE="$(pwd)/rootfs.ext4"
MOUNT_PATH="/tmp/daemon-rootfs"
SERVICE_FILE="app-service.sh"

# Check if daemon binary exists
if [ ! -f "$DAEMON_PATH" ]; then
    echo "Error: Daemon binary not found at $DAEMON_PATH"
    echo "Please build the daemon first using 'go build' in the cmd/daemon directory"
    exit 1
fi

# Create the OpenRC service file for the daemon
cat > $SERVICE_FILE << 'EOF'
#!/sbin/openrc-run

name="FaaS Daemon Service"
description="Runs the FaaS daemon on boot"
command="/usr/local/bin/daemon"
command_background=true
pidfile="/var/run/${RC_SVCNAME}.pid"
output_log="/var/log/${RC_SVCNAME}.log"
error_log="/var/log/${RC_SVCNAME}.err"

start_pre() {
    # Create directory for PID file
    mkdir -p /var/run
    # Create directory for logs
    mkdir -p /var/log
}

depend() {
    after net
}
EOF

chmod +x $SERVICE_FILE

# Create a directory for environment variables
mkdir -p /tmp/daemon-env
cat > /tmp/daemon-env/.env << 'EOF'
VM_IP=172.16.0.1
VM_PORT=8081
EOF

# Create the setup script for Alpine
cat > setup-alpine.sh << 'EOF'
#!/bin/sh

set -xe

# Install basic utilities and OpenRC
apk add --no-cache openrc
apk add --no-cache util-linux
apk add --no-cache gcc libc-dev curl file
apk add --no-cache python3 py3-pip

# Configure serial console
ln -s agetty /etc/init.d/agetty.ttyS0
echo ttyS0 >/etc/securetty
rc-update add agetty.ttyS0 default

# Set root password
echo "root:root" | chpasswd

# Install and configure SSH server for password-less access
apk add --no-cache openssh
rc-update add sshd default

# Configure SSH to allow key-based authentication without password
mkdir -p /etc/ssh
cat > /etc/ssh/sshd_config << 'SSHEOF'
PermitRootLogin yes
PubkeyAuthentication yes
PasswordAuthentication no
ChallengeResponseAuthentication no
UsePAM yes
PrintMotd no
AcceptEnv LANG LC_*
Subsystem sftp /usr/lib/ssh/sftp-server
SSHEOF

# Create SSH directory structure for root
mkdir -p /root/.ssh
chmod 700 /root/.ssh
touch /root/.ssh/authorized_keys

# Add specific public key
cat > /root/.ssh/authorized_keys << 'KEYEOF'
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC3mWQb0h7zXCcIeti0CrmZ9NnX4oxgv/CJ321uqWNeCzCYVdhyD5Qk2soPfYlS1eJ8JaUAJrWLXxdU5tv0K+2X40xzM39TKv/kHWvbWmQh9yykjAXqHoCFOZw9XsIUCzCD6MSZwRNExTPVLtkcrdyHN60r+m1ZWHFwhAn1PDqjo7xw2j1TdW8g+y1sq/1fKxf88bTF1+UxvPDJTGUksPfM3CMGTT9yA6LpaHJOSkD03iVuKiW805QAoXqVlY1TSyr5XF2Dl1iROjTH+GKaL+cMhJxMpNRfr0wMeZfAbEWCDq4bAaGwTmvps68BPNzwHA7kA3UetBqcHiRhvDjS5+em8Z37d3ckqbDw7H2YmY0i9DPwZrScMSj8YQWGYg1C/XP+L9+2eq/R89HgO9RyPa9EpoNj5zFLpEHTyqN8uNN/RTK7W9yM1Iq9viOqCkeCkpBrKtneJvoQ8GnP7cjhVwypdOC9RedpYA5AcigB3IpTvtRiTN44/Izej93ULQfRNPl96+ssd3haGue0bAmlx1y9g8kblAni4LTyXo8fIlHg334hG7BvWxnu3JLla1srUvaCRqbdRdQ/nY+ENm11kHBByTlQ6zAza8eJpiizaRK+WCSA/BQnjVEzVA3zALSSnfkhRUOBkyrnyIHpltxAALKOMQsfBM05TGDh+N1MhCuJVQ==
KEYEOF

chmod 600 /root/.ssh/authorized_keys

# Set DNS
echo "nameserver 1.1.1.1" >>/etc/resolv.conf

# Create directories for the daemon
mkdir -p /my-rootfs/app
mkdir -p /my-rootfs/var/run
mkdir -p /my-rootfs/var/log

# Copy environment variables
cp -a /env/.env /my-rootfs/app/.env

# Add basic boot services
rc-update add devfs boot
rc-update add procfs boot
rc-update add sysfs boot

# Add the daemon service to boot
rc-update add daemon boot

# Copy the root filesystem to the mounted directory
for d in bin etc lib root sbin usr; do tar c "/$d" | tar x -C /my-rootfs; done
for dir in dev proc run sys var tmp; do mkdir -p /my-rootfs/${dir}; done

# Set appropriate permissions
chmod 1777 /my-rootfs/tmp
EOF

chmod +x setup-alpine.sh

# Create the rootfs
dd if=/dev/zero of=$ROOTFS_FILE bs=1M count=1000
mkfs.ext4 $ROOTFS_FILE
mkdir -p $MOUNT_PATH
mount $ROOTFS_FILE $MOUNT_PATH

# Run Alpine in Docker to setup the rootfs
docker run -i --rm \
    -v $MOUNT_PATH:/my-rootfs \
    -v "$DAEMON_PATH:/usr/local/bin/daemon" \
    -v "$(pwd)/$SERVICE_FILE:/etc/init.d/daemon" \
    -v "/tmp/daemon-env:/env" \
    alpine sh < setup-alpine.sh

# Unmount the rootfs
umount $MOUNT_PATH

echo "Rootfs creation completed successfully!"
echo "Rootfs file is available at: $ROOTFS_FILE"

# Clean up temporary files
rm -f setup-alpine.sh $SERVICE_FILE
rm -rf /tmp/daemon-env
rmdir $MOUNT_PATH 2>/dev/null || true 