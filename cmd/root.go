package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/trebuhs/asa-cli/internal/api"
	"github.com/trebuhs/asa-cli/internal/auth"
	"github.com/trebuhs/asa-cli/internal/config"
	"github.com/trebuhs/asa-cli/internal/models"
	"github.com/trebuhs/asa-cli/internal/output"
)

var (
	outputFormat string
	profileName  string
	verbose      bool
	noColor      bool
	globalOrgID  string
)

var rootCmd = &cobra.Command{
	Use:   "asa-cli",
	Short: "Apple Search Ads CLI",
	Long:  "A command-line interface for the Apple Search Ads Campaign Management API v5.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
		config.SetProfile(profileName)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: json or table")
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "", "Config profile name")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().StringVar(&globalOrgID, "org-id", "", "Organization ID (overrides config)")
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	return nil
}

// getFormat returns the output format.
func getFormat() output.Format {
	switch strings.ToLower(outputFormat) {
	case "json":
		return output.FormatJSON
	default:
		return output.FormatTable
	}
}

// newAPIClient creates an authenticated API client from config.
func newAPIClient() (*api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if err := auth.ValidateConfig(cfg); err != nil {
		return nil, err
	}

	// Resolve org ID: flag > config > auto-detect
	orgID := cfg.OrgID
	if globalOrgID != "" {
		orgID = globalOrgID
	}

	tokenProvider := auth.NewTokenProvider(cfg)

	// If no org ID configured, auto-resolve from /acls
	if orgID == "" {
		resolved, err := resolveOrgID(tokenProvider)
		if err != nil {
			return nil, err
		}
		orgID = resolved
	}

	transport := &auth.Transport{
		Token:   tokenProvider,
		OrgID:   orgID,
		Verbose: verbose,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	client := api.NewClient(httpClient)
	client.Verbose = verbose
	return client, nil
}

// newAPIClientNoOrg creates an authenticated client without requiring an org ID.
// Used for commands like whoami that don't need X-AP-Context.
func newAPIClientNoOrg() (*api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if err := auth.ValidateConfig(cfg); err != nil {
		return nil, err
	}

	tokenProvider := auth.NewTokenProvider(cfg)
	transport := &auth.Transport{
		Token:   tokenProvider,
		Verbose: verbose,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	client := api.NewClient(httpClient)
	client.Verbose = verbose
	return client, nil
}

// resolveOrgID fetches /acls and auto-selects the org if there's exactly one.
func resolveOrgID(tokenProvider *auth.TokenProvider) (string, error) {
	transport := &auth.Transport{
		Token:   tokenProvider,
		Verbose: verbose,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	req, err := http.NewRequest("GET", api.BaseURL+"/acls", nil)
	if err != nil {
		return "", fmt.Errorf("creating ACL request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching orgs: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading org response: %w", err)
	}

	var apiResp struct {
		Data []models.UserACL `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("parsing org response: %w", err)
	}

	switch len(apiResp.Data) {
	case 0:
		return "", fmt.Errorf("no organizations found for this account")
	case 1:
		orgID := strconv.FormatInt(apiResp.Data[0].OrgID, 10)
		if verbose {
			fmt.Printf("Auto-selected org: %s (ID: %s)\n", apiResp.Data[0].OrgName, orgID)
		}
		return orgID, nil
	default:
		var lines []string
		for _, acl := range apiResp.Data {
			lines = append(lines, fmt.Sprintf("  %s (ID: %d)", acl.OrgName, acl.OrgID))
		}
		return "", fmt.Errorf("multiple organizations found. Use --org-id flag or set org_id in config:\n%s", strings.Join(lines, "\n"))
	}
}

// parseFilters parses filter strings like "status=ENABLED" into Conditions.
func parseFilters(filters []string) []models.Condition {
	var conditions []models.Condition
	for _, f := range filters {
		// Find operator (check multi-char operators first)
		for _, op := range []string{">=", "<=", "!~", "=", "~", "@", ">", "<"} {
			idx := strings.Index(f, op)
			if idx > 0 {
				field := f[:idx]
				value := f[idx+len(op):]
				apiOp := models.ParseFilterOperator(op)

				var values []string
				if op == "@" {
					values = strings.Split(value, ",")
				} else {
					values = []string{value}
				}

				conditions = append(conditions, models.Condition{
					Field:    field,
					Operator: apiOp,
					Values:   values,
				})
				break
			}
		}
	}
	return conditions
}

// parseSorts parses sort strings like "name:asc" into OrderByItems.
func parseSorts(sorts []string) []models.OrderByItem {
	var items []models.OrderByItem
	for _, s := range sorts {
		parts := strings.SplitN(s, ":", 2)
		field := parts[0]
		order := "ASCENDING"
		if len(parts) > 1 {
			order = models.ParseSortOrder(parts[1])
		}
		items = append(items, models.OrderByItem{
			Field:     field,
			SortOrder: order,
		})
	}
	return items
}

// exitWithError prints an error and exits with the given code.
func exitWithError(msg string, code int) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(code)
}
