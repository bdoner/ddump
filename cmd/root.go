package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bdoner/ddump/crt"
	"github.com/bdoner/ddump/dns"
	"github.com/spf13/cobra"
)

type config struct {
	domain        string
	topDomainOnly bool
}

var conf config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ddump",
	Short: "ddump is an utility to dump (mostly) DNS information about a domain.",
	Long:  `ddump is a tool used to discover subdomains, quickly get an overview of the DNS setup, check HTTP headers and more.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: runProgram,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runProgram(cmd *cobra.Command, args []string) {
	fmt.Printf("Looking for subdomains for: %s\n", conf.domain)

	err := crt.OpenDatabaseConnection()
	if err != nil {
		log.Fatalf("%v", err)
	}

	subdomains, err := crt.QueryDomain(conf.domain, conf.topDomainOnly)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if err := crt.Dispose(); err != nil {
		log.Fatalf("%v", err)
	}

	responses := make([]dns.ResolveResult, len(*subdomains))
	resolver := dns.GetDefaultResolver()

	var wg sync.WaitGroup

	for _, v := range *subdomains {
		wg.Add(1)
		go resolver.Resolve(v, &responses, &wg)
	}

	wg.Wait()

	for _, res := range responses {

		if res.Domain == "" {
			continue
		}

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
		fmt.Println("")
	}
}

func init() {
	rootCmd.Flags().StringVarP(&conf.domain, "domain", "d", "", "Domain to dump data for")
	rootCmd.MarkFlagRequired("domain")
	rootCmd.Flags().BoolVarP(&conf.topDomainOnly, "top-only", "t", false, "Do not find subdomains. Only lookup the given domain name.")

}
