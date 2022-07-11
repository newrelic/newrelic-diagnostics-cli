package mocks

import "bytes"

type MAttachDeps struct{}

func (m MAttachDeps) GetFileSize(file string) int64 {
	var ret int64 = 4
	return ret
}

func (m MAttachDeps) GetReader(file string) (*bytes.Reader, error) {
	return bytes.NewReader([]byte{'m', 'o', 'c', 'k'}), nil
}
