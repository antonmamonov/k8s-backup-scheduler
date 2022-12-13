#!/bin/bash

IMAGENAME=antonm/k8s-pv-backup-scheduler
CONTAINERNAME=backup-scheduler

docker rm -f $CONTAINERNAME

docker run -t -d \
    --name $CONTAINERNAME \
    --entrypoint="/bin/bash" \
    -v $HOME/.kube/config:/root/.kube/config \
    $IMAGENAME

docker exec -it $CONTAINERNAME /bin/bash