# XDMCP Xterminal via VNC as Docker Container

XDMCP query/indirect/broadcast in VNC. Kind of like Xnest or Xephyr but with VNC backend.
 
You can use tightvnc client, on MacOS you can just `open vnc://127.0.0.1`.

## Running

```sh
docker run -d \
    -p 5900:5900 \
    -e GEOMETRY=1920x1200 \
    -e XDMCP="-query 172.17.0.1" \
    tenox7/vncxterminal
```

If XDMCP is unspecified `-broadcast` is the default.
