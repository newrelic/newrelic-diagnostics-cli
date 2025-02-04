package w3wp

import (
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/shirou/w32"
)

type inspectResults struct {
	Pid          int32
	modulesFound map[string]bool
}

type DotNetW3wpValidate struct {
	name string
}

func (p DotNetW3wpValidate) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/W3wp/Validate")
}

func (p DotNetW3wpValidate) Explain() string {
	return "Check if W3wp processes are being profiled by the New Relic .NET agent" //This is the customer visible help text that describes what this particular task does
}

func (p DotNetW3wpValidate) Dependencies() []string {
	return []string{
		// Need the w3wp pids
		"DotNet/W3wp/Collect",
		"Base/Env/CheckWindowsAdmin",
	}
}

func (p DotNetW3wpValidate) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	// set up the return var
	result := tasks.Result{}
	failures := 0
	successes := 0
	modsToCheck := [...]string{"NewRelic.Profiler.dll", "mscoree.dll"}
	adminPerms := true // assume admin permissions, set to false later if needed

	// get results of Base/Env/CheckWindowsAdmin to see what permissions we have
	if upstream["Base/Env/CheckWindowsAdmin"].Status == tasks.Warning {
		adminPerms = false // no admin permissions
	}

	if upstream["DotNet/W3wp/Collect"].Status == tasks.None {
		result.Status = tasks.None
		return result
	}
	// get pids from DotNet/W3wp/Collect and make sure they are type []Process
	w3wpProcesses, ok := upstream["DotNet/W3wp/Collect"].Payload.([]process.Process)
	if !ok {
		logger.Debug("The payload from the w3wp collection is not the correct type! This usually means there were no w3wp processes running.")
		result.Status = tasks.None
		return result
	}

	checkResults := make([]inspectResults, len(w3wpProcesses))

	for w3wpKey, w3wp := range w3wpProcesses {
		checkResults[w3wpKey].Pid = w3wp.Pid
		checkResults[w3wpKey].modulesFound = make(map[string]bool)

		logger.Debug("Checking w3wp with pid: ", w3wp.Pid)

		// initialize checkResults.modulesFound to false for each module we are looking for
		// then set to true later if we find the module
		for _, modCheck := range modsToCheck {
			checkResults[w3wpKey].modulesFound[modCheck] = false
		}

		// get a handle for the process
		handle := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPMODULE, uint32(w3wp.Pid))

		// ensure we got a valid handle
		if handle == 0 {
			logger.Debug("Failed to open a handle to the worker process with pid: ", w3wp.Pid)
			failures++
			if !adminPerms {
				result.Summary += "\n Failed to open a handle to the worker process with pid: " + strconv.FormatInt(int64(w3wp.Pid), 10) + "\n - This is possibly due to permissions. If possible re-run from an Admin cmd prompt or PowerShell."
			} else {
				result.Summary += "\n Failed to open a handle to the worker process with pid: " + strconv.FormatInt(int64(w3wp.Pid), 10)
			}
			continue
		}

		// Make sure we close the handle when done
		defer w32.CloseHandle(handle)

		// variable for module information
		var modEntry w32.MODULEENTRY32
		modEntry.Size = uint32(unsafe.Sizeof(modEntry))

		// tries to get the first module's information and handle failure
		if !w32.Module32First(handle, &modEntry) {
			logger.Debug("Unable to enumerate the modules for the worker process with pid: ", w3wp.Pid)
			failures++
			if !adminPerms {
				result.Summary += "\n Unable to enumerate the modules for the worker process with pid: " + strconv.FormatInt(int64(w3wp.Pid), 10) + "\n - This is possibly due to permissions. If possible re-run from an Admin cmd prompt or PowerShell."
			} else {
				result.Summary += "\n Unable to enumerate the modules for the worker process with pid: " + strconv.FormatInt(int64(w3wp.Pid), 10)
			}
			// go to next w3wp process
			continue
		}

		for _, modCheck := range modsToCheck {
			if checkModuleByName(modCheck, &modEntry) {
				logger.Debug("Found ", modCheck, " attached to the worker process with pid: ", w3wp.Pid)
				successes++
				checkResults[w3wpKey].modulesFound[modCheck] = true
			}
		}

		// loop through rest of modules and keep checking against modsToCheck list
		for w32.Module32Next(handle, &modEntry) {
			for _, modCheck := range modsToCheck {
				if checkModuleByName(modCheck, &modEntry) {
					logger.Debug("Found ", modCheck, " attached to the worker process with pid: ", w3wp.Pid)
					successes++
					checkResults[w3wpKey].modulesFound[modCheck] = true
				}
			}
		}

		// check the modulesFound variable to make sure each module was found
		for modKey, modFound := range checkResults[w3wpKey].modulesFound {
			if modFound == false {
				failures++
				if !adminPerms {
					result.Summary += "\n Failed to find " + modKey + " attached to the worker process with pid: " + strconv.FormatInt(int64(w3wp.Pid), 10) + "\n - This is possibly due to permissions. If possible re-run from an Admin cmd prompt or PowerShell."
				} else {
					result.Summary += "\n Failed to find " + modKey + " attached to the worker process with pid: " + strconv.FormatInt(int64(w3wp.Pid), 10)
				}
				logger.Debug("Failed to find ", modKey, " attached to the worker process with pid: ", w3wp.Pid)
			}
		}
	}

	// checks failures and successes and sets result.Status to match
	if failures == 0 && successes == 0 {
		// no failures and no successes, guess there were no worker processes
		logger.Debug("No worker processes were checked.")
	} else if failures == 0 && successes > 0 {
		// there were no failures and some successes, return Success
		result.Status = tasks.Success
	} else if successes == 0 && failures > 0 {
		// there were no successes and some failures, return Failure
		result.Status = tasks.Failure
		if !adminPerms {
			result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/new-relic-diagnostics#windows-run"
		} else {
			result.URL = "https://docs.newrelic.com/docs/agents/net-agent/configuration/net-agent-configuration"
		}
	} else if successes > 0 && failures > 0 {
		// there were some successes and some failures, return Warning
		result.Status = tasks.Warning
		if !adminPerms {
			result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/new-relic-diagnostics#windows-run"
		} else {
			result.URL = "https://docs.newrelic.com/docs/agents/net-agent/configuration/net-agent-configuration"
		}
	}

	return result
}

// checkModuleByName - returns true if module name matches string (case insensitive)
func checkModuleByName(name string, modEntry *w32.MODULEENTRY32) bool {
	modName := syscall.UTF16ToString(modEntry.SzModule[:])
	return strings.EqualFold(name, modName)
}
