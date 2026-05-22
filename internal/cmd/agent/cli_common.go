package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/henrygd/beszel/agent"
	"github.com/henrygd/beszel/agent/health"
	"github.com/spf13/pflag"
)

// handleCLISubcommand runs health, fingerprint, update, version, or help. Returns true if handled.
func handleCLISubcommand() bool {
	if len(os.Args) < 2 {
		return false
	}
	sub := os.Args[1]
	switch sub {
	case "health":
		if err := health.Check(); err != nil {
			log.Fatal(err)
		}
		fmt.Print("ok")
		return true
	case "fingerprint":
		handleFingerprint()
		return true
	case "update":
		china := false
		for _, a := range os.Args[2:] {
			if a == "-c" || a == "--china-mirrors" {
				china = true
			}
		}
		agent.Update(china)
		return true
	case "help", "-h", "--help":
		printCLIUsage()
		return true
	case "-v", "--version", "version":
		fmt.Println("beszel-agent", getBeszelVersion())
		return true
	}
	return false
}

func printCLIUsage() {
	fmt.Printf("Usage: %s [command] [flags]\n\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  fingerprint  View or reset the agent fingerprint")
	fmt.Println("  health       Check if the agent is running")
	fmt.Println("  update       Update to the latest version")
	fmt.Println()
	fmt.Println("Windows tray (default):")
	fmt.Println("  -service     Run headless (for NSSM / Windows Service)")
	fmt.Println("  -tray        Run with system tray (default when double-clicked)")
	fmt.Println()
	fmt.Println("Flags (CLI / -service mode):")
	pflag.CommandLine.SetOutput(os.Stdout)
	pflag.PrintDefaults()
}

func handleFingerprint() {
	subCmd := ""
	if len(os.Args) > 2 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "", "view":
		dataDir, _ := agent.GetDataDir()
		fmt.Println(agent.GetFingerprint(dataDir, "", ""))
	case "help", "-h", "--help":
		fmt.Print(fingerprintUsage())
	case "reset":
		dataDir, err := agent.GetDataDir()
		if err != nil {
			log.Fatal(err)
		}
		if err := agent.DeleteFingerprint(dataDir); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Fingerprint reset. A new one will be generated on next start.")
	default:
		log.Fatalf("Unknown command: %q\n\n%s", subCmd, fingerprintUsage())
	}
}

func fingerprintUsage() string {
	return fmt.Sprintf("Usage: %s fingerprint [view|reset]\n\nCommands:\n  view   Print fingerprint (default)\n  reset  Reset saved fingerprint\n", os.Args[0])
}

func hasFlag(name string) bool {
	for _, a := range os.Args[1:] {
		if a == name || strings.HasPrefix(a, name+"=") {
			return true
		}
	}
	return false
}
