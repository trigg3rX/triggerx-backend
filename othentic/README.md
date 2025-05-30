# `othentic/` Folder

## Files

* `.env.example`: With values related to Othentic CLI only.
* `Dockerfile`: Simple Docker with Node v22.6.0, ready to run latest version of Othentic CLI.
* `docker-compose.yaml`:
  * Uses values from `othentic/.env`
  * Runs 2 services, 'aggregator' and 'nexus'.
  * Will save the logs to a permanent storage, and aggregator peerstore is permanent while peerstore for nexus is temp.
  * Uses docker network `othentic_p2p`.
