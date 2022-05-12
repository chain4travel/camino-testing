set -euo pipefail

read -p "This script will remove unused, untagged Docker images and optionally all your Docker containers (ENTER to continue, Ctrl-C to quit) "
read -p "Stop & clear *ALL* your Docker containers before cleaning old images? (Y/n) " yn
case $yn in
    [Yy]* ) do_clear_containers=true;;
    * ) do_clear_containers=false;;
esac

if ${do_clear_containers}; then
    docker rm $(docker stop $(docker ps -a -q --format="{{.ID}}"))
else
    echo "Skipped clearing Docker containers"
fi

docker image rm $(docker images --quiet --filter "dangling=true")
