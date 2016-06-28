# snmp-send

Rather than relying on a central SNMP data collecting server such as Cacti,
have each server send the data to InfluxDB.

In order to keep InfluxDB off of the public internet in a cloud environment, 
a simple collector called [fuzz](https://github.com/gaker/fuzz) is used to 
provide rudimentary authorization, and to transform the data into InfluxDB 
[Line Protocol](https://docs.influxdata.com/influxdb/v0.12/)

## Prerequisites

* A [Go](https://golang.org/) build env on the target server environment.
* [GoDep](https://github.com/tools/godep)

## Build

```
$ mkdir -p $GOPATH/src/github.com/gaker && cd $GOPATH/src/github.com/gaker
$ git clone https://github.com/gaker/snmp-send.git && cd snmp-send
$ godep restore
$ go get github.com/gaker/snmp-send
```

And the binary will be in your ``$GOPATH/bin/`` directory.

## Installation

The binary can now be deployed to any server matching the architecture it was built on,
typically as a cron job at whatever interval suits your needs. For instance, as a 1 minute
cron job, the following could be set in the crontab.

```
* * * * * /usr/local/bin/snmp-send -conf=/etc/snmp-send.conf
```

## Configuration file

A config file matching [the following](https://github.com/gaker/snmp-send/blob/master/example.config.json)
should be included on each server the binary will be deployed to.


## Building with Docker

Build into a small alpine linux container

```
$ docker build -t snmp-send-builder -f Dockerfile.build .
$ docker run --rm snmp-send-builder | docker build -t gaker/snmp-send:latest -t gaker/snmp-send:0.1 -f Dockerfile.run -
```

