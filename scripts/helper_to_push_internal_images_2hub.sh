#!/bin/bash -xe

# This script is for internal use in CI
# Assume docker login to the external registry was done in advance.

function usage()
{
    echo $0 [internal-image-path]  [external-image-path]
    exit 1
}

[ $# != 2 ] && { usage; }

internal_image=$1
external_image=$2

echo "1. Validate no external_image exist yet before pushing it."
docker pull $external_image && { echo "Error : the $external_image exist in remote. Cannot overwrite it."; exit 1; } || { echo "$external_image is not exist on the remote."; }
echo ""

echo "2. Validate internal_image not exist yet on local."
docker images $internal_image
docker rmi $internal_image && { echo "Remove the internal_image image to pull it again"; } || { echo "internal_image not exist on local. Continue."; }
echo ""

echo "3. Pull internal_image to local"
docker pull $internal_image
echo ""

echo "4. Tag internal_image to external_image and remove the internal_image"
docker tag  $internal_image $external_image
docker rmi $internal_image
docker push $external_image
echo ""

echo "5. Test pushed image by delete the local and pull it back"
docker rmi $external_image
docker pull $external_image
docker rmi $external_image

echo ""
echo "Succeeded to push [$internal_image]   --->   [$external_image]"
