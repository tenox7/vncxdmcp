docker rm -f vncxdmcp

docker run -d \
  --name vncxdmcp \
  -p 5900:5900 \
  -e GEOMETRY=1920x1200 \
  -e XDMCP="-query 172.17.0.1" \
  tenox7/vncxdmcp:latest
