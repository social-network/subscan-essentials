# Netscan

![License: GPL](https://img.shields.io/badge/license-GPL-blue.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/social-network/netscan)](https://goreportcard.com/report/github.com/social-network/netscan)
![subscan](https://github.com/social-network/netscan/workflows/subscan/badge.svg)

Netscan is a high-precision blockchain explorer data harvester. It supports substrate-based blockchain networks with developer-friendly interface, standard or custom module parsing capabilities. It's a fork of the work done by the subscan team to provide social network and postgres support.  Developers are free to use the codebase to extend functionalities and develop unique user experiences for their audiences.


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
./build.sh build &&  ./cmd/netscan --conf configs install
```

### RUN

> API

```bash

./cmd/netscan --conf configs

```

> Daemon

```bash
./cmd/netscan --conf configs start substrate
./cmd/netscan --conf configs stop substrate
```


### Docker

```bash

docker-compose build

docker-compose up -d

```

## LICENSE

GPL-3.0
