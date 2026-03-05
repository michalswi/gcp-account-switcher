package main

import (
	"fmt"
	"os"
)

const usage = `gcps - GCP Account Switcher

Usage:
  gcps <command> [options]

Commands:
  add                          Add a new profile (interactive or via flags)
  use [profile] [--login]      Switch to a profile (interactive picker if no name given)
  list                         List all profiles
  current                      Show active profile and gcloud state
  delete <profile>             Delete a profile
  init <profile>               Create a profile from current gcloud state

Options for 'add':
  --name,    -n  Profile alias
  --account, -a  GCP account email
  --project, -p  GCP project ID
  --region,  -r  Default compute region
  --zone,    -z  Default compute zone
  --domain,  -d  Domain/org label (for reference)
  --desc         Description
  --login,   -l  Trigger gcloud auth login after adding

Examples:
  gcps add --name work --account alice@company.com --project company-prod --login
  gcps add --name personal --account alice@gmail.com --project my-proj
  gcps use work
  gcps use                     # interactive picker
  gcps use work --login        # force re-auth
  gcps list
  gcps current
  gcps init staging            # snapshot current gcloud state
  gcps delete personal
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(0)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "add":
		err = cmdAdd(args)
	case "use", "sw", "switch":
		err = cmdUse(args)
	case "list", "ls":
		err = cmdList()
	case "current":
		err = cmdCurrent()
	case "delete", "rm":
		err = cmdDelete(args)
	case "init":
		err = cmdInit(args)
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
