#!/bin/bash -xe

# This script is for internal use in CI
# Assume docker login to the external registry was done in advance.

function usage()
{
    echo $0 [internal-image-path]  [external-image-path] [latest-tag-optional]
    exit 1
}

[ $# -eq 2 -o $# -eq 3 ] || { usage; }

internal_image=$1
external_image=$2
latest="$3"

# if also latest, then define latest_external_image (replace the external tag with the $latest wanted tag)
if [ -n "$latest" ]; then
   latest_external_image=`echo $external_image | sed "s|^\(.*/.*:\)\(.*\)$|\1$latest|"`
   latest_str_msg=" And latest tag image [$latest_external_image]"
fi

echo "Preparing to push internal_image --> external_image:"
echo "   internal_image=[$internal_image]"
echo "   external_image=[$external_image]"
[ -n "$latest" ] && echo "   latest_image  =[$latest_external_image]"


echo "1. Validate no external_image exist yet before pushing it."  # Note: no need to test latest tag since its already exist
docker pull $external_image && { echo "Error : the $external_image exist in remote. Cannot overwrite it."; exit 1; } || { echo "$external_image is not exist on the remote."; }
echo ""


echo "2. Validate internal_image not exist yet on local."
docker images $internal_image
docker rmi $internal_image && { echo "Remove the internal_image image to pull it again"; } || { echo "internal_image not exist on local. Continue."; }
echo ""


echo "3. Pull internal_image to local"
docker pull $internal_image
echo ""


echo "4. Tag internal_image to external_image (and latest=[$latest_external_image]) and remove the internal_image"
docker tag  $internal_image $external_image
[ -n "$latest_external_image" ] && docker tag  $internal_image $latest_external_image
docker rmi $internal_image
docker push $external_image
[ -n "$latest_external_image" ] && docker push $latest_external_image
echo ""


echo "5. Test pushed image by delete the local and pull it back"
docker rmi $external_image
docker pull $external_image
docker rmi $external_image


if [ -n "$latest_external_image" ]; then
    echo "6. Test pushed [latest] image by delete the local and pull it back"
    docker rmi $latest_external_image
    docker pull $latest_external_image
    docker rmi $latest_external_image
fi

set +x
echo ""
echo "Succeeded to push internal_image --> external_image"
echo "   internal_image=[$internal_image]"
echo "   external_image=[$external_image]"
[ -n "$latest" ] && echo "   latest_image  =[$latest_external_image]"
set -x

