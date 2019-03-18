package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/bdoner/ddump/crt"
	"github.com/bdoner/ddump/dns"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

type config struct {
	domain string
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
		fmt.Println("")
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ddump.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&conf.domain, "domain", "d", "", "Domain to dump data for")
	rootCmd.MarkFlagRequired("domain")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".ddump" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ddump")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
