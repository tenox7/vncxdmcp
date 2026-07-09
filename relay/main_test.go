package main

import (
	"bytes"
	"net"
	"testing"
)

// capturedRequest is a real XDMCP REQUEST packet captured from Xtightvnc
// querying a live display manager from inside a bridge-mode container
// (docker0 address 172.17.0.3), confirming the wire format assumptions.
var capturedRequest = []byte{
	0x00, 0x01, 0x00, 0x07, 0x00, 0x3c, 0x00, 0x01,
	0x01, 0x00, 0x00, 0x01, 0x00, 0x04, 0xac, 0x11,
	0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00,
	0x12, 0x4d, 0x49, 0x54, 0x2d, 0x4d, 0x41, 0x47,
	0x49, 0x43, 0x2d, 0x43, 0x4f, 0x4f, 0x4b, 0x49,
	0x45, 0x2d, 0x31, 0x00, 0x13, 0x58, 0x44, 0x4d,
	0x2d, 0x41, 0x55, 0x54, 0x48, 0x4f, 0x52, 0x49,
	0x5a, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x2d, 0x31,
	0x00, 0x00,
}

func TestRewriteReplacesEmbeddedAddress(t *testing.T) {
	pkt := append([]byte(nil), capturedRequest...)
	from := []net.IP{net.IPv4(172, 17, 0, 3).To4()}
	to := net.IPv4(192, 168, 1, 10).To4()

	got := rewrite(pkt, from, to)

	if bytes.Contains(got, from[0]) {
		t.Fatalf("original address 172.17.0.3 still present after rewrite")
	}
	want := append([]byte(nil), capturedRequest...)
	copy(want[14:18], to) // address bytes at body offset 8 = packet offset 14
	if !bytes.Equal(got, want) {
		t.Fatalf("rewrite produced unexpected packet\n got: % x\nwant: % x", got, want)
	}
	if len(got) != len(capturedRequest) {
		t.Fatalf("length changed: got %d want %d", len(got), len(capturedRequest))
	}
}

func TestRewriteLeavesNonRequestUntouched(t *testing.T) {
	willing := []byte{0x00, 0x01, 0x00, 0x05, 0x00, 0x02, 0xac, 0x11}
	from := []net.IP{net.IPv4(172, 17, 0, 3).To4()}
	to := net.IPv4(192, 168, 1, 10).To4()

	got := rewrite(append([]byte(nil), willing...), from, to)
	if !bytes.Equal(got, willing) {
		t.Fatalf("non-REQUEST packet was modified: got % x want % x", got, willing)
	}
}

func TestRewriteNoMatchLeavesPacketUntouched(t *testing.T) {
	pkt := append([]byte(nil), capturedRequest...)
	from := []net.IP{net.IPv4(10, 0, 0, 99).To4()}
	to := net.IPv4(192, 168, 1, 10).To4()

	got := rewrite(pkt, from, to)
	if !bytes.Equal(got, capturedRequest) {
		t.Fatalf("packet changed despite no matching address")
	}
}
