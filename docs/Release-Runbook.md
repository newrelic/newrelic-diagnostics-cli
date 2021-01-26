# Diagnostics CLI release runbook - Internal usage only

This assumes the main branch is in a state ready for release. 

## Steps for a new release

1. Smoke test each binary on the appropriate OS (Windows/Mac/Linux) by running ./build.sh on the latest version of main branch. For windows testing, it’s easiest to copy the binary over using a shared folder in VMWare.For Linux, use a Vagrant box or VM. For Mac, just use your New Relic work laptop.Focus on the new tasks that are included in this release. Run the binary on a realistic, not perfect app's environment (specific to Java, .NET or whichever agent you are testing for) and examine the nrdiag output for anything unexpected.

2. Draft the release notes for discuss.newrelic.com and for docs.newrelic.com

3. Create a branch to bump majorMinor version in releaseVersion.txt and open a PR.

4. After the PR gets merged, wait for all the GH workflow checks to have passed. Only then publish a new release(https://github.com/newrelic/newrelic-diagnostics-cli/releases/new) with this tag version format: v.major.minor_buildnumber. The "Release title" MUST only include the build number.

![release image](./images/release.png)

The published release will trigger a release process to deploy a new version of nrdiag in https://download.newrelic.com/nrdiag

5. Publish the drafted release notes.

## Steps for a rollback

1. Manually delete the release from Release page https://github.com/newrelic/newrelic-diagnostics-cli/releases . This will trigger the github action worflow titled rollback.yml

2. After the action has completed, you can verify in https://download.newrelic.com/nrdiag that the new release is not longer there and that the previous release is now the latest one (nrdiag_latest.zip).

3. Once you are ready to deploy the hotfix release, you'll have to repeat the steps outlined above to [create a new release.](#Steps-for-a-new-release)