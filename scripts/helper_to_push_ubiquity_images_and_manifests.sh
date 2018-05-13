#!/bin/bash -ex

# ----------------------------------------------------------------
# This script is for internal use in CI
# The script push all ubiquity images from internal registry to external registry.
# Images for amd64 and ppc64le for each ubiquity image : ubiquity, ubiquity-db, flex and provisioner.
# It also creates and pushes relevant manifests per architecture into the external repository.
# The script validates the whole process. If something gets wrong the script will fail with error.
# Pre-requisites:
#    1. Run docker login to the external registry in advance.
#    2. The internal images should be exist in advance.
#    3. The external images should NOT be exist (the script will creates them).
#    4. Helper scripts should be accessible: ./helper_to_push_internal_images_2hub.sh and ./helper_to_push_manifest.sh
#    5. Scripts input comes from environment variables, see ubiquity_*_envs and optional TAG_LATEST
# ----------------------------------------------------------------

ubiquity_envs="in_UBIQUITY_IMAGE_AMD64 out_UBIQUITY_IMAGE_AMD64 in_UBIQUITY_IMAGE_PPC64LE out_UBIQUITY_IMAGE_PPC64LE out_UBIQUITY_IMAGE_MULTIARCH"
ubiquity_db_envs="in_UBIQUITY_DB_IMAGE_AMD64 out_UBIQUITY_DB_IMAGE_AMD64 in_UBIQUITY_DB_IMAGE_PPC64LE out_UBIQUITY_DB_IMAGE_PPC64LE out_UBIQUITY_DB_IMAGE_MULTIARCH"
ubiquity_provisioner_envs="in_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64 out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64 in_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH"
ubiquity_flex_envs="in_UBIQUITY_K8S_FLEX_IMAGE_AMD64 out_UBIQUITY_K8S_FLEX_IMAGE_AMD64 in_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH"

date
# Validations
[ -f ./helper_to_push_internal_images_2hub.sh -a -f ./helper_to_push_manifest.sh ]  && : || exit 1
for expected_env in $ubiquity_envs $ubiquity_db_envs $ubiquity_provisioner_envs $ubiquity_flex_envs; do
   [ -z "`printenv $expected_env`" ] && { echo "Error: expected env [$expected_env] does not exist. Please set it first."; exit 1; } || :
   echo "$expected_env=`printenv $expected_env`"
done

echo "TAG_LATEST=$TAG_LATEST"

echo ""
echo "Start to push Ubiquity images and manifests..."
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_IMAGE_AMD64                 $out_UBIQUITY_IMAGE_AMD64         $TAG_LATEST
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_IMAGE_PPC64LE               $out_UBIQUITY_IMAGE_PPC64LE       $TAG_LATEST
./helper_to_push_manifest.sh $out_UBIQUITY_IMAGE_MULTIARCH   $out_UBIQUITY_IMAGE_AMD64  $out_UBIQUITY_IMAGE_PPC64LE
if [ -n "$TAG_LATEST" ]; then
    latest_external_image=`echo $out_UBIQUITY_IMAGE_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$TAG_LATEST|"` # replace tag with $TAG_LATEST
    ./helper_to_push_manifest.sh $latest_external_image   $out_UBIQUITY_IMAGE_AMD64  $out_UBIQUITY_IMAGE_PPC64LE
fi

echo ""
echo "Start to push Ubiquity DB images and manifests..."
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_DB_IMAGE_AMD64                $out_UBIQUITY_DB_IMAGE_AMD64             $TAG_LATEST
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_DB_IMAGE_PPC64LE              $out_UBIQUITY_DB_IMAGE_PPC64LE           $TAG_LATEST
./helper_to_push_manifest.sh $out_UBIQUITY_DB_IMAGE_MULTIARCH   $out_UBIQUITY_DB_IMAGE_AMD64  $out_UBIQUITY_DB_IMAGE_PPC64LE $TAG_LATEST
if [ -n "$TAG_LATEST" ]; then
    latest_external_image=`echo $out_UBIQUITY_DB_IMAGE_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$TAG_LATEST|"` # replace tag with $TAG_LATEST
    ./helper_to_push_manifest.sh $latest_external_image   $out_UBIQUITY_DB_IMAGE_AMD64  $out_UBIQUITY_DB_IMAGE_PPC64LE $TAG_LATEST
fi


echo ""
echo "Start to push Ubiquity provisioner images and manifests..."
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64                 $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64        $TAG_LATEST
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE               $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE      $TAG_LATEST
./helper_to_push_manifest.sh $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH   $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64  $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE  $TAG_LATEST
if [ -n "$TAG_LATEST" ]; then
    latest_external_image=`echo $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$TAG_LATEST|"` # replace tag with $TAG_LATEST
    ./helper_to_push_manifest.sh $latest_external_image   $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64  $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE  $TAG_LATEST
fi


echo ""
echo "Start to push Ubiquity flex images and manifests..."
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_K8S_FLEX_IMAGE_AMD64                 $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64     $TAG_LATEST
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE               $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE   $TAG_LATEST
./helper_to_push_manifest.sh $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH   $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64  $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE   $TAG_LATEST
if [ -n "$TAG_LATEST" ]; then
    latest_external_image=`echo $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$TAG_LATEST|"` # replace tag with $TAG_LATEST
    ./helper_to_push_manifest.sh $latest_external_image   $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64  $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE   $TAG_LATEST
fi


date
echo "######################################"
echo "Finish to push successfully all images"
echo "######################################"


echo $out_UBIQUITY_IMAGE_MULTIARCH
echo $out_UBIQUITY_DB_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH

