//go:build windows
// +build windows

package repository

import (
	"errors"

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
		return nil, errors.New("schUseStrongCrypto Parent Key does not exist. Consult these docs to enable SchUseStrongCrypto https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry")

	}
	defer regKey.Close()

	stat, s_Err := regKey.Stat()
	if s_Err != nil {
		return nil, s_Err
	}
	valueCount := stat.ValueCount
	if valueCount == 0 {
		return nil, errors.New("unable to find values for SchUseStrongCrypto Registry Key. Consult these docs to enable TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry")
	}
	valueNames, read_Err := regKey.ReadValueNames(int(valueCount))
	if read_Err != nil {
		return nil, read_Err
	}
	found := false

	for _, value := range valueNames {
		if value == "SchUseStrongCrypto" {
			found = true
		}

	}

	if !found {
		return nil, errors.New("unable to find values for SchUseStrongCrypto Registry Key. Consult these docs to enable TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry")
	}

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
		return nil, errors.New("TLS Registry Client Key does not exist. Consult these docs to enable TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry")
	}
	defer regKey.Close()

	stat, s_Err := regKey.Stat()
	if s_Err != nil {
		return nil, s_Err
	}
	valueCount := stat.ValueCount
	if valueCount == 0 {
		return nil, errors.New("unable to find values for TLS 1.2 Registry Key. Consult these docs to enable TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry")
	}
	valueNames, read_Err := regKey.ReadValueNames(int(valueCount))
	if read_Err != nil {
		return nil, read_Err
	}
	found_en := false
	found_dis := false

	for _, value := range valueNames {
		if value == "Enabled" {
			found_en = true
		}
		if value == "DisabledByDefault" {
			found_dis = true
		}
	}

	if !found_en || !found_dis {
		return nil, errors.New("unable to find values for TLS 1.2 Registry Key. Consult these docs to enable TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry")
	}

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
