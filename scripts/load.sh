#! /bin/bash
curl --unix-socket /tmp/firecracker.sock -i \
    -X PUT 'http://localhost/snapshot/load' \
    -H  'Accept: application/json' \
    -H  'Content-Type: application/json' \
    -d '{
            "snapshot_path": "/home/bluequbit/Dev/faas/snapshot.json",
            "mem_backend": {
                "backend_path": "/home/bluequbit/Dev/faas/snapshot.mem",
                "backend_type": "File"
            },
            "enable_diff_snapshots": false,
            "resume_vm": true
    }'

curl --unix-socket /tmp/firecracker.sock -i \
    -X PATCH 'http://localhost/vm' \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{
            "state": "Resumed"
    }'