package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/trebuhs/asa-cli/internal/models"
	"github.com/trebuhs/asa-cli/internal/output"
	"github.com/trebuhs/asa-cli/internal/services"
)

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Pull campaign reports",
}

var reportsCampaignsCmd = &cobra.Command{
	Use:   "campaigns",
	Short: "Campaign-level report",
	RunE:  runReportCampaigns,
}

var reportsAdGroupsCmd = &cobra.Command{
	Use:   "adgroups",
	Short: "Ad group-level report",
	RunE:  runReportAdGroups,
}

var reportsKeywordsCmd = &cobra.Command{
	Use:   "keywords",
	Short: "Keyword-level report",
	RunE:  runReportKeywords,
}

var reportsSearchTermsCmd = &cobra.Command{
	Use:   "search-terms",
	Short: "Search terms report",
	RunE:  runReportSearchTerms,
}

var (
	rptStartDate   string
	rptEndDate     string
	rptGranularity string
	rptGroupBy     string
	rptCampaignID  int64
	rptLimit       int
	rptGrandTotals bool
)

func init() {
	// Common flags for all report commands
	for _, cmd := range []*cobra.Command{reportsCampaignsCmd, reportsAdGroupsCmd, reportsKeywordsCmd, reportsSearchTermsCmd} {
		cmd.Flags().StringVar(&rptStartDate, "start-date", "", "Start date (YYYY-MM-DD) (required)")
		cmd.Flags().StringVar(&rptEndDate, "end-date", "", "End date (YYYY-MM-DD) (required)")
		cmd.Flags().StringVar(&rptGranularity, "granularity", "", "Granularity: HOURLY, DAILY, WEEKLY, MONTHLY")
		cmd.Flags().StringVar(&rptGroupBy, "group-by", "", "Comma-separated group by fields (e.g. countryOrRegion,deviceClass)")
		cmd.Flags().IntVar(&rptLimit, "limit", 1000, "Result limit")
		cmd.Flags().BoolVar(&rptGrandTotals, "grand-totals", false, "Include grand totals")
		cmd.MarkFlagRequired("start-date")
		cmd.MarkFlagRequired("end-date")
	}

	// Campaign ID for sub-entity reports
	for _, cmd := range []*cobra.Command{reportsAdGroupsCmd, reportsKeywordsCmd, reportsSearchTermsCmd} {
		cmd.Flags().Int64Var(&rptCampaignID, "campaign-id", 0, "Campaign ID (required)")
		cmd.MarkFlagRequired("campaign-id")
	}

	reportsCmd.AddCommand(reportsCampaignsCmd, reportsAdGroupsCmd, reportsKeywordsCmd, reportsSearchTermsCmd)
	rootCmd.AddCommand(reportsCmd)
}

func buildReportRequest() *models.ReportRequest {
	req := &models.ReportRequest{
		StartTime:         rptStartDate,
		EndTime:           rptEndDate,
		ReturnGrandTotals: rptGrandTotals,
		ReturnRowTotals:   true,
		Selector: &models.Selector{
			OrderBy: []models.OrderByItem{
				{Field: "localSpend", SortOrder: "DESCENDING"},
			},
			Pagination: models.SelectorPagination{
				Offset: 0,
				Limit:  rptLimit,
			},
		},
	}

	if rptGranularity != "" {
		req.Granularity = strings.ToUpper(rptGranularity)
	}

	if rptGroupBy != "" {
		req.GroupBy = strings.Split(rptGroupBy, ",")
	}

	return req
}

func printReport(resp *models.ReportingDataResponse) {
	if getFormat() == output.FormatJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(resp)
		return
	}

	// Table format — print summary
	if resp == nil || len(resp.Row) == 0 {
		fmt.Println("No report data.")
		return
	}

	// Print each row
	for _, row := range resp.Row {
		if row.Metadata != nil {
			for k, v := range row.Metadata {
				fmt.Printf("%s: %v  ", k, v)
			}
			fmt.Println()
		}

		if row.Total != nil {
			printMetricsRow(row.Total)
		}

		for _, g := range row.Granularity {
			fmt.Printf("  Date: %s\n", g.Date)
			if g.Metrics != nil {
				printMetricsRow(g.Metrics)
			}
		}
		fmt.Println("---")
	}

	if resp.GrandTotals != nil && resp.GrandTotals.Total != nil {
		fmt.Println("\nGRAND TOTALS:")
		printMetricsRow(resp.GrandTotals.Total)
	}
}

func printMetricsRow(m *models.SpendRow) {
	fmt.Printf("  Impressions: %d | Taps: %d | Installs: %d (tap: %d, view: %d) | NewDL: %d | Redownloads: %d\n",
		m.Impressions, m.Taps, m.TotalInstalls, m.TapInstalls, m.ViewInstalls, m.TotalNewDownloads, m.TotalRedownloads)
	fmt.Printf("  TTR: %.4f | InstallRate: %.4f (tap: %.4f) | CPI: %s %s | AvgCPT: %s %s | Spend: %s %s\n",
		m.TTR, m.TotalInstallRate, m.TapInstallRate,
		m.TotalAvgCPI.Amount, m.TotalAvgCPI.Currency,
		m.AvgCPT.Amount, m.AvgCPT.Currency,
		m.LocalSpend.Amount, m.LocalSpend.Currency)
}

func runReportCampaigns(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	svc := services.NewReportingService(client)
	resp, err := svc.GetCampaignReport(buildReportRequest())
	if err != nil {
		return fmt.Errorf("getting campaign report: %w", err)
	}

	printReport(resp)
	return nil
}

func runReportAdGroups(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	svc := services.NewReportingService(client)
	resp, err := svc.GetAdGroupReport(rptCampaignID, buildReportRequest())
	if err != nil {
		return fmt.Errorf("getting ad group report: %w", err)
	}

	printReport(resp)
	return nil
}

func runReportKeywords(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	svc := services.NewReportingService(client)
	resp, err := svc.GetKeywordReport(rptCampaignID, buildReportRequest())
	if err != nil {
		return fmt.Errorf("getting keyword report: %w", err)
	}

	printReport(resp)
	return nil
}

func runReportSearchTerms(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	svc := services.NewReportingService(client)
	resp, err := svc.GetSearchTermReport(rptCampaignID, buildReportRequest())
	if err != nil {
		return fmt.Errorf("getting search terms report: %w", err)
	}

	printReport(resp)
	return nil
}
