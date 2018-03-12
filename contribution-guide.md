# Contribution guide
The Ubiquity team accepts contributions from IBM employees using GitHub pull requests.
Pushing to master is not allowed. Any direct push to master might interfere with our CI process.

If you want to make a change, create your own branch out of `dev` branch, make your changes. Once you are done, submit a pull request to the `dev` branch. When accepted, the changes will make their way into `master` after we merge.

Verify your changes before submitting a pull request by running the unit, integration and acceptance tests. See the testing section for details. In addition, make sure that your changes are covered by existing or new unit testing.

# Build prerequisites
  * Install [golang](https://golang.org/) (>=1.9.1).
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


# Squash and merge

Upon the merge (by either you or your reviewer), all commits on the review branch must represent meaningful milestones or units of work. Use commit message to detail the development and review process.
Before merging a PR, squash any fix review feedback, typo, and rebased sorts of commits.
