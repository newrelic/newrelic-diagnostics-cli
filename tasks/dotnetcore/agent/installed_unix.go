//go:build linux || darwin
// +build linux darwin

package agent

var DotNetCoreAgentPaths = []string{
	"/usr/local/newrelic-netcore20-agent/",
	"/usr/local/newrelic-dotnet-agent",
}
