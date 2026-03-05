package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runGcloud(args ...string) (string, error) {
	cmd := exec.Command("gcloud", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gcloud error: %s", strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func runGcloudInteractive(args ...string) error {
	cmd := exec.Command("gcloud", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ensureGcloudConfig(name string) error {
	out, err := runGcloud("config", "configurations", "list", "--format=value(name)")
	if err != nil {
		return fmt.Errorf("failed to list configurations: %w", err)
	}
	for _, c := range strings.Split(out, "\n") {
		if strings.TrimSpace(c) == name {
			return nil
		}
	}
	fmt.Printf("  ⟳ Creating gcloud configuration '%s'...\n", name)
	_, err = runGcloud("config", "configurations", "create", name)
	return err
}

// resolveProjectID accepts a project name or project ID and returns the
// canonical project ID. If the lookup fails it returns the original value so
// the caller can surface the real gcloud error.
func resolveProjectID(nameOrID string) string {
	if nameOrID == "" {
		return ""
	}
	id, err := runGcloud("projects", "describe", nameOrID, "--format=value(projectId)")
	if err != nil || id == "" {
		return nameOrID
	}
	return id
}

func ActivateProfile(p Profile) error {
	if err := ensureGcloudConfig(p.Name); err != nil {
		return fmt.Errorf("configuration setup failed: %w", err)
	}

	fmt.Printf("  ⟳ Activating configuration '%s'...\n", p.Name)
	if _, err := runGcloud("config", "configurations", "activate", p.Name); err != nil {
		return fmt.Errorf("failed to activate configuration: %w", err)
	}

	steps := []struct {
		desc string
		fn   func() error
	}{
		{
			"Setting account",
			func() error {
				_, err := runGcloud("config", "set", "account", p.Account)
				return err
			},
		},
		{
			"Setting project",
			func() error {
				projectID := resolveProjectID(p.Project)
				if projectID != p.Project {
					fmt.Printf("     (resolved '%s' → '%s')\n", p.Project, projectID)
				}
				_, err := runGcloud("config", "set", "project", projectID)
				return err
			},
		},
		{
			"Setting region",
			func() error {
				if p.Region == "" {
					return nil
				}
				_, err := runGcloud("config", "set", "compute/region", p.Region)
				return err
			},
		},
		{
			"Setting zone",
			func() error {
				if p.Zone == "" {
					return nil
				}
				_, err := runGcloud("config", "set", "compute/zone", p.Zone)
				return err
			},
		},
	}

	for _, step := range steps {
		fmt.Printf("  ⟳ %s...\n", step.desc)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s failed: %w", step.desc, err)
		}
	}
	return nil
}

func DeleteGcloudConfig(name string) error {
	out, err := runGcloud("config", "configurations", "list", "--format=value(name)")
	if err != nil {
		return nil // best effort
	}
	for _, c := range strings.Split(out, "\n") {
		if strings.TrimSpace(c) == name {
			_, err = runGcloud("config", "configurations", "delete", name, "--quiet")
			return err
		}
	}
	return nil
}

func LoginAccount(account string) error {
	fmt.Printf("Opening browser for login: %s\n", account)
	return runGcloudInteractive("auth", "login", account)
}

func GetCurrentAccount() (string, error) {
	return runGcloud("config", "get-value", "account")
}

func GetCurrentProject() (string, error) {
	return runGcloud("config", "get-value", "project")
}

func ListAuthorizedAccounts() ([]string, error) {
	out, err := runGcloud("auth", "list", "--format=value(account)")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return []string{}, nil
	}
	return strings.Split(out, "\n"), nil
}

func IsAccountAuthorized(account string) (bool, error) {
	accounts, err := ListAuthorizedAccounts()
	if err != nil {
		return false, err
	}
	for _, a := range accounts {
		if strings.EqualFold(a, account) {
			return true, nil
		}
	}
	return false, nil
}
