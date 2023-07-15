package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const colorEscape = "\x1b"
const (
	colorReset = iota
	colorBold
)
const (
	colorFgBlack = iota + 30
	colorFgRed
	colorFgGreen
	colorFgYellow
	colorFgBlue
	colorFgMagenta
	colorFgCyan
	colorFgWhite
)

func main() {
	flag.Parse()

	c, err := wgctrl.New()
	if err != nil {
		log.Fatalf("failed to open wgctrl: %v", err)
	}
	defer c.Close()

	var devices []*wgtypes.Device
	if device := flag.Arg(0); device != "" {
		d, err := c.Device(device)
		if err != nil {
			log.Fatalf("failed to get device %q: %v", device, err)
		}

		devices = append(devices, d)
	} else {
		devices, err = c.Devices()
		if err != nil {
			log.Fatalf("failed to get devices: %v", err)
		}
	}

	for _, d := range devices {
		printDevice(d)

		for _, p := range d.Peers {
			printPeer(p)
		}
	}
}

func printDevice(d *wgtypes.Device) {
	output :=
		greenBoldColor("interface") + ": " + greenColor(d.Name) + " (" + d.Type.String() + ")\n" +
			"  " + boldColor("public key") + ": " + d.PublicKey.String() + "\n" +
			"  " + boldColor("private key") + ": (hidden)\n" +
			"  " + boldColor("listening port") + ": " + strconv.Itoa(d.ListenPort) + "\n\n"

	fmt.Print(output)
}

func printPeer(p wgtypes.Peer) {
	output :=
		yellowBoldColor("peer") + ": " + yellowColor(p.PublicKey.String()) + "\n" +
			"  " + boldColor("endpoint") + ": " + p.Endpoint.String() + "\n" +
			"  " + boldColor("allowed ips") + ": " + strings.ReplaceAll(ipsString(p.AllowedIPs), "/", cyanColor("/")) + "\n"

	if p.LastHandshakeTime.Second() > 0 {
		output +=
			"  " + boldColor("latest handshake") + ": " + formatTime(p.LastHandshakeTime) + " ago" + "\n"
	}

	output +=
		"  " + boldColor("transfer") + ": " + formatBytes(p.ReceiveBytes) + " received, " + formatBytes(p.TransmitBytes) + " sent\n" +
			"  " + boldColor("persistent keepalive") + ": every " + formatTimeUnit(int(p.PersistentKeepaliveInterval.Seconds()), "second") + "\n\n"

	fmt.Print(output)
}

func ipsString(ipns []net.IPNet) string {
	ss := make([]string, 0, len(ipns))
	for _, ipn := range ipns {
		ss = append(ss, ipn.String())
	}

	return strings.Join(ss, ", ")
}

func formatTime(t time.Time) string {
	output := ""

	if t.Minute() > 0 {
		output += formatTimeUnit(t.Minute(), "minute") + " "
	}

	output += formatTimeUnit(t.Second(), "second")

	return output
}

func formatTimeUnit(value int, unit string) string {
	return strconv.Itoa(value) + " " + plural(value, unit)
}

func plural(value int, unit string) string {
	if value > 1 {
		unit += "s"
	}

	return cyanColor(unit)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d %s", b, cyanColor("B"))
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), cyanColor(string("KMGTPE"[exp])+"iB"))
}

func greenBoldColor(s string) string {
	return colorize(s, colorBold, colorFgGreen)
}

func greenColor(s string) string {
	return colorize(s, colorReset, colorFgGreen)
}

func boldColor(s string) string {
	return colorize(s, colorBold, 0)
}

func yellowBoldColor(s string) string {
	return colorize(s, colorBold, colorFgYellow)
}

func yellowColor(s string) string {
	return colorize(s, colorReset, colorFgYellow)
}

func cyanColor(s string) string {
	return colorize(s, colorReset, colorFgCyan)
}

func colorize(s string, style int, color int) string {
	format := strconv.Itoa(style)
	if color >= 30 {
		format = ";" + strconv.Itoa(color)
	}
	
	return fmt.Sprintf("%s[%sm%s%s[%dm", colorEscape, format, s, colorEscape, 0)
}
