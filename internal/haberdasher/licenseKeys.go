package haberdasher

// LicenseKeyRequest is the request body structure for querying the /tasks/license-key endpoint
type LicenseKeyRequest struct {
	LicenseKeys []string `json:"licenseKeys"`
}

// LicenseKeyResponse is the top-level response body structure from the /tasks/license-key endpoint
type LicenseKeyResponse struct {
	Success bool               `json:"success"`
	Data    []LicenseKeyResult `json:"data"`
}

// LicenseKeyResult is the response body -> "data" structure from the /tasks/license-key endpoint
type LicenseKeyResult struct {
	LicenseKey string `json:"licenseKey"`
	IsValid    bool   `json:"result"`
}

// ValidateLicenseKeys will reach out to the Haberdasher API to validate a slice of license keys
func (s *TasksService) ValidateLicenseKeys(licenseKeys []string) ([]LicenseKeyResult, *Response, error) {
	uri := "/tasks/license-key"
	body := LicenseKeyRequest{
		LicenseKeys: licenseKeys,
	}

	req, err := s.client.NewRequest("POST", uri, body)
	if err != nil {
		return nil, nil, err
	}

	licenseKeyResponse := &LicenseKeyResponse{}
	resp, err := s.client.Do(req, licenseKeyResponse)

	return licenseKeyResponse.Data, resp, err
}
