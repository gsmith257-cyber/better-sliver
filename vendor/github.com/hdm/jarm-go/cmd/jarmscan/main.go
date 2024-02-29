package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	log "github.com/sirupsen/logrus"

	"github.com/hdm/jarm-go"
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
func Fingerprint(t target, och chan result) {

	results := []string{}
	for _, probe := range jarm.GetProbes(t.Host, t.Port) {
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
			return
		}

		data := jarm.BuildProbe(probe)
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

		ans, err := jarm.ParseServerHello(buff, probe)
		if err != nil {
			results = append(results, "")
			continue
		}

		results = append(results, ans)
	}

	och <- result{
		Target: t,
		Hash:   jarm.RawHashToFuzzyHash(strings.Join(results, ",")),
	}
}

var DefualtBackoff = func(r, m int) time.Duration {
	return time.Second
}

type target struct {
	Host string
	Port int

	Retries int
	Backoff func(r, m int) time.Duration
}

type result struct {
	Target target
	Hash   string
	Error  error
}

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatalf("usage: ./jarm -p <ports> [host] <host:8443> <https://host:port> <host,port>...")
	}

	if *workerCount < 1 {
		log.Fatalf("invalid worker count: %d", *workerCount)
	}

	if *quietMode {
		dn, _ := os.Create(os.DevNull)
		log.SetOutput(dn)
	}

	defaultPorts, err := CrackPortsWithDefaults(*defaultPorts, []uint16{})
	if err != nil {
		log.Fatalf("invalid ports: %s", err)
	}

	tch := make(chan target, 1)
	och := make(chan result, 1)

	wgo := sync.WaitGroup{}
	wgt := sync.WaitGroup{}

	for x := 0; x <= *workerCount; x++ {
		wgt.Add(1)
		go func() {
			defer wgt.Done()
			for t := range tch {
				Fingerprint(t, och)
			}
		}()
	}

	// Output consolidator
	wgo.Add(1)
	go func() {
		defer wgo.Done()
		for o := range och {
			if o.Error != nil {
				log.Printf("failed to scan %s:%d: %s", o.Target.Host, o.Target.Port, o.Error)
				continue
			}
			if len(o.Target.Host) > 24 {
				fmt.Printf("JARM\t%s:%d\t%s\n", o.Target.Host, o.Target.Port, o.Hash)
			} else {
				fmt.Printf("JARM\t%24s:%d\t%s\n", o.Target.Host, o.Target.Port, o.Hash)
			}

		}
	}()

	// Process targets
	for _, s := range flag.Args() {

		t := target{}

		// Try parsing as a URL first
		if u, err := url.Parse(s); err == nil {
			t.Host = u.Hostname()
			port, _ := strconv.Atoi(u.Port())
			t.Port = port
		}

		// Next try parsing as an address:port pair
		if t.Host == "" {
			host, portStr, _ := net.SplitHostPort(s)
			port, _ := strconv.Atoi(portStr)
			t.Host = host
			t.Port = port
		}

		// Next try parsing as a host,port pair
		if t.Host == "" {
			bits := strings.SplitN(s, ",", 2)
			if len(bits) == 2 && bits[0] != "" {
				t.Host = bits[0]
				port, _ := strconv.Atoi(bits[1])
				t.Port = port

			}
		}

		// Finally try parsing as a host:port pair
		if t.Host == "" {
			bits := strings.SplitN(s, ":", 2)
			if bits[0] != "" {
				t.Host = bits[0]
			}
			if len(bits) == 2 && bits[0] != "" {
				port, _ := strconv.Atoi(bits[1])
				t.Port = port
			}
		}

		hosts := []string{t.Host}

		for _, host := range hosts {

			// Support CIDR networks as targets
			hch := make(chan string, 1)
			qch := make(chan int, 1)
			hwg := sync.WaitGroup{}
			hwg.Add(1)

			go func() {
				defer hwg.Done()
				for thost := range hch {
					ports := defaultPorts
					if t.Port != 0 {
						ports = []int{t.Port}
					}
					for _, port := range ports {
						tch <- target{
							Host: thost,
							Port: port,

							Retries: *retries,
						}
					}
				}
			}()

			// Try to iterate the host as a CIDR range
			herr := jarm.AddressesFromCIDR(host, hch, qch)

			// Not a parseable range, handle it as a bare host instead
			if herr != nil {
				hch <- host
			}

			// Wrap up and wait
			close(hch)
			hwg.Wait()
			close(qch)
		}
	}

	// Wait for scans to complete
	close(tch)
	wgt.Wait()

	// Wait for output to finish
	close(och)
	wgo.Wait()
}
