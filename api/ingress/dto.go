package ingress

// Archive is an argument for UploadArchive.
type Archive struct {
	ContentType string
	Path        string
}

// Uploaded object is returned by Ingress on successful upload.
type Uploaded struct {
	RequestID string `json:"request_id"`
	Upload    Upload `json:"upload"`
}

// Upload object is contained in Uploaded object.
type Upload struct {
	Account int
	OrgID   int
}
