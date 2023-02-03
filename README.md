# Description
Timid is a UDP proxy that tracks connections and stops and starts docker containers.
Developed to be used with game servers that need to save resources.
TODO: Based off gist

# Installation
Timid is available as a docker image, this is the recommended way of running timid

## Container image
Container image is available at <Insert link>
Remember that for the docker capabilites of timid require access to the docker daemon
this is achieved by mounting the docker.sock file to the container, see the example 
[compose.yml](#Docker compose) file

## Requirements
To run the application without docker you require golang and git
Download the repo and run ```sh go install``` then ```sh go run```

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
      PROXY_CONTAINER_NAME: valheim-valheim-1
```

Note that because the containers are running on the same network they are able to communicate.
The docker local DNS resolves the hostname of valheim to the container
## Docker CLI

## Configuration

|Environment variable| Purpose | Type |
|---|---|---|
|PROXY_PORT| Port on host computer the program should listen to|Integer|
|PROXY_TARGET_ADDRESS| Address to reroute traffic to, can be name of docker container running on the same network| String\|/URL|
|PROXY_CONTAINER_NAME| Name of container running service, the container that will be shutdown and started based on number of connections| String|
|PROXY_CONTAINER_SHUTDOWN_DELAY| Time until the proxy shuts down the container after no connections exist| [ Duration string ] (#Duration string)|
|~~~PROXY_PAUSE_CONTAINER~~~| Unimplemented, will make the proxy pause the container instead of stopping it| Boolean|
|PROXY_LOG_VERBOSITY| How verbose should the logs be| Integer, Range 1-6|
|PROXY_CONNECTION_TIMEOUT_DELAY| UDP has no concept of a connection, so this tracks how long a connection must be unused for it to considered disconnected| [ Duration string ] (#Duration string)|

### Duration string
[A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".](https://pkg.go.dev/time#ParseDuration)
