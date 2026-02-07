FROM ubuntu
EXPOSE 5900
RUN apt update -y
RUN apt install -y tightvncserver 
RUN mkdir /root/.vnc
RUN sh -c 'echo vncx11 | vncpasswd -f > /root/.vnc/passwd'
RUN chmod 600 /root/.vnc/passwd
ADD init /init
ENV GEOMETRY=1280x1024
ENV XDMCP=-broadcast
ENTRYPOINT ["/init"]
