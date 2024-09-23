# Specify base image for building service binary
FROM golang:1

# Install go debugger for easier debug life
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# create and set default directory for service  files
RUN mkdir /app
WORKDIR /app

# optimize build time by installing dependencies
# before building so if source code changed but not
# the list of dependencies they don't have to be re-downloaded
COPY go.mod go.sum ./

# download service golang dependencies source code
RUN go mod download

# copy over local sources used to build service
COPY logging logging
COPY main.go main.go
COPY service service
COPY clients clients

# build service from latest sources
# with all compiler optimizations off to support debugging
RUN go install  -gcflags=all="-N -l"

# by default when a container is started from this image
# map port 8080 from the host to the container and run the
# ingest service
EXPOSE 8080
CMD ["nexus-api"]
