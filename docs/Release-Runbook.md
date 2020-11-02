# NR Diag release runbook - Internal usage only

This assumes the main branch is in a state ready for release. 

## Updates docs

1. Smoke test each binary on the appropriate OS (Windows/Mac/Linux) by running ./build.sh on the latest version of main branch.

For windows testing, it’s easiest to copy the binary over using a shared folder in VMWare.
For Linux, use a Vagrant box or VM
For Mac, just use your New Relic work laptop.

Focus on the new tasks that are included in this release. Run the binary on a realistic, not perfect app's environment (specific to Java, .NET or whichever agent you are testing for) and examine the nrdiag output for anything unexpected.

2. Draft the release notes for discuss.newrelic.com and for docs.newrelic.com

3. Create a branch to bump version in majorMinorVersion.txt and open a PR. Right after and prior to requesting a PR review, push a tag that features the new build. Example:

Create the tag `$git tag build/32`
Push the tag `git push origin build/32`

Once this tag gets merged into the main branch, it will trigger our release process implemented on .github/workflows/release.yml.

4. Publish the drafted release notes.

