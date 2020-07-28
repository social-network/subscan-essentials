# Subscan Essentials

![License: GPL](https://img.shields.io/badge/license-GPL-blue.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/social-network/subscan)](https://goreportcard.com/report/github.com/social-network/subscan)
![subscan](https://github.com/social-network/subscan/workflows/subscan/badge.svg)

Subscan Essentials is a high-precision blockchain explorer scaffold project. It supports substrate-based blockchain networks with developer-friendly interface, standard or custom module parsing capabilities. It's developed by the Subscan team and used in subscan.io.  Developers are free to use the codebase to extend functionalities and develop unique user experiences for their audiences.


## API doc

The default API Doc can be found here [DOC](/docs/index.md)


### Feature

1. Separation of API Server and daemon
2. Support Substrate network custom type registration [Custom](/custom_type.md)
3. Support index block, Extrinsic, Event, log
4. More data can be indexed by custom plugins [Plugins](/plugins)
5. [Gen](https://github.com/social-network/subscan-plugin/tree/master/tools) tool can automatically generate plugin templates
6. Built-in default HTTP API [DOC](/docs/index.md)

### Install

```bash
./build.sh build &&  ./cmd/subscan --conf configs install
```

### RUN

> API 

```bash

./cmd/subscan --conf configs

```

> Daemon

```bash
./cmd/subscan --conf configs start substrate
./cmd/subscan --conf configs stop substrate
```


### Docker

```bash

docker-compose build

docker-compose up -d

```

## LICENSE

GPL-3.0


## Resource
 
[ITERING] https://github.com/social-network

[SUBSCAN] https://subscan.io/

[Darwinia] https://github.com/darwinia-network/darwinia

[freehere107] https://github.com/freehere107
