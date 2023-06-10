package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/urfave/cli/v2"
)

type options struct {
	upstreamNormal  []string
	upstreamSpecial []string
	upstreamDomains []string
	masquedDomain   string
	upstreamDomain  string
}

func getRandomUpstream(upstreams []string) (string, string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := r.Intn(len(upstreams))
	upstream := upstreams[index]
	parts := strings.Split(upstream, ":")
	if len(parts) != 2 {
		log.Fatalf("Invalid upstream: %s", upstream)
	}
	return parts[0], parts[1]
}

func HasSuffixInSlice(str string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(str, suffix) {
			return true
		}
	}
	return false
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg, opts options) {
	var upstreamHost, upstreamPort string
	var special bool
	defer func() {
		if err := w.Close(); err != nil {
			log.Println("Error closing connection:", err)
		}
	}()
	c := new(dns.Client)
	m := new(dns.Msg)
	// log.Println(r.Question[0].Name)
	if strings.HasSuffix(r.Question[0].Name, opts.masquedDomain) ||
		HasSuffixInSlice(r.Question[0].Name, opts.upstreamDomains) {
		m.SetQuestion(strings.Replace(r.Question[0].Name, opts.masquedDomain, opts.upstreamDomain, 1), r.Question[0].Qtype)
		upstreamHost, upstreamPort = getRandomUpstream(opts.upstreamSpecial)
		special = true
	} else {
		m.SetQuestion(r.Question[0].Name, r.Question[0].Qtype)
		upstreamHost, upstreamPort = getRandomUpstream(opts.upstreamNormal)
	}
	in, _, err := c.Exchange(m, net.JoinHostPort(upstreamHost, upstreamPort))
	if special && (in == nil || len(in.Answer) == 0) {
		log.Println(r.Question[0].String(), "not found retrying with normal upstream")
		m.SetQuestion(r.Question[0].Name, r.Question[0].Qtype)
		upstreamHost, upstreamPort = getRandomUpstream(opts.upstreamNormal)
		in, _, err = c.Exchange(m, net.JoinHostPort(upstreamHost, upstreamPort))
	}

	if err != nil {
		log.Println(err)
		return
	}

	res := new(dns.Msg)
	res.SetReply(r)
	if strings.HasSuffix(r.Question[0].Name, opts.masquedDomain) {
		for _, ans := range in.Answer {
			if ans.Header().Rrtype == dns.TypeA {
				if a, ok := ans.(*dns.A); ok {
					a.Hdr.Name = strings.Replace(a.Hdr.Name, opts.upstreamDomain, opts.masquedDomain, 1)
				}
			}
		}
	}
	res.Answer = in.Answer
	// log.Println(in.Answer)
	res.RecursionAvailable = true
	w.WriteMsg(res)
}

func appendPeriods(str string) string {
	if !strings.HasPrefix(str, ".") {
		str = "." + str
	}
	if !strings.HasSuffix(str, ".") {
		str = str + "."
	}
	return str
}

func main() {
	app := &cli.App{
		Name:  "DNS Proxy",
		Usage: "A DNS proxy server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				Aliases: []string{"l"},
				Value:   ":1053",
				Usage:   "Address and port to listen on",
			},
			&cli.StringFlag{
				Name:    "masqued-domain",
				Aliases: []string{"m"},
				Value:   ".example.com.",
				Usage:   "Domain to be masqued",
			},
			&cli.StringFlag{
				Name:    "upstream-domain",
				Aliases: []string{"u"},
				Value:   ".service.consul.",
				Usage:   "Upstream domain to be masqued",
			},
			&cli.StringSliceFlag{
				Name:    "upstream-domains",
				Aliases: []string{"uds"},
				Value:   cli.NewStringSlice("service.consul.", ".consul."),
				Usage:   "Upstream domains to be resolved",
			},
			&cli.StringSliceFlag{
				Name:    "upstream-special",
				Aliases: []string{"uc"},
				Value:   cli.NewStringSlice("localhost:8600"),
				Usage:   "Special upstream host:port",
			},
			&cli.StringSliceFlag{
				Name:    "upstream-normal",
				Aliases: []string{"un"},
				Value:   cli.NewStringSlice("1.1.1.1:53"),
				Usage:   "Normal resolution upstream host:port",
			},
		},
		Action: func(c *cli.Context) error {
			var opts options
			listenAddr := c.String("listen")
			opts.upstreamSpecial = c.StringSlice("upstream-special")
			opts.upstreamNormal = c.StringSlice("upstream-normal")
			opts.masquedDomain = appendPeriods(c.String("masqued-domain"))
			opts.upstreamDomain = appendPeriods(c.String("upstream-domain"))
			opts.upstreamDomains = c.StringSlice("upstream-domains")
			server := &dns.Server{Addr: listenAddr, Net: "udp"}
			dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
				handleDNSRequest(w, r, opts)
			})

			log.Printf("DNS proxy listening on %s\n", listenAddr)
			return server.ListenAndServe()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
