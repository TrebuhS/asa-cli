package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/trebuhs/asa-cli/internal/output"
	"github.com/trebuhs/asa-cli/internal/services"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current user and organization info",
	Long:  "Verify authentication by fetching ACL info (GET /acls).",
	RunE:  runWhoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	client, err := newAPIClientNoOrg()
	if err != nil {
		return err
	}

	svc := services.NewACLService(client)
	acls, err := svc.GetACLs()
	if err != nil {
		return fmt.Errorf("fetching ACLs: %w", err)
	}

	if len(acls) == 0 {
		fmt.Println("No organizations found.")
		return nil
	}

	output.Print(getFormat(), acls, []output.Column{
		{Header: "ORG NAME", Field: "OrgName", Width: 30},
		{Header: "ORG ID", Field: "OrgID", Width: 15},
		{Header: "CURRENCY", Field: "Currency", Width: 10},
		{Header: "ROLES", Field: "RoleNames", Width: 40},
	})

	// For table format, also print a summary
	if getFormat() == output.FormatTable {
		fmt.Printf("\nAuthenticated. %d organization(s) accessible.\n", len(acls))
		for _, acl := range acls {
			fmt.Printf("  %s (ID: %d) — %s\n", acl.OrgName, acl.OrgID, strings.Join(acl.RoleNames, ", "))
		}
	}

	return nil
}
