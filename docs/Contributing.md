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

1. Fork this repo and clone your forked version of this repo
2. When you create branch, name it after your Github Handle followed by a slash, and assign it a descriptive name of the changes you will make. Examples:
`git branch DanaScully/update-notempirical-docs`
`git branch fMulder/create-BaseEnvFindTruth-task`
`git branch TheLoneGunmen/issue29-fixfalsepositive` (assuming this contributor is addressing one of our github issues)
3. Once you open a Pull Request, sign the License Agreement which will show up in the UI as one of our checks. We'll talk in more detail about this in the next section.
4. After the PR checks have ran, if you notice the banner `this branch is out-of-date with the base branch`, get the latest changes from upstream into your origin branch (you can find some recommended steps on how to do this at the bottom of this doc.)
5. You may merge the Pull Request in once you have the sign-off of two other developers, or if you do not have permission to do that, you may request the second reviewer to merge it for you.

## Contributor License Agreement

Keep in mind that when you submit your Pull Request, you'll need to sign the CLA via the click-through using CLA-Assistant. If you'd like to execute our corporate CLA, or if you have any questions, please drop us an email at opensource@newrelic.com.

For more information about CLAs, please check out Alex Russell’s excellent post,
[“Why Do I Need to Sign This?”](https://infrequently.org/2008/06/why-do-i-need-to-sign-this/).

## Slack

We host a public Slack with a dedicated channel for contributors and maintainers of open source projects hosted by New Relic.  If you are contributing to this project, you're welcome to request access to the #oss-contributors channel in the newrelicusers.slack.com workspace.  To request access, see https://newrelicusers-signup.herokuapp.com/.

## Developing for the Diagnostics CLI

### Requirements

* Go 1.13.0+
* GNU Make
* git
* Docker
* Fork this repo

#### Go Version Support

We'll aim to support the latest supported release of Go, along with the
previous release (this project is using Go 1.14).  This doesn't mean that building with an older version of Go
will not work, but we don't intend to support a Go version in this project that
is not supported by the larger Go community.  Please see the [Go
releases](https://golang.org/doc/go1.14) page for more details.

### Guidance

If you have an idea for a new health check or any additional steps the Diagnostics CLI should take to validate that a New Relic product has been configured correctly, then you should build a task for the Diagnostics CLI, it's easy!

All the docs provided in this directory are meant to guide you through how to build a task for the Diagnostics CLI and how to write appropriate tests for the task. Those docs are very thorough, and were written considering contributors of all levels of technical knowledge.

Besides the documentation itself, you can take a look at the files within the task directory, and soon you'll notice that each task in the Diagnostics CLI has a very clear pattern and structure that you can use as reference to build your own task.

One important thing to keep in mind is that we have already written a lot of good, basic health checks. Please make sure that your idea for a health check has not yet been implemented in one way or another in our tasks directory. If you do not find it and you are ready to start building your task, then take advantage of the helper functions provided in our taskHelpers files inside the tasks directory. This is boiler plate logic that we found is applicable and useful to most New Relic health checks.

Additionally, take advantage of other Diagnostics CLI tasks to build your own task on top of them. Imagine you want to build a task to validate that a customer is using only Node.js supported versions for the Node Agent, then you could use another Diagnostics CLI task that already gathers the Node version from customer's environment. To get more details on how take advantage of `upstream` tasks, take a look at the [code snippets in our Coding Guidelines.](./Coding-Guidelines.md)

### Testing your task

Before opening a PR, make sure all the test are passing.

To run all unit tests, go to the root directory and run: `./scripts/test.sh`

To run integration test for Linux and Darwin, invoke this script: `./scripts/integrationTest.sh`

You also want to make sure that, after your changes, you can still build our binary for different operating systems. To build binaries for all of them, you can run the script `./scripts/build.sh`. You can find the binaries that were built inside the `bin/` directory

Finally, if you wish to see your task in action, we suggest that, after building the binaries with your changes included, you take the binary for your OS and drop it in an app's directory that is using the New Relic product you want to troubleshoot for. Then run `./nrdiag` (or whatever the name assigned to your binary is) and you should see your task in action! If it doesn't show up, make sure you have register your tasks as indicated in the `Anatomy-of-a-task.md` doc.

### Keeping your forked repo in-sync with our original remote repo

List the current configured remote repository for your fork: `git remote -v`

You'll only see:

```bash
origin  https://github.com/YOUR_USERNAME/YOUR_FORK.git (fetch)

origin  https://github.com/YOUR_USERNAME/YOUR_FORK.git (push)
```

Now specify a new remote upstream repository that will be synced with the fork: `git remote add upstream https://github.com/ORIGINAL_OWNER/ORIGINAL_REPOSITORY.git`

Verify the new upstream set for your fork:

```bash
$git remote -v

origin    https://github.com/YOUR_USERNAME/YOUR_FORK.git (fetch)

origin    https://github.com/YOUR_USERNAME/YOUR_FORK.git (push)

upstream  https://github.com/ORIGINAL_OWNER/ORIGINAL_REPOSITORY.git (fetch)

upstream  https://github.com/ORIGINAL_OWNER/ORIGINAL_REPOSITORY.git (push)
```

Bring the recent changes from upstream into your origin:

```bash
$git checkout main

$git fetch upstream main
```

Merge changes from upstream/main into your local main branch:

```bash
$git checkout myfeaturebranch

$git merge upstream/main
```