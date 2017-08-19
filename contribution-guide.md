# Contribution guide
The Ubiquity team accepts contributions from IBM employees using GitHub pull requests.
Pushing to master is not allowed. Any direct push to master might interfere with our CI process.

If you want to make a change, create your own branch out of `dev` branch, make your changes. Once you are done, submit a pull request to the `dev` branch. When accepted, the changes will make their way into `master` after we merge.

Verify your changes before submitting a pull request by running the unit, integration and acceptance tests. See the testing section for details. In addition, make sure that your changes are covered by existing or new unit testing.

# Build prerequisites
  * Install [golang](https://golang.org/) (>=1.6).
  * Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).
  * Install gcc.
  * Configure go. GOPATH environment variable must be set correctly before starting the build process. Create a new directory and set it as GOPATH.

### Download and build source code
* Configure ssh-keys for github.com. go tools require passwordless ssh access to github. If you have not set up ssh keys for your github profile, follow these [instructions](https://help.github.com/enterprise/2.7/user/articles/generating-an-ssh-key/) before you proceed. 
* Build Ubiquity service from source. 
```bash
mkdir -p $HOME/workspace
export GOPATH=$HOME/workspace
mkdir -p $GOPATH/src/github.com/IBM
cd $GOPATH/src/github.com/IBM
git clone git@github.com:IBM/ubiquity.git
cd ubiquity
./scripts/build
```

# Testing Ubiquity

Install these go packages to test Ubiquity:
```bash
# Install ginkgo
go install github.com/onsi/ginkgo/ginkgo
# Install gomega
go install github.com/onsi/gomega
```
Run the tests:
```bash
./scripts/run-unit-tests
```

# Running Ubiquity
You can run Ubiquity as a [root](README.md) or non-root user. Follow the instructions below for running Ubiquity with non-root user:

* Create user and group named 'ubiquity'.

```bash
adduser ubiquity
```

* Modify the sudoers file to provide the user and group 'ubiquity' with sufficient privileges.

```bash
## Entries for Ubiquity
ubiquity ALL= NOPASSWD: /usr/lpp/mmfs/bin/, /usr/bin/, /bin/
Defaults:%ubiquity !requiretty
Defaults:%ubiquity secure_path = /sbin:/bin:/usr/sbin:/usr/bin:/usr/lpp/mmfs/bin
```

# Integration tests 
If you have Kubernetes and/or Docker environment together with an active storage backend, you can run integration and acceptance tests, detailed in ubiquity-k8s and ubiquity-docker-plugin.


# Squash and merge

Upon the merge (by either you or your reviewer), all commits on the review branch must represent meaningful milestones or units of work. Use commit message to detail the development and review process.
Before merging a PR, squash any fix review feedback, typo, and rebased sorts of commits.
