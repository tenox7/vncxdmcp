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
//
// The upstream host is read from $targetFile rather than from the environment,
// so the X server can come up first and be pointed at a display manager later
// by whatever the user types into the prompt. "-probe <host>" instead just asks
// one host whether it is willing to manage a display, which is what the
// prompt uses to reject a typo.
package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const (
	opcodeQuery     = 2
	opcodeWilling   = 5
	opcodeUnwilling = 6
	opcodeReq       = 7
	xdmcpPort       = 177
	targetFile      = "/tmp/xdmcp-target"
	pollInterval    = time.Second
)

func main() {
	if len(os.Args) == 3 && os.Args[1] == "-probe" {
		os.Exit(probe(os.Args[2]))
	}
	advertise := os.Getenv("DOCKER_HOST_IP")
	if advertise == "" {
		log.Fatal("DOCKER_HOST_IP must be set")
	}
	newIP := net.ParseIP(advertise).To4()
	if newIP == nil {
		log.Fatalf("invalid DOCKER_HOST_IP %q", advertise)
	}
	from := localIPv4s()
	log.Printf("xdmcp-relay: 127.0.0.1:177 -> %s, rewriting %v -> %s", targetFile, from, advertise)

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: xdmcpPort})
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer conn.Close()

	// unconnected, so the target can change without reopening the socket
	upstream, err := net.ListenUDP("udp4", nil)
	if err != nil {
		log.Fatalf("upstream socket: %v", err)
	}
	defer upstream.Close()

	var target atomic.Pointer[net.UDPAddr]
	var clientAddr atomic.Pointer[net.UDPAddr]
	var lastQuery atomic.Pointer[[]byte]

	setTarget := func(host string) {
		addr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(host, "177"))
		if err != nil {
			log.Printf("resolve %s: %v", host, err)
			return
		}
		if old := target.Load(); old != nil && old.String() == addr.String() {
			return
		}
		target.Store(addr)
		log.Printf("target %s", addr)
		// replay the X server's last QUERY so the login screen comes up at
		// once, instead of waiting out its ~32s retransmission interval
		if q := lastQuery.Load(); q != nil {
			upstream.WriteToUDP(*q, addr)
		}
	}

	go func() {
		for {
			if b, err := os.ReadFile(targetFile); err == nil {
				if h := strings.TrimSpace(string(b)); h != "" {
					setTarget(h)
				}
			}
			time.Sleep(pollInterval)
		}
	}()

	go func() {
		buf := make([]byte, 65535)
		for {
			n, addr, err := upstream.ReadFromUDP(buf)
			if err != nil {
				log.Printf("upstream read: %v", err)
				time.Sleep(time.Second)
				continue
			}
			t, c := target.Load(), clientAddr.Load()
			if t == nil || c == nil || !addr.IP.Equal(t.IP) {
				continue
			}
			conn.WriteToUDP(buf[:n], c)
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
		if opcode(pkt) == opcodeQuery {
			q := append([]byte(nil), pkt...)
			lastQuery.Store(&q)
		}
		t := target.Load()
		if t == nil {
			continue // no host chosen yet, the X server will retry
		}
		if _, err := upstream.WriteToUDP(pkt, t); err != nil {
			log.Printf("upstream write: %v", err)
		}
	}
}

func opcode(pkt []byte) int {
	if len(pkt) < 6 {
		return -1
	}
	return int(binary.BigEndian.Uint16(pkt[2:4]))
}

// rewrite replaces any occurrence of the container's local IPv4 addresses
// with newIP inside a REQUEST packet's body. Other opcodes pass through
// unchanged. In-place substitution keeps the packet length constant, so no
// re-framing of the surrounding XDMCP structure is needed.
func rewrite(pkt []byte, from []net.IP, newIP net.IP) []byte {
	if opcode(pkt) != opcodeReq {
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
