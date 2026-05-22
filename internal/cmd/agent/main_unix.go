//go:build !windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/henrygd/beszel"
	"github.com/henrygd/beszel/agent"
	"github.com/henrygd/beszel/agent/utils"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
)

type cmdOptions struct {
	key    string
	listen string
	hubURL string
	token  string
}

func (opts *cmdOptions) parse() bool {
	subcommand := ""
	if len(os.Args) > 1 {
		subcommand = os.Args[1]
	}

	pflag.StringVarP(&opts.key, "key", "k", "", "Public key(s) for SSH authentication")
	pflag.StringVarP(&opts.listen, "listen", "l", "", "Address or port to listen on")
	pflag.StringVarP(&opts.hubURL, "url", "u", "", "URL of the Beszel hub")
	pflag.StringVarP(&opts.token, "token", "t", "", "Token to use for authentication")
	chinaMirrors := pflag.BoolP("china-mirrors", "c", false, "Use mirror for update (gh.beszel.dev) instead of GitHub")
	version := pflag.BoolP("version", "v", false, "Show version information")
	help := pflag.BoolP("help", "h", false, "Show this help message")

	flagsToConvert := []string{"key", "listen", "url", "token"}
	for i, arg := range os.Args {
		for _, flag := range flagsToConvert {
			singleDash := "-" + flag
			doubleDash := "--" + flag
			if arg == singleDash {
				os.Args[i] = doubleDash
				break
			} else if strings.HasPrefix(arg, singleDash+"=") {
				os.Args[i] = doubleDash + arg[len(singleDash):]
				break
			}
		}
	}

	pflag.Usage = func() {
		fmt.Printf("Usage: %s [command] [flags]\n\n", os.Args[0])
		fmt.Println("Commands:")
		fmt.Println("  fingerprint  View or reset the agent fingerprint")
		fmt.Println("  health       Check if the agent is running")
		fmt.Println("  update       Update to the latest version")
		fmt.Println("\nFlags:")
		pflag.PrintDefaults()
	}

	pflag.Parse()

	switch {
	case *version:
		fmt.Println(beszel.AppName+"-agent", beszel.Version)
		return true
	case *help || subcommand == "help":
		pflag.Usage()
		return true
	case subcommand == "update":
		agent.Update(*chinaMirrors)
		return true
	}

	if opts.hubURL != "" {
		os.Setenv("HUB_URL", opts.hubURL)
	}
	if opts.token != "" {
		os.Setenv("TOKEN", opts.token)
	}
	return false
}

func (opts *cmdOptions) loadPublicKeys() ([]ssh.PublicKey, error) {
	if opts.key != "" {
		return agent.ParseKeys(opts.key)
	}
	if key, ok := utils.GetEnv("KEY"); ok && key != "" {
		return agent.ParseKeys(key)
	}
	keyFile, ok := utils.GetEnv("KEY_FILE")
	if !ok {
		return nil, fmt.Errorf("no key provided: must set -key flag, KEY env var, or KEY_FILE env var")
	}
	pubKey, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}
	return agent.ParseKeys(string(pubKey))
}

func main() {
	if handleCLISubcommand() {
		return
	}
	var opts cmdOptions
	if opts.parse() {
		return
	}
	runHeadlessAgent(opts)
}

func runHeadlessAgent(opts cmdOptions) {
	serverConfig, err := buildServerOptions(opts)
	if err != nil {
		log.Fatal(err)
	}
	a, err := agent.NewAgent()
	if err != nil {
		log.Fatal("Failed to create agent: ", err)
	}
	if err := a.Start(serverConfig); err != nil {
		log.Fatal("Failed to start: ", err)
	}
}

func (opts *cmdOptions) getAddress() string {
	return agent.GetAddress(opts.listen)
}

func buildServerOptions(opts cmdOptions) (agent.ServerOptions, error) {
	var serverConfig agent.ServerOptions
	var err error
	serverConfig.Keys, err = opts.loadPublicKeys()
	if err != nil {
		return serverConfig, err
	}
	addr := agent.GetAddress(opts.listen)
	serverConfig.Addr = addr
	serverConfig.Network = agent.GetNetwork(addr)
	return serverConfig, nil
}
