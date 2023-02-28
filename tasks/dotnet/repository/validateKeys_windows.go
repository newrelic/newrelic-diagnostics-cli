//go:build windows
// +build windows

package repository

import (
	"github.com/newrelic/newrelic-diagnostics-cli/domain/entity"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"golang.org/x/sys/windows/registry"
)

var tlsRegKeyPath = `SYSTEM\CurrentControlSet\Control\SecurityProviders\SCHANNEL\Protocols\TLS 1.2\Client`

type ValidateKeys struct {
}

func (v ValidateKeys) ValidateSchUseStrongCryptoKeys(path string) (*int, error) {
	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		log.Debug("SchUseStrongCrypto Registry Key Check. Error opening SchUseStrongCrypto Registry Key. Error = ", err.Error())
		return nil, err
	}
	defer regKey.Close()

	schUseStrongCrypto, _, eErr := regKey.GetIntegerValue("SchUseStrongCrypto")
	if eErr != nil {
		return nil, err
	}
	n := int(schUseStrongCrypto)
	return &n, nil
}

func (v ValidateKeys) ValidateTLSRegKeys() (*entity.TLSRegKey, error) {
	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, tlsRegKeyPath, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		log.Debug("TLS Registry Key Check. Error opening TLS Registry Key. Error = ", err.Error())
		return nil, err
	}
	defer regKey.Close()

	enabled, _, eErr := regKey.GetIntegerValue("Enabled")
	if eErr != nil {
		return nil, err
	}

	disabledByDefault, _, dErr := regKey.GetIntegerValue("DisabledByDefault")
	if dErr != nil {
		return nil, err
	}

	return &entity.TLSRegKey{Enabled: int(enabled), DisabledByDefault: int(disabledByDefault)}, nil
}
