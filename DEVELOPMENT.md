# Developing Nexus App

The purpose of this document is to provide instructions, examples, and tips & tricks to use when working on this codebase

# Pre-requisites

Before starting development on this project, ensure you have installed the following tools

* [NPM](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm) for ui development
* [Golang](https://go.dev/doc/install) for api development
* [Docker](https://docs.docker.com/engine/install/)


# Running the app using docker

To run both the ui and api service

```bash
make up
```

To view the logs of the ui and api service

```bash
make logs
```

To stop the ui and api

```bash
make stop
```


To rebuild the api and restart the app (useful for when you want to test out changes to the api code)

```
make reset
```

To restart the api (useful when you have changed the configuration for the api or ui)

```bash
make restart
```

> Note: The UI will automatically reload whenever you make a change to its code so doesn't get restarted using the above command



