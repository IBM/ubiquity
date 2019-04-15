set -x
set -e
export PATH=$PATH:/usr/local/go/bin:$WORKSPACE/work/bin

pwd
ls
PWDO=`pwd`
export GOPATH=`pwd`/work
rm -rf work/src/github.com/IBM/ubiquity || :
mkdir -p work/src/github.com/IBM/ubiquity
mv * work/src/github.com/IBM/ubiquity || :
cd work/src/github.com/IBM/ubiquity
pwd
ls


echo "------- Docker image build and push - Start"

branch=`echo $GIT_BRANCH| sed 's|/|.|g'`  #not sure if docker accept / in the version
specific_tag="${IMAGE_VERSION}_b${BUILD_NUMBER}_${branch}"

if [ "$GIT_BRANCH" = "dev" -o "$GIT_BRANCH" = "origin/dev" -o "$GIT_BRANCH" = "master" -o "$to_tag_latest_also_none_dev_branches" = "true" ]; then
   tag_latest="true"
   echo "will tag latest \ version in addition to the branch tag $GIT_BRANCH"
else
   tag_latest="false"
   echo "NO latest \ version tag for you $GIT_BRANCH"
fi

echo "build ubiquity image"
echo "===================="
ubiquity_registry="${DOCKER_REGISTRY}/${UBIQUITY_IMAGE}"
ubiquity_tag_specific="${ubiquity_registry}:${specific_tag}"
ubiquity_tag_latest=${ubiquity_registry}:latest
ubiquity_tag_version=${ubiquity_registry}:${IMAGE_VERSION}
[ "$tag_latest" = "true" ] && taglatestflag="-t ${ubiquity_tag_latest} -t ${ubiquity_tag_version}" || taglatestflag=""
# Build and tags togather
docker build -t ${ubiquity_tag_specific} ${taglatestflag} -f Dockerfile .

# push the tags
docker push ${ubiquity_tag_specific}
[ "$tag_latest" = "true" ] && docker push ${ubiquity_tag_latest} || :
[ "$tag_latest" = "true" ] && docker push ${ubiquity_tag_version} || :

echo "build ubiquity-db image"
echo "======================="
ubiquity_db_registry="${DOCKER_REGISTRY}/${UBIQUITY_DB_IMAGE}"
ubiquity_db_tag_specific="${ubiquity_db_registry}:${specific_tag}"
ubiquity_db_tag_latest=${ubiquity_db_registry}:latest
ubiquity_db_tag_version=${ubiquity_db_registry}:${IMAGE_VERSION}

[ "$tag_latest" = "true" ] && taglatestflag="-t ${ubiquity_db_tag_latest} -t ${ubiquity_db_tag_version}" || taglatestflag=""
cp LICENSE ./scripts/ubiquity-postgresql
cd ./scripts/ubiquity-postgresql
# Build and tags togather
docker build -t ${ubiquity_db_tag_specific} ${taglatestflag} -f Dockerfile .

# push the tags
docker push ${ubiquity_db_tag_specific}
[ "$tag_latest" = "true" ] && docker push ${ubiquity_db_tag_latest} || :
[ "$tag_latest" = "true" ] && docker push ${ubiquity_db_tag_version} || :

cd -

echo "------- Docker image build and push - Done"



cd $PWDO


echo "============================="
echo "ubiquity server IMAGE name : "
echo "   specific tag : ${ubiquity_tag_specific}"
[ "$tag_latest" = "true" ] && echo "   latest \ version tag   : ${ubiquity_tag_latest}       ${ubiquity_tag_version}" || echo "no latest tag"
echo "============================="
echo "ubiquity-db IMAGE name : "
echo "   specific tag : ${ubiquity_db_tag_specific}"
[ "$tag_latest" = "true" ] && echo "   latest \ version tag   : ${ubiquity_db_tag_latest}       ${ubiquity_db_tag_version}"  || echo "no latest tag"

echo ${ubiquity_tag_specific} > ubiquity_tags
echo ${ubiquity_db_tag_specific} >> ubiquity_tags
