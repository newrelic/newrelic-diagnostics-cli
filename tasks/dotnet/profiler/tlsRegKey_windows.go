//go:build windows
// +build windows

package profiler

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"golang.org/x/sys/windows/registry"
)

var tlsRegKeyPath = `SYSTEM\CurrentControlSet\Control\SecurityProviders\SCHANNEL\Protocols\TLS 1.0\Client`

type DotNetProfilerTLSRegKey struct {
	name string
}

type TLSRegKey struct {
	Enabled           int
	DisabledByDefault int
}

func (p DotNetProfilerTLSRegKey) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Profiler/TLSRegKey")
}

func (p DotNetProfilerTLSRegKey) Explain() string {
	return "Validate at least one version of TLS is enabled: Required by .NET"
}

func (p DotNetProfilerTLSRegKey) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

func (p DotNetProfilerTLSRegKey) Execute(op tasks.Options, upstream map[string]tasks.Result) tasks.Result {
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
	tlsRegKeys, err := validateTLSRegKeys()
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: err.Error(),
		}
	}
	return compareTLSRegKeys(tlsRegKeys)
}

func validateTLSRegKeys() (*TLSRegKey, error) {
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

func compareTLSRegKeys(tlsRegKeys *TLSRegKey) tasks.Result {
	if tlsRegKeys.Enabled == 0 && tlsRegKeys.DisabledByDefault == 1 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "You have disabled TLS 1.0. When this is disabled, no data appears to New Relic. Consider enabling TLS 1.2",
		}
	} else if tlsRegKeys.Enabled == 1 && tlsRegKeys.DisabledByDefault == 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "You have enabled TLS 1.0. TLS 1.0/1.1 have been EOL'd so consider enabling TLS 1.2",
		}
	}
	return tasks.Result{
		Status:  tasks.Error,
		Summary: "Check your setting for TLS 1.0 and consider enabling TLS 1.2",
	}
}
