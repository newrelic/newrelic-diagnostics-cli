# Contributing

Contributions are always welcome. Before contributing please read the
[code of conduct](./../CODE_OF_CONDUCT.md) and [search the issue tracker](../../../issues); your issue may have already been discussed or fixed in `main`. To contribute,
[fork](https://help.github.com/articles/fork-a-repo/) this repository, commit your changes, and [send a Pull Request](https://help.github.com/articles/using-pull-requests/).

Note that our [code of conduct](./../CODE_OF_CONDUCT.md) applies to all platforms and venues related to this project; please follow it in all your interactions with the project and its participants.

## Feature Requests

Feature requests should be submitted in the [Issue tracker](../../../issues), with a description of the expected behavior & use case, where they’ll remain closed until sufficient interest, [e.g. :+1: reactions](https://help.github.com/articles/about-discussions-in-issues-and-pull-requests/), has been [shown by the community](../../issues?q=label%3A%22votes+needed%22+sort%3Areactions-%2B1-desc).
Before submitting an Issue, please search for similar ones in the
[closed issues](../../../issues?q=is%3Aissue+is%3Aclosed+label%3Aenhancement).

## Pull Requests

1. Ensure any install or build dependencies are removed before the end of the layer when doing a build.
2. Increase the version numbers in any examples files and the README.md to the new version that this Pull Request would represent. The versioning scheme we use is [SemVer](http://semver.org/).
3. You may merge the Pull Request in once you have the sign-off of two other developers, or if you do not have permission to do that, you may request the second reviewer to merge it for you.

## Contributor License Agreement

Keep in mind that when you submit your Pull Request, you'll need to sign the CLA via the click-through using CLA-Assistant. If you'd like to execute our corporate CLA, or if you have any questions, please drop us an email at opensource@newrelic.com.

For more information about CLAs, please check out Alex Russell’s excellent post,
[“Why Do I Need to Sign This?”](https://infrequently.org/2008/06/why-do-i-need-to-sign-this/).

## Slack

We host a public Slack with a dedicated channel for contributors and maintainers of open source projects hosted by New Relic.  If you are contributing to this project, you're welcome to request access to the #oss-contributors channel in the newrelicusers.slack.com workspace.  To request access, see https://newrelicusers-signup.herokuapp.com/.

## Developing for NR diag

### Requirements

* Go 1.13.0+
* GNU Make
* git


### Guidance

If you have an idea for a new health check or any additional steps nrdiag should take to validate that a New Relic product has been configured correctly, then you should build an NR Diag task, it's easy! 

All the docs provided in this directory are meant to guide you through how to build an NR Diag task and how to write appropriate tests for it. Those docs are very thorough, and were written considering contributors of all levels of technical knowledge.

Besides the documentation itself, you can take a look at the files within the task directory, and soon you'll notice that each NR Diag task has a very clear pattern and structure that you can use as reference to build your task. 


### Testing your task
Before opening a PR, make sure all the test are passing.

To run all unit tests, you can run one of scripts (inside of our scripts directory): `build.sh`

To run integration test for Linux and Darwin, you can run another script named `integrationTest.sh`

You also want to make sure that, after your changes, you can still build our binary for different operating systems. To build binaries for all of them, you can run the script `build.sh`

Finally, if you wish to see your task in action, we suggest that, after building the binaries with your changes included, you take the binary for your OS and drop it in an app's directory that is using the New Relic product you want to troubleshoot for. Then run `./nrdiag` and you should see your task running! If it doesn't show up, make sure you have register your tasks as indicated in the `Anatomy-of-a-task.md` doc.


#### Go Version Support

We'll aim to support the latest supported release of Go, along with the
previous release.  This doesn't mean that building with an older version of Go
will not work, but we don't intend to support a Go version in this project that
is not supported by the larger Go community.  Please see the [Go
releases][go_releases] page for more details.

