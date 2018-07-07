# HFS

[![Build Status](https://travis-ci.org/jiajunhuang/hfs.svg?branch=master)](https://travis-ci.org/jiajunhuang/hfs)
[![codecov](https://codecov.io/gh/jiajunhuang/hfs/branch/master/graph/badge.svg)](https://codecov.io/gh/jiajunhuang/hfs)

Huang's Chunked Distributed File System.

My toy.

TODO:

- [ ] concurrently download file

## Make it works

0. install and start etcd:

```bash
$ # check https://coreos.com/etcd/docs/latest/dl_build.html
```

1. first, create a directory which stores chunks:

```bash
$ sudo mkdir -p /hfs/chunks/
$ sudo chown -R <your username>:<your group> /hfs
```

2. clone code:

```bash
$ git clone https://github.com/jiajunhuang/hfs.git
$ cd hfs
$ make
```

3. start chunk server:

```bash
$ ./bin/chunkserver
```

4. open another terminal, run the client:

```bash
$ ./bin/hfsclient upload ~/Downloads/ubuntu-16.04.4-server-amd64.iso
file created, uuid is 60aca0d4-28d9-481b-9a62-460f642664d0
$ ./bin/hfsclient download 60aca0d4-28d9-481b-9a62-460f642664d0
file with UUID 60aca0d4-28d9-481b-9a62-460f642664d0 download successful! origin file name is ubuntu-16.04.4-server-amd64.iso
$ md5sum 60aca0d4-28d9-481b-9a62-460f642664d0
6a7f31eb125a0b2908cf2333d7777c82  60aca0d4-28d9-481b-9a62-460f642664d0
$ md5sum ~/Downloads/ubuntu-16.04.4-server-amd64.iso
6a7f31eb125a0b2908cf2333d7777c82  /Users/neo.huang/Downloads/ubuntu-16.04.4-server-amd64.iso
```

5. check chunks:

```bash
$ ls /hfs/chunks/
06c520d5-9c01-438f-9760-7004decce303  28749da5-2ad7-4d01-8dff-3b4d5426b0bf  6cba4f9c-4121-4388-b1e9-5015d7aac9bb  d246ea5c-c66e-4ba6-a4aa-b4b3d46f0ef1
0949ed79-4b75-4b9d-aa24-799e725aeffb  39d242e5-4c63-4d0f-8145-629a61e1b4a2  9021960b-b283-4f21-bcbb-9fc425609e99  e97c274b-0467-4295-a488-f917dceb2f70
1dcb04a8-4f3e-4df4-a1bf-28b38de2ac7c  44884d3e-4017-42cd-b0e3-baf3bb0090f2  acb103bb-d140-4cd8-869f-a7ef0f14b1c0
2581cc6c-3ae1-4b8f-8f69-86290d9e2191  4c18bf25-d652-4ed4-ab01-cda48f87a5e6  c3983a62-b770-43df-81c5-c8f2684951ea
```

6. check KVs in etcd:

```bash
$ ETCDCTL_API=3 etcdctl get "" --prefix=true
/hfs/chunks/06c520d5-9c01-438f-9760-7004decce303
...(ignore the rest)
```

7. delete file:

```bash
$ ./bin/hfsclient delete 60aca0d4-28d9-481b-9a62-460f642664d0
```
