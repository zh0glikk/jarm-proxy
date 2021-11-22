package jarm

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// Version is set by the goreleaser build
var Version = "dev"

var defaultPorts = flag.String("p", "443", "default ports")
var workerCount = flag.Int("w", 256, "worker count")
var quietMode = flag.Bool("q", false, "quiet mode")
var retries = flag.Int("r", 0, "number of times to retry dialing")

// ValidPort determines if a port number is valid
func ValidPort(pnum int) bool {
	if pnum < 1 || pnum > 65535 {
		return false
	}
	return true
}

// CrackPortsWithDefaults turns a comma-delimited port list into an array, handling defaults
func CrackPortsWithDefaults(pspec string, defaults []uint16) ([]int, error) {
	results := []int{}

	// Use a map to dedup and shuffle ports
	ports := make(map[int]bool)

	bits := strings.Split(pspec, ",")
	for _, bit := range bits {

		// Support the magic strings "default" and "defaults"
		if bit == "default" || bit == "defaults" {
			for _, pnum := range defaults {
				ports[int(pnum)] = true
			}
			continue
		}

		// Split based on dash
		prange := strings.Split(bit, "-")

		// Scan all ports if the specifier is a single dash
		if bit == "-" {
			prange = []string{"1", "65535"}
		}

		// No port range
		if len(prange) == 1 {
			pnum, err := strconv.Atoi(bit)
			if err != nil || !ValidPort(pnum) {
				return results, fmt.Errorf("invalid port %s", bit)
			}
			// Record the valid port
			ports[pnum] = true
			continue
		}

		if len(prange) != 2 {
			return results, fmt.Errorf("invalid port range %s (%d)", prange, len(prange))
		}

		pstart, err := strconv.Atoi(prange[0])
		if err != nil || !ValidPort(pstart) {
			return results, fmt.Errorf("invalid start port %d", pstart)
		}

		pstop, err := strconv.Atoi(prange[1])
		if err != nil || !ValidPort(pstop) {
			return results, fmt.Errorf("invalid stop port %d", pstop)
		}

		if pstart > pstop {
			return results, fmt.Errorf("invalid port range %d-%d", pstart, pstop)
		}

		for pnum := pstart; pnum <= pstop; pnum++ {
			ports[pnum] = true
		}
	}

	// Create the results from the map
	for port := range ports {
		results = append(results, port)
	}
	return results, nil
}

// Fingerprint probes a single host/port
func Fingerprint(t Target) *Result {

	results := []string{}
	for _, probe := range GetProbes(t.Host, t.Port) {
		dialer := proxy.FromEnvironmentUsing(&net.Dialer{Timeout: time.Second * 2})
		addr := net.JoinHostPort(t.Host, fmt.Sprintf("%d", t.Port))

		c := net.Conn(nil)
		n := 0

		for c == nil && n <= t.Retries {
			// Ignoring error since error message was already being dropped.
			// Also, if theres an error, c == nil.
			if c, _ = dialer.Dial("tcp", addr); c != nil || t.Retries == 0 {
				break
			}

			bo := t.Backoff
			if bo == nil {
				bo = DefualtBackoff
			}

			time.Sleep(bo(n, t.Retries))

			n++
		}

		if c == nil {
			return nil
		}

		data := BuildProbe(probe)
		c.SetWriteDeadline(time.Now().Add(time.Second * 5))
		_, err := c.Write(data)
		if err != nil {
			results = append(results, "")
			c.Close()
			continue
		}

		c.SetReadDeadline(time.Now().Add(time.Second * 5))
		buff := make([]byte, 1484)
		c.Read(buff)
		c.Close()

		ans, err := parseServerHello(buff, probe)
		if err != nil {
			results = append(results, "")
			continue
		}

		results = append(results, ans)
	}

	return &Result{
		Target: t,
		Hash:   RawHashToFuzzyHash(strings.Join(results, ",")),
	}
}

var DefualtBackoff = func(r, m int) time.Duration {
	return time.Second
}

type Target struct {
	Host string
	Port int

	Retries int
	Backoff func(r, m int) time.Duration
}

type Result struct {
	Target Target
	Hash   string
	Error  error
}
