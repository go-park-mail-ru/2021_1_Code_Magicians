docker kill $(docker ps -q)
docker system prune -a -f
docker image prune -f