# Contribution guide
The Ubiquity team uses GitHub and accepts contributions via pull requests.
Pushing to master is not allowed for any reason, any direct push to master might brake the CI/CD process that we are following.

If you wish to make a change to Ubiquity, create your own branch out of `dev` branch, make your changes, once you are done submit a pull request to the `dev` branche. Once accepted, those changes should make their way into `master` when we merge.

To verify your changes before submitting a pull request, run unit tests, the integration test suite, and the acceptance Tests. See the testing section for more detail.

If your changes are related to a bug fix, it should be realtively easy to review since test coverage is submitted with the patch. Bug fixes don't usually require alot of extra testing: But please update the unit tests so that they catch the bug.

Testing Ubiquity:

In order to be able to test ubiquity you need to install these  go packages:
```bash
# Install ginkgo
go install github.com/onsi/ginkgo/ginkgo
# Install gomega
go install github.com/onsi/gomega
```
Once you've followed the steps above to install ginkgo and omega needed for testing, execute the following script to run all unit tests in ubiquity:

```bash
./scripts/run-unit-tests
```


# Integration tests (In progress)
If you have a running kubernetes and/or docker environment set up, as well as a running storage backend, you can also run integration and acceptance tests that you can find in
ubiquity-k8s and/or ubiquity-docker-plugin.


# Squash and Merge

Upon merge (by either you or your reviewer), all commits left on the review branch should represent meaningful milestones or units of work. Use commits to add clarity to the development and review process.
Before merging a PR, squash any fix review feedback, typo, and rebased sorts of commits.
