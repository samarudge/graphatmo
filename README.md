# Graphatmo

## Send data from [Netatmo](http://netatmo.com/) weather stations to Graphite

**Work in progress**

Usage;

Build with

    export GOPATH=./
    go build ./graphatmo.go

Create config file `~/.graphatmo.yml`

    auth:
      client_id: ""
      client_secret: ""
    graphite: "localhost:2003"

with API keys created at https://dev.netatmo.com/dev/createapp

run `graphatmo login` to authenticate

run `graphatmo` to start sending data