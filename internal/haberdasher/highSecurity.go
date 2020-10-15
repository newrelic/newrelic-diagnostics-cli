package haberdasher

// HSMrequest is the request body structure for querying the /tasks/hsm endpoint
type HSMrequest struct {
	LicenseKeys []string `json:"licenseKeys"`
}

// HSMresponse is the top-level response body structure from the /tasks/hsm endpoint
type HSMresponse struct {
	Success bool               `json:"success"`
	Data    []HSMresult 	   `json:"data"`
}

// HSMresult is the response body -> "data" structure from the /tasks/hsm endpoint
type HSMresult struct {
	LicenseKey string   `json:"licenseKey"`
	IsEnabled    bool   `json:"result"`
}

// CheckHSM will reach out to the Haberdasher API to check the HSM state for the account associated with a license key
func (s *TasksService) CheckHSM(licenseKeys []string) ([]HSMresult, *Response, error) {
	uri := "/tasks/hsm"
	body := HSMrequest{
		LicenseKeys: licenseKeys,
	}

	req, err := s.client.NewRequest("POST", uri, body)
	if err != nil {
		return nil, nil, err
	}

	HSMresponse := &HSMresponse{}
	resp, err := s.client.Do(req, HSMresponse)

	return HSMresponse.Data, resp, err
}

