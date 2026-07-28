FROM golang:1.24-bookworm AS gobuilder
WORKDIR /build
COPY relay/ .
RUN CGO_ENABLED=0 go build -o xdmcp-relay .

FROM ubuntu
EXPOSE 5900
EXPOSE 6000
RUN apt update -y && apt install -y \
	tightvncserver \
	xfonts-base \
	xfonts-75dpi \
	xfonts-100dpi \
	xfonts-scalable \
	xfonts-terminus \
	xfonts-terminus-oblique \
	xfonts-jmk \
	xfonts-x3270-misc \
	xfonts-nexus \
	gsfonts-x11 \
	&& rm -rf /var/lib/apt/lists/*
COPY dec/ /usr/share/fonts/X11/dec/
RUN mkdir /root/.vnc && \
	sh -c 'echo vncx11 | vncpasswd -f > /root/.vnc/passwd' && \
	chmod 600 /root/.vnc/passwd
ADD init /init
COPY --from=gobuilder /build/xdmcp-relay /xdmcp-relay
ENV GEOMETRY=1280x1024
ENV XDMCP=-broadcast
ENV DEPTH=8
ENTRYPOINT ["/init"]
