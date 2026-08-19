package main

import (
	"log"
	"net"
	"time"
)

// query is a complete XDMCP QUERY packet with an empty authentication-name
// list, byte for byte what Xtightvnc emits with -query.
var query = []byte{0, 1, 0, opcodeQuery, 0, 1, 0}

const probeWait = 3 * time.Second

// probe reports whether a display manager on host is willing to manage a
// display, so a typo in the prompt gets a second try instead of a black
// screen. Returns a process exit status.
func probe(host string) int {
	conn, err := net.Dial("udp4", net.JoinHostPort(host, "177"))
	if err != nil {
		log.Print(err)
		return 1
	}
	defer conn.Close()
	if _, err := conn.Write(query); err != nil {
		log.Print(err)
		return 1
	}
	conn.SetReadDeadline(time.Now().Add(probeWait))
	buf := make([]byte, 65535)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("%s: no answer", host)
			return 1
		}
		switch {
		case opcode(buf[:n]) == opcodeUnwilling:
			log.Printf("%s: unwilling to manage a display", host)
			return 1
		case opcode(buf[:n]) == opcodeWilling:
			name, status, ok := parseWilling(buf[:n])
			if !ok {
				continue
			}
			log.Printf("%s: %s, %s", host, name, status)
			return 0
		}
	}
}

// parseWilling splits a WILLING packet's body into its three ARRAY8 fields
// and returns the second and third, the responder's hostname and status.
func parseWilling(pkt []byte) (host, status string, ok bool) {
	if opcode(pkt) != opcodeWilling {
		return "", "", false
	}
	b := pkt[6:]
	var f [3]string
	for i := range f {
		if len(b) < 2 {
			return "", "", false
		}
		n := int(b[0])<<8 | int(b[1])
		if len(b) < 2+n {
			return "", "", false
		}
		f[i], b = string(b[2:2+n]), b[2+n:]
	}
	return f[1], f[2], true
}
