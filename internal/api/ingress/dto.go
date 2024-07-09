package ingress

type Uploaded struct {
	RequestID string `json:"request_id"`
	Upload    Upload `json:"upload"`
}

type Upload struct {
	Account int
	OrgID   int
}
