package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bdoner/ddump/crt"
	"github.com/bdoner/ddump/dns"
)

type config struct {
	domain string
}

var conf config

func main() {
	parseFlags()

	fmt.Printf("Looking for subdomains for:\n%s\n", conf.domain)

	err := crt.OpenDatabaseConnection()
	if err != nil {
		log.Fatalf("%v", err)
	}

	subdomains, err := crt.QueryDomain(conf.domain)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if err := crt.Dispose(); err != nil {
		log.Fatalf("%v", err)
	}

	resolved := make(chan dns.ResolveResult, len(*subdomains))
	resolver := dns.GetDefaultResolver()

	for _, v := range *subdomains {
		go resolver.Resolve(v, resolved)
	}

	for i := 0; i < len(*subdomains); i++ {
		res := <-resolved

		fmt.Printf("%s\n", res.Domain)
		if 0 < len(res.Records) {
			for _, v := range res.Records {
				if v.Error != nil {
					//fmt.Printf("Error: %v\n", v.Error)
					continue
				}

				if v.Type == "" {
					continue
				}

				fmt.Printf("  %6s: %s\n", v.Type, v.Answer)
			}
		}
	}

}

func parseFlags() {
	flag.StringVar(&conf.domain, "d", "", "The domain to dump data for.")
	flag.Parse()

	if flag.NFlag() < 1 {
		flag.Usage()
		os.Exit(1)
	}
}
