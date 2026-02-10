package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/trebuhs/asa-cli/internal/config"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Apple Search Ads credentials",
	Long: `Configure credentials for Apple Search Ads API access.

Credential Setup:
  1. Sign in at https://ads.apple.com
  2. Go to Settings > API tab
  3. Create an API user (or use existing)
  4. Generate a key pair and upload the public key
  5. Note the Client ID, Team ID, Key ID from the API settings page
  6. Run: asa-cli configure --client-id "..." --team-id "..." --key-id "..." --private-key-path "/path/to/key.pem"
  7. Verify with: asa-cli whoami

Org ID is optional — if your account has a single org, it's auto-detected.
For multiple orgs, set it via --org-id flag or in config.`,
	RunE: runConfigure,
}

var (
	cfgClientID       string
	cfgTeamID         string
	cfgKeyID          string
	cfgOrgID          string
	cfgPrivateKeyPath string
)

func init() {
	configureCmd.Flags().StringVar(&cfgClientID, "client-id", "", "Apple Search Ads Client ID")
	configureCmd.Flags().StringVar(&cfgTeamID, "team-id", "", "Apple Developer Team ID")
	configureCmd.Flags().StringVar(&cfgKeyID, "key-id", "", "API Key ID")
	configureCmd.Flags().StringVar(&cfgOrgID, "org-id", "", "Organization ID (optional — auto-detected for single-org accounts)")
	configureCmd.Flags().StringVar(&cfgPrivateKeyPath, "private-key-path", "", "Path to private key (.pem or .p8 file)")
	rootCmd.AddCommand(configureCmd)
}

func runConfigure(cmd *cobra.Command, args []string) error {
	// If no flags provided, run interactive mode
	if cfgClientID == "" && cfgTeamID == "" && cfgKeyID == "" && cfgOrgID == "" && cfgPrivateKeyPath == "" {
		return runInteractiveConfigure()
	}

	// Non-interactive mode — validate required fields (org-id is optional)
	if cfgClientID == "" || cfgTeamID == "" || cfgKeyID == "" || cfgPrivateKeyPath == "" {
		return fmt.Errorf("required flags: --client-id, --team-id, --key-id, --private-key-path\nOptional: --org-id (auto-detected for single-org accounts)")
	}

	cfgPrivateKeyPath = expandPath(cfgPrivateKeyPath)

	// Validate key file exists
	if _, err := os.Stat(cfgPrivateKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("private key file not found: %s", cfgPrivateKeyPath)
	}

	cfg := &config.Config{
		ClientID:       cfgClientID,
		TeamID:         cfgTeamID,
		KeyID:          cfgKeyID,
		OrgID:          cfgOrgID,
		PrivateKeyPath: cfgPrivateKeyPath,
	}

	if err := config.Save(cfg, profileName); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	profile := profileName
	if profile == "" {
		profile = "default"
	}
	fmt.Printf("Configuration saved for profile '%s'.\n", profile)
	fmt.Println("Verify with: asa-cli whoami")
	return nil
}

func runInteractiveConfigure() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Apple Search Ads CLI Configuration")
	fmt.Println("===================================")
	fmt.Println()
	fmt.Println("You'll need your API credentials from https://ads.apple.com (Settings > API tab).")
	fmt.Println()

	clientID := prompt(reader, "Client ID")
	teamID := prompt(reader, "Team ID")
	keyID := prompt(reader, "Key ID")
	orgID := promptOptional(reader, "Org ID (press Enter to skip — auto-detected for single-org accounts)")
	privateKeyPath := expandPath(prompt(reader, "Private Key Path (.pem or .p8 file)"))

	// Validate key file
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("private key file not found: %s", privateKeyPath)
	}

	cfg := &config.Config{
		ClientID:       clientID,
		TeamID:         teamID,
		KeyID:          keyID,
		OrgID:          orgID,
		PrivateKeyPath: privateKeyPath,
	}

	if err := config.Save(cfg, profileName); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	profile := profileName
	if profile == "" {
		profile = "default"
	}
	fmt.Printf("\nConfiguration saved for profile '%s'.\n", profile)
	fmt.Println("Verify with: asa-cli whoami")
	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func prompt(reader *bufio.Reader, label string) string {
	for {
		fmt.Printf("%s: ", label)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			return input
		}
		fmt.Println("  Value cannot be empty. Please try again.")
	}
}

func promptOptional(reader *bufio.Reader, label string) string {
	fmt.Printf("%s: ", label)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
