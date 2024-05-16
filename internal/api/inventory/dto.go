package inventory

type Hosts struct {
	Total   uint64 `json:"total"`
	Count   uint64 `json:"count"`
	Page    uint64 `json:"page"`
	PerPage uint64 `json:"per_page"`
	Results []Host `json:"results"`
}

type Host struct {
	InsightsInventoryID   string   `json:"id"`
	InsightsClientID      string   `json:"insights_id"`
	SubscriptionManagerID string   `json:"subscription_manager_id"`
	SatelliteID           string   `json:"satellite_id"`
	BiosUUID              string   `json:"bios_uuid"`
	IPAddresses           []string `json:"ip_addresses"`
	FQDN                  string   `json:"fqdn"`
	MACAddresses          []string `json:"mac_addresses"`
	ProviderID            string   `json:"provider_id"`
	ProviderType          string   `json:"provider_type"`
	Account               string   `json:"account"`
	DisplayName           string   `json:"display_name"`
	AnsibleHost           string   `json:"ansible_host"`
	Groups                []string `json:"groups"`
	Tags                  []any    `json:"tags"`
}
