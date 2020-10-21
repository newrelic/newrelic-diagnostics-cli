[![Community Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Project.png)](https://opensource.newrelic.com/oss-category/#community-project)

# NR Diag

NR Diag was built, and is maintained, by New Relic Global Technical Support. 
It is a diagnostic and troubleshooting tool used to find common issues with New Relic Supported Products and Projects. Great for apps not reporting data to New Relic. Issues with connectivity, [configuration](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#h2-validate-your-config-file-settings), and compatibility. If it finds an issue it will suggest resolutions and relevant documentation. If it can’t help you resolve the issue, the information it gathers will help when troubleshooting with New Relic Support.

## Installation

**NR Diag supports Linux, macOS, and Windows.**

To install in a Docker container see [here](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#h2-run-new-relic-diagnostics-in-a-docker-container)

1. Download a zip of the latest version and see changes in our [releases notes](https://docs.newrelic.com/docs/release-notes/platform-release-notes/diagnostics-release-notes)
2. Extract the zip, and select the executable file for your OS. 
3. Place the executable in your application's root directory. *(May not find all issues if located outside the apps root directory)*



## Usage
NR Diag can help you troubleshoot issues and confirm that everything is working as expected. If after running NR Diag you think you are still having issues review [Global Technical Support options](https://docs.newrelic.com/docs/licenses/license-information/general-usage-licenses/global-technical-support-offerings) that may be available for your issue, NR Diag output can help us resolve issues faster so keep it around!
*Optional, but highly recommended*: Temporarily raise the logging level for the New Relic agent for more complete troubleshooting. Instructions can be found in [this documentation](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/generate-new-relic-agent-logs-troubleshooting)

 ### Troubleshooting and Diagnostics 
 1. Open an elevated command prompt 
 2. Navigate to root directory of your application (or where ever you placed the NR Diag binary) 
 3. Run `nrdiag` 
   * Target functionality by using cmd line args such as [suites to target specific products or issues](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#task-suites) or see [all cmd line argos](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#cli-options)
4. Review results( [tips on interpreting output](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#interpret-output)). 

### Working with Global Technical Support
If after running NR Diag, reviewing the output, and attempting to resolve the issue you are still having difficulties understanding what the issue is, the data gathered by NR Diag can be used by Global Technical Support to help resolve the issue, often in quicker time then without the data. Note if you have fixed any issues called out by NR Diag, either rerun it or let us know what you tried or changed (up to date results ensure more accurate troubleshooting) 
If you have or are going to open a ticket with Global Technical Support, then [uploading the data gathered](https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/troubleshooting/new-relic-diagnostics#attach-ticket-results) by NR Diag will as early as possible will speed up the troubleshooting process.
To upload your results automatically to a New Relic Support ticket all you need to do is run nrdiag binary using the attachment flag `-a MY-ATTACHMENT-KEY` You can get your attachment key by viewing the ticket in the New Relic Support Portal. You can also request a support engineer to provide you the attachment key.



## Support for NR Diag

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub. If you have any questions, concerns or issues while running NR Diag, reach out to us through our [Explorers Hub](https://discuss.newrelic.com/t/new-relic-diagnostic-aka-nr-diag/118819). Or if the issue has been confirmed as a bug or is a Feature request, please file a Github issue. Either way we'll get back to you soon!

## Contributing
Have you ever dealt with a New Relic installation and/or configuration issue? Do you have suggestions on how to automate those steps to diagnose and solve the issue? Then we would love for you to contribute to NR Diag (or at least file a very detailed feature request)! NR Diag's main goal is to speed up and simplify the resolution of issues, no matter if you are working on your own or with our Support teams.
All the information on how to build a new health check using NR Diag, as well as the requirements to submit a PR, can be found in our docs directory. Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project. If you have any questions, or to execute our corporate CLA, required if your contribution is on behalf of a company, please drop us an email at opensource@newrelic.com.

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
[NR Diag] is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
>[[NR Diag] also uses source code from third-party libraries. You can find full details on which libraries are used and the terms under which they are licensed in the third-party notices document.]
