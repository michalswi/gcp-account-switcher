package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cmdAdd(args []string) error {
	var (
		name    string
		account string
		project string
		region  string
		zone    string
		domain  string
		desc    string
		login   bool
		skip    bool

		// track which fields were explicitly set via flags (skip prompting for those)
		flagProject bool
		flagRegion  bool
		flagZone    bool
	)

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name", "-n":
			name = nextArg(args, &i)
		case "--account", "-a":
			account = nextArg(args, &i)
		case "--project", "-p":
			project = nextArg(args, &i)
			flagProject = true
		case "--region", "-r":
			region = nextArg(args, &i)
			flagRegion = true
		case "--zone", "-z":
			zone = nextArg(args, &i)
			flagZone = true
		case "--domain", "-d":
			domain = nextArg(args, &i)
		case "--desc":
			desc = nextArg(args, &i)
		case "--login", "-l":
			login = true
		case "--skip", "-s":
			skip = true
		}
	}

	name = promptIfEmpty("Profile name", name)
	account = promptIfEmpty("Account (email)", account)
	if !skip {
		domain = promptIfEmpty("Domain (optional)", domain)
		desc = promptIfEmpty("Description (optional)", desc)
	}

	// For project/region/zone: if supplied via flag, use as-is.
	// Otherwise, pre-populate from current gcloud state and let the user confirm or override.
	if !flagProject {
		cur, _ := GetCurrentProject()
		project = promptWithDefault("Project ID", gcloudVal(cur))
	}
	if !flagRegion && !skip {
		cur, _ := runGcloud("config", "get-value", "compute/region")
		region = promptWithDefault("Default region (optional, e.g. us-central1)", gcloudVal(cur))
	}
	if !flagZone && !skip {
		cur, _ := runGcloud("config", "get-value", "compute/zone")
		zone = promptWithDefault("Default zone (optional, e.g. us-central1-a)", gcloudVal(cur))
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if existing, _ := cfg.FindProfile(name); existing != nil {
		return fmt.Errorf("profile '%s' already exists, delete it first with: gcps delete %s", name, name)
	}

	cfg.Profiles = append(cfg.Profiles, Profile{
		Name:        name,
		Account:     account,
		Project:     project,
		Region:      region,
		Zone:        zone,
		Domain:      domain,
		Description: desc,
	})

	if err := saveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("✓ Profile '%s' added\n", name)

	if login {
		if err := LoginAccount(account); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		profile, _ := cfg.FindProfile(name)
		if err := ActivateProfile(*profile); err != nil {
			return err
		}

		cfg.ActiveProfile = name
		if err := saveConfig(cfg); err != nil {
			return err
		}

		fmt.Printf("\n✅ Switched to profile '%s'\n", name)
	}
	return nil
}

func cmdUse(args []string) error {
	var (
		name       string
		forceLogin bool
	)

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--login", "-l":
			forceLogin = true
		default:
			if !strings.HasPrefix(args[i], "-") && name == "" {
				name = args[i]
			}
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles found, run 'gcps add' first")
	}

	if name == "" {
		name, err = selectProfile(cfg.Profiles)
		if err != nil {
			return err
		}
	}

	profile, _ := cfg.FindProfile(name)
	if profile == nil {
		return fmt.Errorf("profile '%s' not found", name)
	}

	fmt.Printf("\n🔄 Switching to profile: %s\n", profile.Name)
	if profile.Domain != "" {
		fmt.Printf("   Domain  : %s\n", profile.Domain)
	}
	fmt.Printf("   Account : %s\n", profile.Account)
	fmt.Printf("   Project : %s\n\n", profile.Project)

	authorized, err := IsAccountAuthorized(profile.Account)
	if err != nil {
		fmt.Printf("⚠ Could not check auth status: %v\n", err)
	}

	if !authorized || forceLogin {
		fmt.Printf("⚠ Account '%s' not logged in, starting auth...\n", profile.Account)
		if err := LoginAccount(profile.Account); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	if err := ActivateProfile(*profile); err != nil {
		return err
	}

	cfg.ActiveProfile = name
	if err := saveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("\n✅ Switched to profile '%s'\n", name)
	fmt.Printf("   Run 'gcloud auth list' to verify\n\n")
	return nil
}

func cmdList() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured. Run 'gcps add' to add one.")
		return nil
	}

	fmt.Printf("\n%-15s %-30s %-25s %-15s %s\n",
		"NAME", "ACCOUNT", "PROJECT", "REGION", "STATUS")
	fmt.Println(strings.Repeat("─", 95))

	for _, p := range cfg.Profiles {
		status := ""
		if p.Name == cfg.ActiveProfile {
			status = "● active"
		}
		fmt.Printf("%-15s %-30s %-25s %-15s %s\n",
			p.Name, p.Account, p.Project, p.Region, status)
		if p.Description != "" {
			fmt.Printf("  └─ %s\n", p.Description)
		}
	}
	fmt.Println()
	return nil
}

func cmdCurrent() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	account, _ := GetCurrentAccount()
	project, _ := GetCurrentProject()

	fmt.Printf("\n📍 Current gcloud state:\n")
	fmt.Printf("   Account : %s\n", account)
	fmt.Printf("   Project : %s\n\n", project)

	if cfg.ActiveProfile != "" {
		if p, _ := cfg.FindProfile(cfg.ActiveProfile); p != nil {
			fmt.Printf("📌 Active gcps profile : %s\n", p.Name)
			if p.Domain != "" {
				fmt.Printf("   Domain              : %s\n", p.Domain)
			}
			if p.Description != "" {
				fmt.Printf("   Description         : %s\n", p.Description)
			}
		}
	}

	fmt.Println("\n🔑 Authorized accounts:")
	accounts, err := ListAuthorizedAccounts()
	if err != nil {
		fmt.Printf("   (could not retrieve: %v)\n", err)
	} else {
		for _, a := range accounts {
			marker := "  "
			if a == account {
				marker = "→"
			}
			fmt.Printf("   %s %s\n", marker, a)
		}
	}
	fmt.Println()
	return nil
}

func cmdDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gcps delete <profile-name>")
	}
	name := args[0]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	_, idx := cfg.FindProfile(name)
	if idx == -1 {
		return fmt.Errorf("profile '%s' not found", name)
	}

	cfg.Profiles = append(cfg.Profiles[:idx], cfg.Profiles[idx+1:]...)
	if cfg.ActiveProfile == name {
		cfg.ActiveProfile = ""
	}

	if err := saveConfig(cfg); err != nil {
		return err
	}

	if err := DeleteGcloudConfig(name); err != nil {
		fmt.Printf("⚠ Could not remove gcloud configuration '%s': %v\n", name, err)
	}

	fmt.Printf("✓ Profile '%s' deleted\n", name)
	return nil
}

func cmdInit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gcps init <profile-name>")
	}
	name := args[0]

	account, err := GetCurrentAccount()
	if err != nil || account == "" {
		return fmt.Errorf("no active gcloud account found")
	}
	project, _ := GetCurrentProject()

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if existing, _ := cfg.FindProfile(name); existing != nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	cfg.Profiles = append(cfg.Profiles, Profile{
		Name:    name,
		Account: account,
		Project: project,
	})

	if err := saveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("✓ Profile '%s' created from current gcloud state\n", name)
	fmt.Printf("  Account : %s\n", account)
	fmt.Printf("  Project : %s\n", project)
	return nil
}

// gcloudVal normalises gcloud config get-value output: returns "" for "(unset)".
func gcloudVal(v string) string {
	if v == "(unset)" {
		return ""
	}
	return v
}

func nextArg(args []string, i *int) string {
	*i++
	if *i >= len(args) {
		return ""
	}
	return args[*i]
}

func promptIfEmpty(label, value string) string {
	if value != "" {
		return value
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", label)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// promptWithDefault shows the current default value and returns it unchanged if
// the user presses Enter without typing anything.
func promptWithDefault(label, defaultVal string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultVal
	}
	return text
}

func selectProfile(profiles []Profile) (string, error) {
	fmt.Println("\nAvailable profiles:")
	for i, p := range profiles {
		desc := ""
		if p.Domain != "" {
			desc = fmt.Sprintf(" (%s)", p.Domain)
		}
		fmt.Printf("  [%d] %-12s%s — %s / %s\n", i+1, p.Name, desc, p.Account, p.Project)
	}
	fmt.Print("\nSelect profile number: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(profiles) {
		return "", fmt.Errorf("invalid selection '%s'", input)
	}
	return profiles[idx-1].Name, nil
}
