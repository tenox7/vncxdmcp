// xdmcp-relay sits between Xtightvnc and a remote XDMCP server (X display
// manager) so the container can be reached over XDMCP despite Docker NAT.
//
// XDMCP's REQUEST packet embeds the querying host's own IP address, which
// the remote server dials back to start the actual X11 session (like FTP's
// PORT command). Behind Docker NAT that embedded address is the container's
// private IP, unreachable from the remote host, so the callback fails. This
// relay rewrites any of the container's own local IPv4 addresses found in
// outbound REQUEST packets to DOCKER_HOST_IP, a LAN-routable address whose
// TCP callback port (6000+display) Docker publishes back into the
// container. Every other XDMCP packet type is forwarded unmodified.
package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"os"
	"sync/atomic"
)

const (
	xdmcpPort = "177"
	opcodeReq = 7
)

func main() {
	target := os.Getenv("XDMCP_TARGET")
	advertise := os.Getenv("DOCKER_HOST_IP")
	if target == "" || advertise == "" {
		log.Fatal("XDMCP_TARGET and DOCKER_HOST_IP must be set")
	}
	newIP := net.ParseIP(advertise).To4()
	if newIP == nil {
		log.Fatalf("invalid DOCKER_HOST_IP %q", advertise)
	}
	from := localIPv4s()
	log.Printf("xdmcp-relay: 127.0.0.1:%s -> %s:%s, rewriting %v -> %s", xdmcpPort, target, xdmcpPort, from, advertise)

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 177})
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer conn.Close()

	upstream, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.ParseIP(target), Port: 177})
	if err != nil {
		log.Fatalf("dial target: %v", err)
	}
	defer upstream.Close()

	var clientAddr atomic.Pointer[net.UDPAddr]

	go func() {
		buf := make([]byte, 65535)
		for {
			n, err := upstream.Read(buf)
			if err != nil {
				log.Fatalf("upstream read: %v", err)
			}
			if a := clientAddr.Load(); a != nil {
				conn.WriteToUDP(buf[:n], a)
			}
		}
	}()

	buf := make([]byte, 65535)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("client read: %v", err)
			continue
		}
		clientAddr.Store(addr)
		pkt := rewrite(buf[:n], from, newIP)
		if _, err := upstream.Write(pkt); err != nil {
			log.Printf("upstream write: %v", err)
		}
	}
}

// rewrite replaces any occurrence of the container's local IPv4 addresses
// with newIP inside a REQUEST packet's body. Other opcodes pass through
// unchanged. In-place substitution keeps the packet length constant, so no
// re-framing of the surrounding XDMCP structure is needed.
func rewrite(pkt []byte, from []net.IP, newIP net.IP) []byte {
	if len(pkt) < 6 || binary.BigEndian.Uint16(pkt[2:4]) != opcodeReq {
		return pkt
	}
	body := pkt[6:]
	for i := 0; i+4 <= len(body); i++ {
		for _, ip := range from {
			if bytes.Equal(body[i:i+4], ip) {
				copy(body[i:i+4], newIP)
			}
		}
	}
	return pkt
}

func localIPv4s() []net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("interface addrs: %v", err)
	}
	var ips []net.IP
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}
		if ip4 := ipnet.IP.To4(); ip4 != nil {
			ips = append(ips, ip4)
		}
	}
	return ips
}
