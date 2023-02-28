//go:build windows
// +build windows

package profiler

import (
	"github.com/newrelic/newrelic-diagnostics-cli/domain/entity"
	"github.com/newrelic/newrelic-diagnostics-cli/domain/repository"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var SchCryptoRegKeyPath_1 = `SOFTWARE\WOW6432Node\Microsoft\.NETFramework\v4.0.30319`
var SchCryptoRegKeyPath_2 = `SOFTWARE\Microsoft\.NETFramework\v4.0.30319`

type DotNetTLSRegKey struct {
	name         string
	validateKeys repository.IValidateKeys
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
	schCryptoKey_1, schErr_1 := p.validateKeys.ValidateSchUseStrongCryptoKeys(SchCryptoRegKeyPath_1)
	if schErr_1 != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: schErr_1.Error(),
		}
	}
	schCryptoKey_2, schErr_2 := p.validateKeys.ValidateSchUseStrongCryptoKeys(SchCryptoRegKeyPath_2)
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
	tlsRegKeys, err := p.validateKeys.ValidateTLSRegKeys()
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: err.Error(),
		}
	}
	return p.compareTLSRegKeys(tlsRegKeys)
}

func (p DotNetTLSRegKey) compareTLSRegKeys(tlsRegKeys *entity.TLSRegKey) tasks.Result {
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
