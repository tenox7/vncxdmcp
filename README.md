# XDMCP Xterminal via VNC as Docker Container

https://hub.docker.com/r/tenox7/vncxdmcp

XDMCP query/indirect/broadcast in VNC. Kind of like Xnest or Xephyr but with VNC backend.

You can use any VNC client, on MacOS you can just `open vnc://127.0.0.1`.

## Running

```sh
docker run -d \
    -p 5900:5900 \
    -p 6000:6000 \
    -e GEOMETRY=1920x1200 \
    -e DEPTH=8 \
    -e XDMCP_TARGET=192.168.1.197 \
    -e DOCKER_HOST_IP=192.168.1.10 \
    tenox7/vncxdmcp
```

VNC Password is `vncx11`.

- `XDMCP_TARGET` — remote host to connect via XDMCP Query. Leave it unset to be asked for the host in an input box.
- `DOCKER_HOST_IP` — Docker host's real LAN IP, reachable from `XDMCP_TARGET`. This is needed for traversing Docker's NAT.
- `FONT_SERVER` — remote X font server for CDE/DT fonts, e.g. `tcp/192.168.1.112:7000`. Autodetected on the target's port 7100/7000; set it empty to disable.

## Host prompt

With `DOCKER_HOST_IP` set and `XDMCP_TARGET` left out, a blank input box asks for the host to connect to, and asks again if what you typed does not answer XDMCP. Set `XDMCP_TARGET` to skip the prompt.
