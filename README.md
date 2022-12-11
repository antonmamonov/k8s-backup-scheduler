# Simple Kubernetes Persistent Volume Backup with Scheduling

<img src='./images/servers.png' width="420" height="420">

## Quick Local Development

```bash
# build the docker image locally
./buildDockerImage.sh

# Run the built docker image as a local container
./runDockerDev.sh

# Run the main file inside the container
go run ./main.go
```