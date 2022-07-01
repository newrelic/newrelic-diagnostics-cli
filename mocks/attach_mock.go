package mocks

import (
	"bytes"
	"net/http"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
)

type MAttachDeps struct{}

func (m MAttachDeps) GetWrapper(file *bytes.Reader, fileSize int64, attachmentKey string) httpHelper.RequestWrapper {
	headers := make(map[string]string)
	headers["Attachment-Key"] = attachmentKey

	wrapper := httpHelper.RequestWrapper{
		Method:         "POST",
		URL:            m.GetAttachmentsEndpoint(),
		Payload:        file,
		Length:         file.Size(),
		TimeoutSeconds: 1000,
	}
	wrapper.Headers = headers
	return wrapper
}

func (m MAttachDeps) GetFileSize(file string) int64 {
	var ret int64 = 1000
	return ret
}

func (m MAttachDeps) GetReader(file string) (*bytes.Reader, error) {
	return bytes.NewReader([]byte{'m', 'o', 'c', 'k'}), nil
}

func (m MAttachDeps) MakeRequest(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return httpHelper.MakeHTTPRequest(wrapper)
}

func (m MAttachDeps) GetAttachmentsEndpoint() string {
	return config.AttachmentEndpoint
}
