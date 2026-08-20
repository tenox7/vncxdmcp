#!/bin/bash
docker rm -f vncxterminal

case "$(uname -s)" in
	Darwin) DOCKER_HOST_IP=$(ipconfig getifaddr "$(route get 1.1.1.1 | awk '/interface:/{print $2}')") ;;
	Linux)  DOCKER_HOST_IP=$(hostname -I | awk '{print $1}') ;;
	*) echo "unsupported OS: set DOCKER_HOST_IP manually" >&2; exit 1 ;;
esac

docker run -d \
  --name vncxterminal \
  -p 5900:5900 \
  -p 6000:6000 \
  -e GEOMETRY=1920x1200 \
  -e DOCKER_HOST_IP=${DOCKER_HOST_IP} \
  tenox7/vncxterminal:latest

open vnc://127.0.0.1
