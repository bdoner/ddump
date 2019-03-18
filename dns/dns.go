package dns

import (
	"fmt"
	"net"
	"sync"
)

// ResolverSettings determins the behavior of Resolve(string)
type ResolverSettings struct {
	RecordTypes []string
}

// ResolveResult is a domain with all it's records
type ResolveResult struct {
	Domain  string
	Records []Record
}

// Record holds the response after resolving a DNS record
type Record struct {
	Type   string
	Answer string
	Error  error
}

// GetDefaultResolver returns a default resolver
func GetDefaultResolver() *ResolverSettings {
	return &ResolverSettings{
		RecordTypes: []string{"A", "AAAA", "TXT", "CNAME", "MX", "NS"},
	}
}

// Resolve resolves a fully qualified domain name's records into a Record struct
func (resolver *ResolverSettings) Resolve(fqdn string, responses *[]ResolveResult, wg *sync.WaitGroup) {

	defer wg.Done()

	result := ResolveResult{
		Domain: fqdn,
	}

	result.Records = make([]Record, 1)
	for _, rec := range resolver.RecordTypes {
		r := Record{
			Type:   rec,
			Answer: "",
		}

		switch rec {
		case "A", "AAAA":
			ips, err := net.LookupIP(fqdn)
			if err != nil {
				continue
			}

			for _, i := range ips {
				if c := i.To4(); c == nil {
					if r.Type != "AAAA" {
						continue
					}
				} else {
					if r.Type != "A" {
						continue
					}
				}
				r.Answer = i.String()
				result.Records = append(result.Records, r)
			}

		case "TXT":
			txts, err := net.LookupTXT(fqdn)
			if err != nil {
				continue
			}

			for _, i := range txts {
				r.Answer = i
				result.Records = append(result.Records, r)
			}

		case "CNAME":
			cname, err := net.LookupCNAME(fqdn)
			if err != nil {
				continue
			}

			r.Answer = cname
			result.Records = append(result.Records, r)

		case "MX":
			mxs, err := net.LookupMX(fqdn)
			if err != nil {
				continue
			}

			for _, i := range mxs {
				r.Answer = fmt.Sprintf("%d %s", i.Pref, i.Host)
				result.Records = append(result.Records, r)
			}

		case "NS":
			nss, err := net.LookupNS(fqdn)
			if err != nil {
				continue
			}

			for _, i := range nss {
				r.Answer = i.Host
				result.Records = append(result.Records, r)
			}

		// case "SRV":
		// 	nss, err := net.Look(fqdn)
		// 	if err != nil {
		// 		return nil, fmt.Errorf("error resolving %s's %s record: %v", fqdn, rec, err)
		// 	}

		// 	for _, i := range nss {
		// 		r.Answer = i.Host
		// 		resultRecords = append(resultRecords, r)
		// 	}

		default:
			r.Error = fmt.Errorf("unknown DNS record type: %s", rec)
			result.Records = append(result.Records, r)
		}
	}

	*responses = append(*responses, result)
}
