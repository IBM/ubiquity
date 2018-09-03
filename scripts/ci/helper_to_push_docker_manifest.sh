#!/bin/bash -xe

# ----------------------------------------------------------------
# This script is for internal use in CI
# The script create, inspect and push new docker manifest.
# The script validates the whole process. If something gets wrong the script will fail with error.
# Pre-requisites:
#    1. Run docker login to the external registry in advance.
#    2. The internal images should be exist in advance.
#    3. The designated manifest should NOT be exist (the script will create it).
#    4. "docker manifest" is enabled.
#    5. Optional: set environment MANIFEST_FLAG with dedicated flags for docker manifest commands.
#       For example to push docker manifest for internal registry without certificates, one can use: export MANIFEST_FLAG="--insecure"
# ----------------------------------------------------------------

function abort()
{
    exitcode=$1
    msg=$2
    echo "Error: $msg"
    exit $exitcode
}

USAGE="Usage : $0 [manifest-path] [image-amd64-path] [image-ppc64le-path] [image-s390x-path] [fail-if-exist-optional]"

[ $# -eq 4 -o $# -eq 5 ] || abort 1 $USAGE

expected_manifest_arch_number=3 # X, P and Z
manifest=$1
image_amd64=$2
image_ppc64le=$3
image_s390x=$4

fail_if_exist="${5:-true}"  # by default fail if external manifest already exist

echo "Preparing to create and push manifest for image_amd64, image_ppc64le and image_s390x:"
echo "   manifest=[$manifest]"
echo "   image_amd64=[$image_amd64]"
echo "   image_ppc64le=[$image_ppc64le]"
echo "   image_s390x=[$image_s390x]"


manifest_dirname="~/.docker/manifests"
dockerhub_prefix_manifest="docker.io_"
convert_manifest_filename=`echo "$manifest" | sed -e 's|/|_|g' -e 's|:|-|g'`
specific_manifest_dirname="$manifest_dirname/${convert_manifest_filename}"
specific_manifest_dirname_with_prefix="$manifest_dirname/${dockerhub_prefix_manifest}${convert_manifest_filename}"
echo "1. Make sure architecture images are not exist locally and if so remove them first for clean state..."
[ -n "`docker images -q $image_amd64`" ]   && { docker rmi $image_amd64   || abort 1 "fail to clean image before creating manifest. [$image_amd64]"; }   || :
[ -n "`docker images -q $image_ppc64le`" ] && { docker rmi $image_ppc64le || abort 1 "fail to clean image before creating manifest. [$image_ppc64le]"; } || :
[ -n "`docker images -q $image_s390x`" ] && { docker rmi $image_s390x || abort 1 "fail to clean image before creating manifest. [$image_s390x]"; } || :


echo "2. Manifest validation \(manifest should not exit, not local nor remote\)..."
ls -ld "$specific_manifest_dirname $specific_manifest_dirname_with_prefix" 	&& abort 1 "local manifest dir should NOT exist before creating the manifest. Please clean it and rerun."
if [ "$fail_if_exist" = "true" ]; then
    docker manifest inspect $MANIFEST_FLAG $manifest 	&& abort 1 "manifest inspect should NOT exist before pushing it. Please clean it and rerun." || :
else
    docker manifest inspect $MANIFEST_FLAG $manifest 	&& echo "manifest inspect should NOT exist before pushing it. BUT fail_if_exist is not true so skip validation." || :
fi

echo "3. Manifest creation and push..."
docker manifest create $MANIFEST_FLAG $manifest ${image_amd64} ${image_ppc64le} ${image_s390x}	|| abort 2 "fail to create manifest."
docker manifest inspect $MANIFEST_FLAG $manifest 	                                ||   abort 2 "fail to inspect local manifest."
actual_manifest_arch_number=`docker manifest inspect $MANIFEST_FLAG $manifest | grep architecture | wc -l`
[ $actual_manifest_arch_number -ne $expected_manifest_arch_number ] && abort 3 "Manifest created but its not contain [$expected_manifest_arch_number] architectures as expected."
docker manifest push $MANIFEST_FLAG --purge $manifest 	||  abort 2 "fail to push manifest to remote repo"
ls -ld $specific_manifest_dirname $specific_manifest_dirname_with_prefix       && abort 2 "Local manifest file should NOT exist after successful manifest push. Please check." || :



echo "4. Remote manifest validation..."
docker manifest inspect $MANIFEST_FLAG $manifest 	    ||   abort 3 "fail to inspect remote manifest."
docker pull $manifest                   ||   abort 3 "fail pull remote manifest."
expected_arch=`uname -m`
docker run  --rm $manifest uname -m | grep $expected_arch || abort 3 "The manifest run did not bring the expected arch"
docker rmi $manifest  # just remove the local manifest that was pulled for testing
actual_manifest_arch_number=`docker manifest inspect $MANIFEST_FLAG $manifest | grep architecture | wc -l`
[ $actual_manifest_arch_number -ne $expected_manifest_arch_number ] && abort 3 "Manifest pushed but its not contain [$expected_manifest_arch_number] architectures as expected."


set +x
echo "================================================================================"
echo "Succeeded to create and push manifest for image_amd64, image_ppc64le and image_s390x:"
echo "   manifest=[$manifest]"
echo "   image_amd64=[$image_amd64]"
echo "   image_ppc64le=[$image_ppc64le]"
echo "   image_s390x=[$image_s390x]"
set -x

