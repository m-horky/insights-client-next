package inventory

import (
	"time"
)

// Hosts object is returned by Inventory `/hosts/` endpoint.
type Hosts struct {
	Total   uint64 `json:"total"`
	Count   uint64 `json:"count"`
	Page    uint64 `json:"page"`
	PerPage uint64 `json:"per_page"`
	Results []Host `json:"results"`
}

// Host object is contained in Hosts object.
type Host struct {
	InsightsInventoryID   string                       `json:"id"`
	InsightsClientID      string                       `json:"insights_id"`
	SubscriptionManagerID string                       `json:"subscription_manager_id"`
	SatelliteID           string                       `json:"satellite_id"`
	BiosUUID              string                       `json:"bios_uuid"`
	IPAddresses           []string                     `json:"ip_addresses"`
	FQDN                  string                       `json:"fqdn"`
	MACAddresses          []string                     `json:"mac_addresses"`
	ProviderID            string                       `json:"provider_id"`
	ProviderType          string                       `json:"provider_type"`
	Account               string                       `json:"account"`
	OrganizationID        string                       `json:"org_id"`
	DisplayName           string                       `json:"display_name"`
	AnsibleHost           string                       `json:"ansible_host"`
	Groups                []map[string]string          `json:"groups"`
	Tags                  []any                        `json:"tags"`
	Facts                 []any                        `json:"facts"`
	Reporter              string                       `json:"reporter"`
	PerReporterStaleness  map[string]ReporterStaleness `json:"per_reporter_staleness"`
	StaleTimestamp        time.Time                    `json:"stale_timestamp"`
	StaleWarningTimestamp time.Time                    `json:"stale_warning_timestamp"`
	CulledTimestamp       time.Time                    `json:"culled_timestamp"`
	Created               time.Time                    `json:"created"`
	Updated               time.Time                    `json:"updated"`
}

type ReporterStaleness struct {
	CheckInSucceeded      bool      `json:"check_in_succeeded"`
	StaleTimestamp        time.Time `json:"stale_timestamp"`
	LastCheckIn           time.Time `json:"last_check_in"`
	StaleWarningTimestamp time.Time `json:"stale_warning_timestamp"`
	CulledTimestamp       time.Time `json:"culled_timestamp"`
}

// HostID object is returned by Inventory `/host_exists` endpoint.
type HostID struct {
	InsightsInventoryID string `json:"id"`
}
