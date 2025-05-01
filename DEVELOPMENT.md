# Developing Nexus App

The purpose of this document is to provide instructions, examples, and tips & tricks to use when developing Nexus

## Pre-requisites

Before starting development on this project, ensure you have installed the following tools

* [NPM](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm) for ui development
* [Golang](https://go.dev/doc/install) for api development
* [Docker](https://docs.docker.com/engine/install/)
* [psql](https://www.timescale.com/blog/how-to-install-psql-on-mac-ubuntu-debian-windows/)
## Running Nexus App Locally Using Docker

To run both the ui and api service

```bash
make up
```

To view the logs of the ui and api service

```bash
make logs
```

You can view the logs of just a single component by specifying the service name of the component as defined in the [docker-compose.yml file](./docker-compose.yml)

e.g. to view just the api logs

```bash
make logs S=nexus-api
```

To stop the ui and api

```bash
make stop
```

To rebuild the api and restart the app (useful for when you want to test out changes to the api code)

```bash
make refresh
```

To restart the api (useful when you have changed the configuration for the api or ui)

```bash
make restart
```

> Note: The UI will automatically reload whenever you make a change to its code so doesn't get restarted using the above command

To reset the database to zero state, rebuild the api, and restart the database api and ui (useful when trying to test things from a "clean" state)

```bash
make reset
```

## Debugging

### API Service

```bash
make debug-nexus-api
```

### Database

```bash
make debug-database
```
### Git flow

git stash- saving local changes

git pull - pulling local changes

git stash apply - re-adding local changes I had pushed away

make down to stop all the services

make reset refresh ready to restart datatabse stuff and wait for the database to be ready

make it p=TestE2E to run a just a single test

make test- runs all the tests