> The shy container
# Description
Timid is a UDP proxy that tracks connections and stops and starts docker containers.
Developed to be used with game servers that need to save resources.
The container will start as soon as a connection attempt is made to the server, and will shut down
if no connections exist for a customizable amount of time.
## Motivation
The use case for Timid was running game servers that don't need to be instantly accessible,
but should still be available.

I run a small server where I sometimes host game servers for friends, since the server
doesn't have many resources and I don't want to manually control whar servers are running, as that
loses the advantage of having a dedicated server.
Idle game servers can sometimes use an unexpected amount of resources and that is why 
the use case for Timid emerged.

The concept of disabling unused servers is not new, and this project was mainly inspired by 
[Infrared](https://github.com/haveachin/infrared).
The code for the UDP proxy is mainly based on this [gist](https://gist.github.com/mike-zhang/3853251) by [mike-zhang](https://github.com/mike-zhang) 

# Installation
Timid is available as a docker image, this is the recommended way of running timid.

## Container image
Container image is available at https://hub.docker.com/r/fuglesteg/timid
Remember that the docker capabilites of timid require access to the docker daemon
this is achieved by mounting the docker.sock file to the container, see the example 
[compose.yml](#docker compose) file.

## Requirements
To run the application without docker you require golang and git.

Download the repo and run ```go install``` then ```go run .```.

# Usage
## Docker compose
Timid is recommended to be used with docker compose, here is an example compose.yml file
using valheim:
```yaml
version: '3.4'

services:
  valheim:
    image: lloesche/valheim-server
    restart: always
    stop_grace_period: 2m
    cap_add:
      - sys_nice
    volumes:
      - "./server/config:/config"
      - "./server/data:/opt/valheim"
  timid:
    image: timid
    restart: always
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
    ports:
      - "2456-2458:2456-2458/udp"
    environment:
      PROXY_PORT: 2456
      PROXY_TARGET_ADDRESS: valheim:2456
      PROXY_CONTAINER_NAME: valheim-valheim-1 # dependent on the project being name valheim
```

Note that because the containers are running on the same docker network they are able to communicate.
The docker local DNS resolves the hostname of valheim to the container.
This docker compose setup will start both containers and if no connections are made the valheim container will stop.
The Timid container will listen to connections and start the valheim container up again if a connection is made.

## Configuration

|Environment variable| Purpose | Type | Default Value |
|---|---|---|---|
|PROXY_PORT| Port on host computer the program should listen to|Integer| Unset & required |
|PROXY_TARGET_ADDRESS| Address to reroute traffic to, can be name of docker container running on the same network| String\|/URL| Unset & required |
|PROXY_CONTAINER_NAME| Name of container running service, the container that will be shutdown and started based on number of connections| String| Unset & required |
|PROXY_CONTAINER_SHUTDOWN_DELAY| Time until the proxy shuts down the container after no connections exist| <a href="#duration string">Duration string</a>| 1 minute |
|<s>PROXY_PAUSE_CONTAINER</s>| Unimplemented, will make the proxy pause the container instead of stopping it| Boolean| false |
|PROXY_LOG_VERBOSITY| How verbose should the logs be| Integer, Range 1-6| 1 |
|PROXY_CONNECTION_TIMEOUT_DELAY| UDP has no concept of a connection, so this tracks how long a connection must be unused for it to be considered disconnected| <a href="#duration string">Duration string</a> | 1 minute |

### [ Duration string ](https://pkg.go.dev/time#ParseDuration)
"A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"."
\- go documentation

# TODO
- [ ] Implement option of pausing container instead of stopping it
- [ ] Implement option of pausing container when inactive, but then stopping it if it is paused for a certain duration
