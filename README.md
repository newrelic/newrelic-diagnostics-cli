[![Community Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Project.png)](https://opensource.newrelic.com/oss-category/#community-project)

# Diagnostics CLI (nrdiag)

The Diagnostics CLI was built, and is maintained, by New Relic Global Technical Support.
It is a diagnostic and troubleshooting tool used to find common issues with New Relic Supported Products and Projects. Great for apps not reporting data to New Relic. Issues with connectivity, [configuration](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#h2-validate-your-config-file-settings), and compatibility. If it finds an issue it will suggest resolutions and relevant documentation. If it can’t help you resolve the issue, the information it gathers will help when troubleshooting with New Relic Support.

## Installation

**Diagnostics CLI  supports Linux, macOS, and Windows.**

To install in a Docker container see [here](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/diagnostics-cli-nrdiag/run-diagnostics-cli-nrdiag/#docker)

1. Download a zip of the [latest version here](http://download.newrelic.com/nrdiag/nrdiag_latest.zip), see changes in our [releases notes](https://docs.newrelic.com/docs/release-notes/diagnostics-release-notes/diagnostics-cli-release-notes/)
2. Extract the zip, and select the executable file for your OS.
3. Place the executable in your application's root directory. *(May not find all issues if located outside the apps root directory)*



## Usage
The Diagnostics CLI can help you troubleshoot issues and confirm that everything is working as expected. If after running the Diagnostics CLI you think you are still having issues review [Global Technical Support options](https://docs.newrelic.com/docs/licenses/license-information/general-usage-licenses/global-technical-support-offerings) that may be available for your issue, the Diagnostics CLI output can help us resolve issues faster so keep it around!
*Optional, but highly recommended*: Temporarily raise the logging level for the New Relic agent for more complete troubleshooting. Instructions can be found in [this documentation](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/generate-new-relic-agent-logs-troubleshooting)

 ### Troubleshooting and Diagnostics
 1. Open an elevated command prompt
 2. Navigate to root directory of your application (or where ever you placed the binary `nrdiag`)
 3. Run `nrdiag`
   * Target functionality by using cmd line args such as [suites to target specific products or issues](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#task-suites) or see [all cmd line args](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#cli-options)
4. Review results ([tips on interpreting output](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#interpret-output)).

### Working with Global Technical Support
If after running the Diagnostics CLI, reviewing the output, and attempting to resolve the issue you are still having difficulties understanding what the issue is, the data gathered by the Diagnostics CLI can be used by Global Technical Support to help resolve the issue, often in quicker time then without the data. Note, if you have fixed any issues called out by the Diagnostics CLI, either rerun it or let us know what you tried or changed (up to date results ensure more accurate troubleshooting).


If you have or are going to open a ticket with Global Technical Support, then [uploading the data gathered](https://docs.newrelic.com/docs/new-relic-solutions/solve-common-issues/diagnostics-cli-nrdiag/run-diagnostics-cli-nrdiag#attach-account-results) by the Diagnostics CLI as early as possible will speed up the troubleshooting process. Ensure to let the support engineer know that Diagnostics CLI results are available. 

To upload your results automatically to a New Relic Support ticket all you need to do is run `nrdiag` binary using the attachment flag `-a`. This uses a validated license key from your environment to upload the results to your New Relic account. The results can be viewed in the Diagnostics CLI output app [here](https://one.newrelic.com/diagnostics-cli-output).



## Support for the Diagnostics CLI

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub. If you have any questions, concerns or issues while running the Diagnostics CLI, reach out to us through our [Explorers Hub](https://discuss.newrelic.com/t/new-relic-diagnostic-aka-nr-diag/118819). Or if the issue has been confirmed as a bug or is a Feature request, please file a Github issue. Either way we'll get back to you soon!

## Contributing
Have you ever dealt with a New Relic installation and/or configuration issue? Do you have suggestions on how to automate those steps to diagnose and solve the issue? Then we would love for you to contribute to the Diagnostics CLI (or at least file a very detailed feature request)! The Diagnostics CLI's main goal is to speed up and simplify the resolution of issues, no matter if you are working on your own or with our Support teams.
All the information on how to build a new health check using the Diagnostics CLI, as well as the requirements to submit a PR, can be found in our docs directory. Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

Additionally, you'll have to fork this repo prior to cloning or you'll get a permissions error.

If you have any questions, or to execute our corporate CLA, required if your contribution is on behalf of a company, please drop us an email at opensource@newrelic.com.

### Recommended Doc reading order

1. [Contributing overview](./docs/Contributing.md)
2. [Coding Guidelines](./docs/Coding-Guidelines.md)
3. [Anatomy of a Task](./docs/Anatomy-of-a-Task.md)
4. [How To Build A Task](./docs/How-To-Build-A-Task.md)
5. [Testing Overview](./docs/Testing-Overview.md)
6. [Unit Testing](./docs/Unit-Testing.md)
7. [Dependency Injection](./docs/Dependency-Injection.md)
8. [Integration Testing](./docs/Integration-Testing.md)

## License

The Diagnostics CLI is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
The Diagnostics CLI also uses source code from third-party libraries. You can find full details on which libraries are used and the terms under which they are licensed in the third-party notices document.
