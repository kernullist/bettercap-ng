package modules

import (
	"fmt"

	"github.com/evilsocket/bettercap-ng/core"
	"regexp"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// poor man's TLS Client Hello with SNI extension parser :P
var sniRe = regexp.MustCompile("\x00\x00.{4}\x00.{2}([a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,6})\x00")

func sniParser(ip *layers.IPv4, pkt gopacket.Packet, tcp *layers.TCP) bool {
	data := tcp.Payload
	dataSize := len(data)

	if dataSize < 2 || data[0] != 0x16 || data[1] != 0x03 {
		return false
	}

	m := sniRe.FindSubmatch(data)
	if len(m) < 2 {
		return false
	}

	domain := string(m[1])
	if tcp.DstPort != 443 {
		domain = fmt.Sprintf("%s:%d", domain, tcp.DstPort)
	}

	NewSnifferEvent(
		pkt.Metadata().Timestamp,
		"https",
		ip.SrcIP.String(),
		domain,
		SniffData{
			"Domain": domain,
		},
		"%s %s > %s",
		core.W(core.BG_YELLOW+core.FG_WHITE, "sni"),
		vIP(ip.SrcIP),
		core.Yellow("https://"+domain),
	).Push()

	return true
}
