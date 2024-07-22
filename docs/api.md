The API should **NOT** be publicly exposed!

|Route|Purpose|Return value|
|---|---|---|
|GET /info| General info on the state of Timid | `{"connections": int, "containerGroup": {"name": string, "state": "Stopped" \| "Running" \| "Paused"}}` |
|POST /proxy/trigger| Trigger the proxy as if a connection was made |
|GET /proxy| Get general info on the proxy | `{"connections": int, "port": int, "targetAddress": string}` |
|GET /containers| Get a list of the containers in the container group | `[{"id": string, "name": "string", "state": "Stopped" \| "Running" \| "Paused"}]` |
|GET /containers/{containerId}| Get a certain container given an ID | `{"id": string, "name": "string", "state": "Stopped" \| "Running" \| "Paused"}` |
|POST /containers/start| Start all containers in group | null |
|POST /containers/stop| Stop all containers in group | null |
|POST /containers/pause| Pause all containers in group | null |
|POST /containers/restart| Restart all containers in group | null |
|POST /containers/{containerId}/start| Start a certain container given an ID | null |
|POST /containers/{containerId}/stop| Stop a certain container given an ID | null |
|POST /containers/{containerId}/pause| Pause a certain container given an ID | null |
|POST /containers/{containerId}/restart| Restart a certain container given an ID | null |
