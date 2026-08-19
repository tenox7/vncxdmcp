FROM ubuntu AS cbuilder
RUN apt update -y && apt install -y gcc libc6-dev libx11-dev
COPY xprompt.c /
RUN gcc -O2 -Wall -o /xprompt /xprompt.c -lX11

FROM golang:1.24-bookworm AS gobuilder
WORKDIR /build
COPY relay/ .
RUN CGO_ENABLED=0 go build -o xdmcp-relay .

FROM ubuntu
EXPOSE 5900
EXPOSE 6000
RUN apt update -y && apt install -y \
	tightvncserver \
	x11-utils \
	x11-xserver-utils \
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
COPY sco/ /usr/share/fonts/X11/sco/
RUN mkdir /root/.vnc && \
	sh -c 'echo vncx11 | vncpasswd -f > /root/.vnc/passwd' && \
	chmod 600 /root/.vnc/passwd
ADD init /init
COPY --from=gobuilder /build/xdmcp-relay /xdmcp-relay
COPY --from=cbuilder /xprompt /xprompt
ENV GEOMETRY=1280x1024
ENV XDMCP=-broadcast
ENV DEPTH=8
ENTRYPOINT ["/init"]
