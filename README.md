# Simple Kubernetes Persistent Volume Backup with Scheduling

<img src='./images/servers.png' width="420" height="420">

## Prerequisite

Given this needs access to your Kubernetes cluster to backup a volume it would require an authenticated context. The running docker container assumes a working Kubernetes config file at `$HOME/.kube/config` to do work.

## Quick Backup Volume Job

```bash
# setup docker container
./runDockerProd.sh

# inside the running docker container
main help

main backup --help
```

## Backup Scheduling

For scheduled backups, just swap out the Kubernetes `Job` specification with a `CronJob`

## Quick Local Development

```bash
# build the docker image locally
./buildDockerImage.sh

# Run the built docker image as a local container
./runDockerDev.sh

# Run the main file inside the container with the 'help' argument
go run ./main.go help

# Use '--help' for a command to get command specific flags
go run main.go backup --help
```