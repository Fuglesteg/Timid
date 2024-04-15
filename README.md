> The shy container
# Description
Timid is a UDP proxy that tracks connections and stops and starts docker containers.
Developed to be used with game servers that need to save resources.
Timid will start the containers it controls as soon as a connection attempt is made to the specified port, and will shut down the containers
if no connections exist for a customizable amount of idle time.

## Motivation
The use case for Timid was running game servers that don't need to be instantly accessible,
but should still be available.

I run a small server where I sometimes host game servers for friends, since the server
doesn't have many resources and I don't want to manually control what servers are running, as that
loses the advantage of having a dedicated server, I found a need for a tool like Timid.
Idle game servers can sometimes use an unexpected amount of resources and that is why 
the use case for Timid emerged.

The concept of disabling unused servers is not new, and this project was mainly inspired by 
[Infrared](https://github.com/haveachin/infrared), which has all the same capabilites as Timid (and more) for minecraft servers (which are TCP based).
The code for the UDP proxy is mainly based on this [gist](https://gist.github.com/mike-zhang/3853251) by [mike-zhang](https://github.com/mike-zhang).

There also exists a project that is very similar to Timid and might fit your usecase better, [Lazytainer](https://github.com/vmorganp/Lazytainer) 
is an older project than timid and has the ability listen on multiple ports which Timid currently does not.

# Installation
Timid is available as a docker image, this is the recommended way of running Timid.

## Container image
Container image is available at https://hub.docker.com/r/fuglesteg/timid.

Remember that the docker capabilites of Timid require access to the docker daemon,
this is achieved by mounting the docker.sock file to the container, see the example 
[compose.yml](#docker-compose) file.

## Requirements
To run the application without docker you require golang and git.

Download the repo and run ```go install``` then ```go run .```.

# Usage
## Docker compose
Timid is recommended to be used with docker compose, here is an example compose.yml file
using valheim:
```yaml
version: '3.4'

name: valheim # Set project name
services:
  valheim:
    image: lloesche/valheim-server
    restart: always
    stop_grace_period: 2m
    volumes:
      - "./server/config:/config"
      - "./server/data:/opt/valheim"
    labels:
      - timid.group.valheim
  timid:
    image: fuglesteg/timid
    restart: always
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
    ports:
      - "2456-2458:2456-2458/udp"
    environment:
      TIMID_PORT: 2456
      TIMID_TARGET_ADDRESS: valheim:2456
      TIMID_CONTAINER_GROUP: valheim
```

Note that because the containers are running on the same docker network they are able to communicate.
The docker local DNS resolves the hostname of valheim to the container.
This docker compose setup will start both containers and if no connections are made the valheim container will stop.
The Timid container will listen to connections and start the valheim container up again if a connection is made.
You can also start only the Timid container itself without starting the other container(s), the important thing to remember
in that case is that the container has to be built in order for Timid to start it.

## Configuration

|Environment variable| Purpose | Type | Default Value |
|---|---|---|---|
|TIMID_PORT| Port on host computer the program should listen to|Integer| Unset & required |
|TIMID_TARGET_ADDRESS| Address to reroute traffic to, can be name of docker container running on the same network| String\|/URL| Unset & required |
|<s>TIMID_CONTAINER_NAME</s>| DEPRECATED as of 1.2 (use TIMID_GROUP_NAME)    <s>Name of container running service, the container that will be shutdown and started based on number of connections</s>| String| Unset |
|TIMID_GROUP_NAME| Name of container group, which is used to look up label of containers | String | Unset & required |
|TIMID_CONTAINER_SHUTDOWN_DELAY| Time until the proxy shuts down the container after no connections exist| <a href="#duration-string">Duration string</a>| 1 minute |
|TIMID_PAUSE_CONTAINER| Timid will pause the container instead of pausing it | Boolean| false |
|TIMID_LOG_VERBOSITY| How verbose should the logs be| Integer, Range 1-6| 1 |
|TIMID_CONNECTION_TIMEOUT_DELAY| UDP has no concept of a connection, so this tracks how long a connection must be unused for it to be considered disconnected| <a href="#duration-string">Duration string</a> | 1 minute |

### [Duration string](https://pkg.go.dev/time#ParseDuration)
"A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"."

\- [go documentation](https://pkg.go.dev/time#ParseDuration)

# TODO
- [x] Implement option of pausing container instead of stopping it
- [x] Implement option of pausing container when inactive, but then stopping it if it is paused for a certain duration
- [x] Implement using labels to control multiple containers
- [ ] Listen on multiple ports
- [ ] Support TCP?
- [ ] REST API?
