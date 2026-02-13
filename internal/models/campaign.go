package models

// Campaign represents an Apple Search Ads campaign.
type Campaign struct {
	ID                                 int64                  `json:"id,omitempty"`
	OrgID                              int64                  `json:"orgId,omitempty"`
	Name                               string                 `json:"name"`
	BudgetAmount                       *Money                 `json:"budgetAmount,omitempty"`
	DailyBudgetAmount                  *Money                 `json:"dailyBudgetAmount,omitempty"`
	AdamID                             int64                  `json:"adamId,omitempty"`
	PaymentModel                       string                 `json:"paymentModel,omitempty"`
	Status                             string                 `json:"status,omitempty"`
	ServingStatus                      string                 `json:"servingStatus,omitempty"`
	ServingStateReasons                []string               `json:"servingStateReasons,omitempty"`
	DisplayStatus                      string                 `json:"displayStatus,omitempty"`
	SupplySources                      []string               `json:"supplySources,omitempty"`
	AdChannelType                      string                 `json:"adChannelType,omitempty"`
	BillingEvent                       string                 `json:"billingEvent,omitempty"`
	CountriesOrRegions                 []string               `json:"countriesOrRegions,omitempty"`
	CountryOrRegionServingStateReasons map[string]interface{} `json:"countryOrRegionServingStateReasons,omitempty"`
	ModificationTime                   string                 `json:"modificationTime,omitempty"`
	StartTime                          string                 `json:"startTime,omitempty"`
	EndTime                            string                 `json:"endTime,omitempty"`
	LOCInvoiceDetails                  *LOCInvoiceDetails     `json:"locInvoiceDetails,omitempty"`
}

// LOCInvoiceDetails for billing.
type LOCInvoiceDetails struct {
	BillingContactEmail string `json:"billingContactEmail,omitempty"`
	BuyerName           string `json:"buyerName,omitempty"`
	BuyerEmail          string `json:"buyerEmail,omitempty"`
	OrderNumber         string `json:"orderNumber,omitempty"`
	ClientName          string `json:"clientName,omitempty"`
}

// CampaignUpdate contains fields that can be updated on a campaign.
type CampaignUpdate struct {
	Name               string   `json:"name,omitempty"`
	BudgetAmount       *Money   `json:"budgetAmount,omitempty"`
	DailyBudgetAmount  *Money   `json:"dailyBudgetAmount,omitempty"`
	Status             string   `json:"status,omitempty"`
	CountriesOrRegions []string `json:"countriesOrRegions,omitempty"`
}

// UpdateCampaignRequest is the v5 update payload wrapper.
type UpdateCampaignRequest struct {
	Campaign                                 *CampaignUpdate `json:"campaign,omitempty"`
	ClearGeoTargetingOnCountryOrRegionChange bool            `json:"clearGeoTargetingOnCountryOrRegionChange,omitempty"`
}
