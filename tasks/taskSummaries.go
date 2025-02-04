package tasks

// ThisProgramFullName is a constant for the name of the program to be used in task summaries
const ThisProgramFullName = "Diagnostics CLI"

// NotifyIssueSummary is the standardized message we can add to a tasks.Result summary to suggest users ways to reach out to us and notify about an nrdiag issue
const NotifyIssueSummary = "\nPlease notify this issue to us whenever possible through https://discuss.newrelic.com/ by creating a new topic or through https://github.com/newrelic/newrelic-diagnostics-cli/issues\n"

// AssertionErrorSummary is the standardized message we display to the user whenever Diagnostics CLI was unable to finish a task due a type assertion error
const AssertionErrorSummary = ThisProgramFullName + " was unable to complete this health check because we ran into an unexpected type assertion error." + NotifyIssueSummary

// NoAgentDetectedSummary is the standard tasks.None summary we want to display when nrdiag does not detect an agent after looking for its config file or other relevant ways of configuration
const NoAgentDetectedSummary = "New Relic configuration not detected for this specific agent at the location where Diagnostics CLI was ran. This will stop other health checks (targeting this agent) from running. If you are attempting to troubleshoot for this specific agent, re-run Diagnostics CLI using the command-line option '-config-file' to point to the path to this agent configuration file."

// NoAgentUpstreamSummary is the standard tasks.None summary that we display when a task did not run because the upstream task for detecting an agent failed at detecting the agent. Beware! this summary expects a string concatenation at the end
const NoAgentUpstreamSummary = "This task did not run because the following upstream task was unable to detect New Relic configuration for this agent: "

// UpstreamFailedSummary is the tasks.None summary that we display when we cannot run the current task because the previous one had some sort of failure. Beware! this summary expects a string concatenation at the end
const UpstreamFailedSummary = "This task did not run because the following upstream task will need to succeed before the current one can run: "
