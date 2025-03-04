#!/bin/sh
# curl --unix-socket /tmp/firecracker.sock -i \
#     -X PUT 'http://localhost/boot-source'   \
#     -H 'Accept: application/json'           \
#     -H 'Content-Type: application/json'     \
#     -d '{
#         "kernel_image_path": "/home/bluequbit/Dev/faas/firecracker/resources/x86_64/vmlinux-6.1.129",
#         "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
#     }'
# curl --unix-socket /tmp/firecracker.sock -i \
#     -X PUT 'http://localhost/drives/rootfs' \
#     -H 'Accept: application/json'           \
#     -H 'Content-Type: application/json'     \
#     -d '{
#         "drive_id": "rootfs",
#         "path_on_host": "/home/bluequbit/Dev/faas/hello-rootfs.ext4",
#         "is_root_device": true,
#         "is_read_only": false
#     }'
# curl -fsS --unix-socket /tmp/firecracker.sock -i \
#     -X PUT "http://localhost/logger" \
#     -H  "accept: application/json" \
#     -H  "Content-Type: application/json" \
#     -d '{
#         "level": "Trace",
#         "module": "api_server::request"
#     }'
# curl --unix-socket /tmp/firecracker.sock -i \
#     -X PUT 'http://localhost/actions'       \
#     -H  'Accept: application/json'          \
#     -H  'Content-Type: application/json'    \
#     -d '{
#         "action_type": "InstanceStart"
#      }'
# curl --unix-socket /tmp/firecracker.sock -i \
#     -X PATCH 'http://localhost/vm' \
#     -H 'Accept: application/json' \
#     -H 'Content-Type: application/json' \
#     -d '{
#         "state": "Paused"
#     }'
# curl --unix-socket /tmp/firecracker.sock -i \
#     -X PUT 'http://localhost/snapshot/create' \
#     -H 'Accept: application/json' \
#     -H 'Content-Type: application/json' \
#     -d '{
#         "mem_file_path": "/home/bluequbit/Dev/faas/snapshot.mem",
#         "snapshot_path": "/home/bluequbit/Dev/faas/snapshot.json",
#         "snapshot_type": "Full"
#     }'

# curl --unix-socket /tmp/firecracker.sock -i \
#     -X PUT 'http://localhost/snapshot/load' \
#     -H  'Accept: application/json' \
#     -H  'Content-Type: application/json' \
#     -d '{
#             "snapshot_path": "/home/bluequbit/Dev/faas/snapshot.json",
#             "mem_backend": {
#                 "backend_path": "/home/bluequbit/Dev/faas/snapshot.mem",
#                 "backend_type": "File"
#             },
#             "enable_diff_snapshots": false,
#             "resume_vm": false
#     }'

# curl --unix-socket /tmp/firecracker.sock -i \
#     -X PATCH 'http://localhost/vm' \
#     -H 'Accept: application/json' \
#     -H 'Content-Type: application/json' \
#     -d '{
#             "state": "Resumed"
#     }'