package dns_spoof

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/gobwas/glob"

	"github.com/evilsocket/islazy/str"
)

var hostsSplitter = regexp.MustCompile(`\s+`)

type HostEntry struct {
	Host    string
	Suffix  string
	Expr    glob.Glob
	Address net.IP
}

func (e HostEntry) Matches(host string) bool {
	return e.Host == host || strings.HasSuffix(host, e.Suffix) || (e.Expr != nil && e.Expr.Match(host))
}

type Hosts []HostEntry

func NewHostEntry(host string, address net.IP) HostEntry {
	entry := HostEntry{
		Host:    host,
		Address: address,
	}

	if host[0] == '.' {
		entry.Suffix = host
	} else {
		entry.Suffix = "." + host
	}

	if expr, err := glob.Compile(host); err == nil {
		entry.Expr = expr
	}

	return entry
}

func HostsFromFile(filename string) (err error, entries []HostEntry) {
	input, err := os.Open(filename)
	if err != nil {
		return
	}
	defer input.Close()

	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := str.Trim(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if parts := hostsSplitter.Split(line, 2); len(parts) == 2 {
			address := net.ParseIP(parts[0])
			domain := parts[1]
			entries = append(entries, NewHostEntry(domain, address))
		} else {
			return fmt.Errorf("'%s' invalid hosts line", line), nil
		}
	}

	return
}

func (h Hosts) Resolve(host string) net.IP {
	for _, entry := range h {
		if entry.Matches(host) {
			return entry.Address
		}
	}
	return nil
}
