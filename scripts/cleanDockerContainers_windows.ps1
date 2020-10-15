echo "Cleaning up old images"

$images = docker images -q --filter "dangling=true"
docker rmi $images