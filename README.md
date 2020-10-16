[![Community Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Project.png)](https://opensource.newrelic.com/oss-category/#community-project)

# NR Diag

[NR Diag](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics) is a troubleshooting tool used to find common issues with New Relic installation and configuration. If you don't see your app reporting to New Relic, then run NR Diag to find out what went wrong in your app's environment configuration. If no issue is found, NR Diag can still collect relevant files and data that the Support team will use to figure out the issue.

## Installation

1. Download the latest version from [our releases notes](https://docs.newrelic.com/docs/release-notes/platform-release-notes/diagnostics-release-notes)

2. Once you extract the zip, it will contain executable files for Linux, macOS, and Windows. Select the one for your OS.

3. Place the executable into the location of your application's root directory.

4. Optional, but highly recommended: Temporarily raise the logging level for the New Relic agent for more accurate troubleshooting. Note that changing the logging level requires you to restart your application.

5. Run the executable 
You can simply do `./nrdiag` and it will attempt to troubleshoot for all New Relic agents. Or you can `./nrdiag -suites java`, for example, to troubleshoot for an specific agent, the Java Agent. To find out what is the suite name provided to your New Relic agent run `./nrdiag --help suites`

[Here you can find](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#cli-options) more details and CLI options to run NR Diag


## Ticket Usage
If after running NR Diag, you still have difficulties understading what the issue is, and/or you have an open ticket with the Support department, then you can upload the data gathered by NR Diag to speed up the troubleshooting process.

To upload your results automatically to a New Relic Support ticket all you need to do is run nrdiag using the attachment flag like this: `./nrdiag -a MY-ATTACHMENT-KEY`

You can get your attachment key by viewing the ticket in the [New Relic Support Portal](https://support.newrelic.com). You can also request a support engineer to provide you the attachment key.


## Support for NR Diag

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the [New Relic Explorers Hub](https://discuss.newrelic.com/). If you have any questions, concerns or issues while running NR Diag, reach out to us through our Explorers Hub. We'll get back to you soon!

## Contributing
Have you ever dealt with a New Relic installation and/or configuration issue? Do you have suggestions on how to automate those steps to diagnose and solve the issue? Then you must contribute to NR Diag! 
NR diag's main goal is to offer customers self-service for technical support so they do not have to go through the trouble of contacting New Relic's support department. If ever wish you did not have to reach out to Support, but still find assistance to troubleshoot by yourself, then we share the same goals :) 

All the information on how to build a new health check for NR diag, as well as the requirements to submit a PR, can be found in our docs directory.
Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.
If you have any questions, or to execute our corporate CLA, required if your contribution is on behalf of a company, please drop us an email at opensource@newrelic.com.

## License
[Project Name] is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
>[If applicable: The [project name] also uses source code from third-party libraries. You can find full details on which libraries are used and the terms under which they are licensed in the third-party notices document.]
