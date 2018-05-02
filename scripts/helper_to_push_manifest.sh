#!/bin/bash -xe

# ----------------------------------------------------------------
# This script is for internal use in CI
# The script create, inspect and push new docker manifest.
# The script validates the whole process. If something gets wrong the script will fail with error.
# Pre-requisites:
#    1. Assume docker login to the external registry was already done.
#    2. Assume the multi architecture images were already pushed.
# ----------------------------------------------------------------

function abort()
{
    exitcode=$1
    msg=$2
    echo "Error: $msg"
    exit $exitcode
}

USAGE="Usage : $0 [manifest-path] [image-amd64-path] [image-ppc64le-path]"

[ $# -ne 3 ] && abort 1 $USAGE

expected_manifest_arch_number=2
manifest=$1
image_amd64=$2
image_ppc64le=$3
manifest_dirname="~/.docker/manifests"
dockerhub_prefix_manifest="docker.io_"
convert_manifest_filename=`echo "$manifest" | sed -e 's|/|_|g' -e 's|:|-|g'`
specific_manifest_dirname="$manifest_dirname/${convert_manifest_filename}"
specific_manifest_dirname_with_prefix="$manifest_dirname/${dockerhub_prefix_manifest}${convert_manifest_filename}"
echo "1. Make sure architecture images are not exist locally and if so remove them first for clean state..."
[ -n "`docker images -q $image_amd64`" ]   && { docker rmi $image_amd64   || abort 1 "fail to clean image before creating manifest. [$image_amd64]"; }   || :
[ -n "`docker images -q $image_ppc64le`" ] && { docker rmi $image_ppc64le || abort 1 "fail to clean image before creating manifest. [$image_ppc64le]"; } || :


echo "2. Manifest validation \(manifest should not exit, not local nor remote\)..."
ls -ld "$specific_manifest_dirname $specific_manifest_dirname_with_prefix" 	&& abort 1 "local manifest dir should NOT exist before creating the manifest. Please clean it and rerun."
docker manifest inspect $manifest 	&& abort 1 "manifest inspect should NOT exist before pushing it. Please clean it and rerun." || :


echo "3. Manifest creation and push..."
docker manifest create $manifest ${image_amd64} ${image_ppc64le} 	|| abort 2 "fail to create manifest."
docker manifest inspect $manifest 	                                ||   abort 2 "fail to inspect local manifest."
actual_manifest_arch_number=`docker manifest inspect $manifest | grep architecture | wc -l`
[ $actual_manifest_arch_number -ne $expected_manifest_arch_number ] && abort 3 "Manifest created but its not contain [$expected_manifest_arch_number] architectures as expected."
docker manifest push --purge $manifest 	||  abort 2 "fail to push manifest to remote repo"
ls -ld $specific_manifest_dirname $specific_manifest_dirname_with_prefix       && abort 2 "Local manifest file should NOT exist after successful manifest push. Please check." || :



echo "4. Remote manifest validation..."
docker manifest inspect $manifest 	    ||   abort 3 "fail to inspect remote manifest."
docker pull $manifest                   ||   abort 3 "fail pull remote manifest."
expected_arch=`uname -m`
docker run -it --rm $manifest uname -m | grep $expected_arch || abort 3 "The manifest run did not bring the expected arch"
docker rmi $manifest  # just remove the local manifest that was pulled for testing
actual_manifest_arch_number=`docker manifest inspect $manifest | grep architecture | wc -l`
[ $actual_manifest_arch_number -ne $expected_manifest_arch_number ] && abort 3 "Manifest pushed but its not contain [$expected_manifest_arch_number] architectures as expected."



echo "================================================================================"
echo "Successfully created and pushed a manifest [$manifest] with [$expected_arch] architectures. "

