to build the rootfs for the daemon, run the following command:

```bash
cd scripts
./build_daemon_rootfs.sh
```

to build the daemon binary, run the following command in the root directory:

```bash
cd cmd/daemon
go build -o daemon
```

without optimisation


    HTTP
    http_req_duration.......................................................: avg=121.19ms min=146.13µs med=2.76ms max=2.01s p(90)=554.25ms p(95)=566.19ms
      { expected_response:true }............................................: avg=121.19ms min=146.13µs med=2.76ms max=2.01s p(90)=554.25ms p(95)=566.19ms
    http_req_failed.........................................................: 0.00%  0 out of 6624
    http_reqs...............................................................: 6624   27.24294/s

    EXECUTION
    iteration_duration......................................................: avg=1.61s    min=1.55s    med=1.58s  max=3.12s p(90)=1.63s    p(95)=1.67s   
    iterations..............................................................: 1316   5.412396/s
    vus.....................................................................: 10     min=1         max=10
    vus_max.................................................................: 10     min=10        max=10

    NETWORK
    data_received...........................................................: 62 MB  256 kB/s
    data_sent...............................................................: 1.4 MB 5.8 kB/s




running (4m03.2s), 00/10 VUs, 1316 complete and 10 interrupted iterations
default ✗ [=============================>--------] 08/10 VUs  4m03.1s/5m00.0s


with optimisation



    ✓ invocation has request ID
    ✓ list VMs successful
    ✓ VMs response is array

    HTTP
    http_req_duration.......................................................: avg=127.43ms min=148.68µs med=3.17ms max=3.28s p(90)=555.37ms p(95)=570.9ms
      { expected_response:true }............................................: avg=127.43ms min=148.68µs med=3.17ms max=3.28s p(90)=555.37ms p(95)=570.9ms
    http_req_failed.........................................................: 0.00%  0 out of 5300
    http_reqs...............................................................: 5300   26.01268/s

    EXECUTION
    iteration_duration......................................................: avg=1.64s    min=1.08s    med=1.58s  max=4.3s  p(90)=1.67s    p(95)=1.77s  
    iterations..............................................................: 1052   5.163272/s
    vus.....................................................................: 10     min=1         max=10
    vus_max.................................................................: 10     min=10        max=10

    NETWORK
    data_received...........................................................: 51 MB  251 kB/s
    data_sent...............................................................: 1.1 MB 5.6 kB/s

