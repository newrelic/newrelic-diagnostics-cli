package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// PHPEnvPHPinfoCLI - This struct defined the sample plugin which can be used as a starting point
type PHPEnvPHPinfoCLI struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p PHPEnvPHPinfoCLI) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("PHP/Env/PHPinfoCLI")
}

// Explain - Returns the help text for each individual task
func (p PHPEnvPHPinfoCLI) Explain() string {
	return "Collect PHP CLI configuration"
}

// Dependencies - Returns the dependencies for ech task.
func (p PHPEnvPHPinfoCLI) Dependencies() []string {
	return []string{"PHP/Config/Agent"}
}

// Execute - The core work within each task
func (p PHPEnvPHPinfoCLI) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "PHP Agent was not detected on this host. Skipping PHP info check.",
	}
	if upstream["PHP/Config/Agent"].Status != tasks.Success {
		return result
	}

	//Running PHP info

	gatheredOutput, err := p.gatherPHPInfoCLI()

	if err != nil {
		result.Status = tasks.Error
		result.Summary = "error executing PHP -i"
		return result
	}

	result.Status = tasks.Success
	result.Summary = "PHP info has been gathered"
	result.Payload = gatheredOutput

	return result
}

func (p PHPEnvPHPinfoCLI) gatherPHPInfoCLI() (string, error) {
	outputByteSlice, err := p.cmdExec("php", "-i")

	return string(outputByteSlice), err
}
