
# Mikrocount (atajsic fork)

## Introduction

Mikrocount is a tool written in go that pulls data from Mikrotik's accounting service, parses it, and stores it into influxdb.

## Getting Started
### Mikrotik router

Your Mikrotik router will need to have accounting with web access enabled.

[Mikrotik Manual:IP/Accounting](https://wiki.mikrotik.com/wiki/Manual:IP/Accounting)

Example:
```mikrotik
/ip accounting
set enabled=yes threshold=2000
/ip accounting web-access
set accessible-via-web=yes
```

Notes:

* `threshold`: 	maximum number of IP pairs in the accounting table (maximal value is 8192).
* The accounting feature of Mikrotik doesn't appear to work properly at all for packets that are [fasttracked](https://wiki.mikrotik.com/wiki/Manual:IP/Fasttrack).  If you want accurate results from this tool, disable fasttrack.

### Usage

#### docker

```
docker create \
  --name=mikrocount
  -e INFLUX_URL=http://influxdb:8086 \
  -e INFLUX_USER=username \ #remove if no user
  -e INFLUX_PWD=password \ #remove if no password
  -e LOCAL_CIDR=192.168.0.0/16 \
  -e MIKROTIK_ADDR=192.168.0.1 \
  -e MIKROCOUNT_TIMER=15 \
  --restart unless-stopped \
  atajsic/mikrocount
```

#### docker-compose
```
version: "3"
services:
  influxdb:
    image: influxdb
    volumes:
      - influxdb:/var/lib/influxdb
  mikrocount:
    image: atajsic/mikrocount
    depends_on:
      - influxdb
    environment:
      - INFLUX_URL=http://influxdb:8086
      - INFLUX_USER=username
      - INFLUX_PWD=password
      - LOCAL_CIDR=192.168.0.0/16
      - MIKROTIK_ADDR=192.168.0.1
      - MIKROCOUNT_TIMER=15
    restart: unless-stopped
```

## License

This software is distributed under the MIT License.  See the LICENSE file for more details.
