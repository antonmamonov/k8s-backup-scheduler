#!/bin/bash

IMAGENAME=antonm/k8s-pv-backup-scheduler:v0.0.1
CONTAINERNAME=backup-scheduler

docker rm -f $CONTAINERNAME

docker run -t -d \
    --name $CONTAINERNAME \
    -v $PWD:/app \
    -v $HOME/.kube/config:/root/.kube/config \
    $IMAGENAME

docker exec -it $CONTAINERNAME /bin/bash