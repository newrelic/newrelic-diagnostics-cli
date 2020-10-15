#!/bin/sh

echo "Cleaning up old images"
docker rmi $(docker images -q -f dangling=true)