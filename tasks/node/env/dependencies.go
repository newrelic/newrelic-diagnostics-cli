package env

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type NodeEnvDependencies struct {
	cmdExec tasks.CmdExecFunc
}

type NodeModuleVersion struct {
	Module  string
	Version string
}

func (p NodeEnvDependencies) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Env/Dependencies")
}

func (p NodeEnvDependencies) Explain() string {
	return "Collect Nodejs application dependencies"
}

func (p NodeEnvDependencies) Dependencies() []string {
	return []string{
		"Node/Env/NpmVersion",
		"Node/Config/Agent",
	}
}

func (p NodeEnvDependencies) Execute(option tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Node/Env/NpmVersion"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "NPM is not installed. This task did not run",
		}
	}

	if upstream["Node/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Node agent config file not detected. This task did not run",
		}
	}

	modulesList, npmErr := p.getModulesListStr()
	// create a channel to stream modulesList and zip file with tasks.FileCopyEnvelope
	stream := make(chan string)
	//start go routine
	go streamSource(modulesList, stream)

	filesToCopy := []tasks.FileCopyEnvelope{{Path: "npm_ls_output.txt", Stream: stream}}
	//The npm error exit status 1 should be an expected error
	//if npm throws the famous npm ERR!, those messages are long. I rather not concatenate that output in the Summary, but still collect the output in a txt file to study the error
	if npmErr != nil && (npmErr.Error() != "exit status 1") {
		return tasks.Result{
			Status:      tasks.Error,
			Summary:     npmErr.Error() + ": npm threw an error while running the command npm ls --depth=0 --parseable=true --long=true. Please verify that the " + tasks.ThisProgramFullName + " is running in your Node application directory. Possible causes for npm errors: https://docs.npmjs.com/common-errors. The output of 'npm ls' is used by Support Engineers to find out if your application is using unsupported technologies.",
			FilesToCopy: filesToCopy,
		}
	}

	//returns a slice of NodeModuleVersion struct which can be used for dependency injection later
	NodeModulesVersions := p.getNodeModulesVersions(modulesList)

	if len(NodeModulesVersions) < 1 {
		return tasks.Result{
			Status:      tasks.Error,
			Summary:     "We failed to parse the output of npm ls, but have included it in nrdiag-output.zip. The output of 'npm ls' is used by Support Engineers to find out if your application is using unsupported technologies.",
			FilesToCopy: filesToCopy,
		}
	}

	return tasks.Result{
		Status:      tasks.Info,
		Summary:     "We have successfully retrieved a list of dependencies from your node_modules folder",
		Payload:     NodeModulesVersions,
		FilesToCopy: filesToCopy,
	}
}

func (p NodeEnvDependencies) getModulesListStr() (string, error) {
	cmdOutput, cmdError := p.cmdExec("npm", "ls", "--parseable=true", "--long=true", "--depth=0")
	modulesList := string(cmdOutput)
	if cmdError != nil {
		return modulesList, cmdError
	}
	return modulesList, nil
}

func streamSource(input string, ch chan string) {

	defer close(ch)

	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ch <- scanner.Text() + "\n"
	}
}

func (p NodeEnvDependencies) getNodeModulesVersions(modulesList string) []NodeModuleVersion {
	//Example of output:
	///Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/express:express@4.16.4:undefined
	///Users/shuayhuaca/Desktop/projects/node/nannynow/server/node_modules/mongoose:mongoose@5.4.0:undefined

	var modulesVersions []NodeModuleVersion
	modulesSlice := strings.Split(modulesList, "\n")

	for _, line := range modulesSlice {

		if strings.Contains(line, "npm ERR!") {
			continue
		}
		regex := regexp.MustCompile(`:([\S]+)@([0-9.]+)`)
		result := regex.FindStringSubmatch(line)
		var (
			moduleName    string
			moduleVersion string
		)
		if len(result) < 3 {
			continue
		}
		moduleName, moduleVersion = result[1], result[2]
		dependencyInfo := NodeModuleVersion{
			Module:  moduleName,
			Version: moduleVersion,
		}
		modulesVersions = append(modulesVersions, dependencyInfo)
	}

	return modulesVersions
}
