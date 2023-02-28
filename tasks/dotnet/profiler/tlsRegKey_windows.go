//go:build windows
// +build windows

package profiler

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"golang.org/x/sys/windows/registry"
)

var tlsRegKeyPath = `SYSTEM\CurrentControlSet\Control\SecurityProviders\SCHANNEL\Protocols\TLS 1.2\Client`
var SchCryptoRegKeyPath_1 = `SOFTWARE\WOW6432Node\Microsoft\.NETFramework\v4.0.30319`
var SchCryptoRegKeyPath_2 = `SOFTWARE\Microsoft\.NETFramework\v4.0.30319`

type DotNetTLSRegKey struct {
	name                           string
	validateTLSRegKeys             func() (*TLSRegKey, error)
	validateSchUseStrongCryptoKeys func(string) (*int, error)
}

type TLSRegKey struct {
	Enabled           int
	DisabledByDefault int
}

func (p DotNetTLSRegKey) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Profiler/TLSRegKey")
}

func (p DotNetTLSRegKey) Explain() string {
	return "Validate at least one version of TLS is enabled: Required by .NET"
}

func (p DotNetTLSRegKey) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

func (p DotNetTLSRegKey) Execute(op tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		if upstream["DotNet/Agent/Installed"].Summary == tasks.NoAgentDetectedSummary {
			return tasks.Result{
				Status:  tasks.None,
				Summary: tasks.NoAgentUpstreamSummary + "DotNet/Agent/Installed",
			}
		}
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.UpstreamFailedSummary + "DotNet/Agent/Installed",
		}
	}
	schCryptoKey_1, schErr_1 := p.validateSchUseStrongCryptoKeys(SchCryptoRegKeyPath_1)
	if schErr_1 != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: schErr_1.Error(),
		}
	}
	schCryptoKey_2, schErr_2 := p.validateSchUseStrongCryptoKeys(SchCryptoRegKeyPath_2)
	if schErr_2 != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: schErr_1.Error(),
		}
	}
	if *schCryptoKey_1 != 1 || *schCryptoKey_2 != 1 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "SchUseStrongCrypto must be enabled.  See more in these docs https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry",
		}
	}
	tlsRegKeys, err := p.validateTLSRegKeys()
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: err.Error(),
		}
	}
	return p.compareTLSRegKeys(tlsRegKeys)
}

func ValidateSchUseStrongCryptoKeys(path string) (*int, error) {
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

func ValidateTLSRegKeys() (*TLSRegKey, error) {
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

	return &TLSRegKey{Enabled: int(enabled), DisabledByDefault: int(disabledByDefault)}, nil
}

func (p DotNetTLSRegKey) compareTLSRegKeys(tlsRegKeys *TLSRegKey) tasks.Result {
	if tlsRegKeys.Enabled == 1 && tlsRegKeys.DisabledByDefault == 0 {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: "You have TLS 1.2 enabled",
		}
	} else if tlsRegKeys.Enabled == 0 && tlsRegKeys.DisabledByDefault == 1 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "You have disabled TLS 1.2. Consult these docs to enable TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry",
		}
	}
	return tasks.Result{
		Status:  tasks.Warning,
		Summary: "Check your registry settings and consider enabling TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry",
	}
}
