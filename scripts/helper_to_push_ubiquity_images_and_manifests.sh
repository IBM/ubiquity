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
#    4. Helper scripts should be accessible: ./helper_to_push_docker_image.sh and ./helper_to_push_docker_manifest.sh
#    5. Scripts input comes from environment variables, see ubiquity_*_envs and optional TAG_LATEST
# ----------------------------------------------------------------

ubiquity_envs="in_UBIQUITY_IMAGE_AMD64 out_UBIQUITY_IMAGE_AMD64 in_UBIQUITY_IMAGE_PPC64LE out_UBIQUITY_IMAGE_PPC64LE out_UBIQUITY_IMAGE_MULTIARCH"
ubiquity_db_envs="in_UBIQUITY_DB_IMAGE_AMD64 out_UBIQUITY_DB_IMAGE_AMD64 in_UBIQUITY_DB_IMAGE_PPC64LE out_UBIQUITY_DB_IMAGE_PPC64LE out_UBIQUITY_DB_IMAGE_MULTIARCH"
ubiquity_provisioner_envs="in_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64 out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64 in_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH"
ubiquity_flex_envs="in_UBIQUITY_K8S_FLEX_IMAGE_AMD64 out_UBIQUITY_K8S_FLEX_IMAGE_AMD64 in_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH"

HELPER_PUSH_IMAGE=./helper_to_push_docker_image.sh
HELPER_PUSH_MANIFEST=./helper_to_push_docker_manifest.sh

date
# Validations
[ -f $HELPER_PUSH_IMAGE -a -f $HELPER_PUSH_MANIFEST ]  && : || exit 1
for expected_env in $ubiquity_envs $ubiquity_db_envs $ubiquity_provisioner_envs $ubiquity_flex_envs; do
   [ -z "`printenv $expected_env`" ] && { echo "Error: expected env [$expected_env] does not exist. Please set it first."; exit 1; } || :
   echo "$expected_env=`printenv $expected_env`"
done

echo "TAG_LATEST=$TAG_LATEST"

echo ""
echo "Start to push Ubiquity images and manifests..."
$HELPER_PUSH_IMAGE $in_UBIQUITY_IMAGE_AMD64                 $out_UBIQUITY_IMAGE_AMD64         $TAG_LATEST
$HELPER_PUSH_IMAGE $in_UBIQUITY_IMAGE_PPC64LE               $out_UBIQUITY_IMAGE_PPC64LE       $TAG_LATEST
$HELPER_PUSH_MANIFEST $out_UBIQUITY_IMAGE_MULTIARCH   $out_UBIQUITY_IMAGE_AMD64  $out_UBIQUITY_IMAGE_PPC64LE
if [ -n "$TAG_LATEST" ]; then
    latest_external_image=`echo $out_UBIQUITY_IMAGE_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$TAG_LATEST|"` # replace tag with $TAG_LATEST
    $HELPER_PUSH_MANIFEST $latest_external_image   $out_UBIQUITY_IMAGE_AMD64  $out_UBIQUITY_IMAGE_PPC64LE no
fi

echo ""
echo "Start to push Ubiquity DB images and manifests..."
$HELPER_PUSH_IMAGE $in_UBIQUITY_DB_IMAGE_AMD64                $out_UBIQUITY_DB_IMAGE_AMD64             $TAG_LATEST
$HELPER_PUSH_IMAGE $in_UBIQUITY_DB_IMAGE_PPC64LE              $out_UBIQUITY_DB_IMAGE_PPC64LE           $TAG_LATEST
$HELPER_PUSH_MANIFEST $out_UBIQUITY_DB_IMAGE_MULTIARCH   $out_UBIQUITY_DB_IMAGE_AMD64  $out_UBIQUITY_DB_IMAGE_PPC64LE $TAG_LATEST
if [ -n "$TAG_LATEST" ]; then
    latest_external_image=`echo $out_UBIQUITY_DB_IMAGE_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$TAG_LATEST|"` # replace tag with $TAG_LATEST
    $HELPER_PUSH_MANIFEST $latest_external_image   $out_UBIQUITY_DB_IMAGE_AMD64  $out_UBIQUITY_DB_IMAGE_PPC64LE no
fi


echo ""
echo "Start to push Ubiquity provisioner images and manifests..."
$HELPER_PUSH_IMAGE $in_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64                 $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64        $TAG_LATEST
$HELPER_PUSH_IMAGE $in_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE               $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE      $TAG_LATEST
$HELPER_PUSH_MANIFEST $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH   $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64  $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE  $TAG_LATEST
if [ -n "$TAG_LATEST" ]; then
    latest_external_image=`echo $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$TAG_LATEST|"` # replace tag with $TAG_LATEST
    $HELPER_PUSH_MANIFEST $latest_external_image   $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64  $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE no
fi


echo ""
echo "Start to push Ubiquity flex images and manifests..."
$HELPER_PUSH_IMAGE $in_UBIQUITY_K8S_FLEX_IMAGE_AMD64                 $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64     $TAG_LATEST
$HELPER_PUSH_IMAGE $in_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE               $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE   $TAG_LATEST
$HELPER_PUSH_MANIFEST $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH   $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64  $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE   $TAG_LATEST
if [ -n "$TAG_LATEST" ]; then
    latest_external_image=`echo $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$TAG_LATEST|"` # replace tag with $TAG_LATEST
    $HELPER_PUSH_MANIFEST $latest_external_image   $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64  $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE  no
fi


date
echo "######################################"
echo "Finish to push successfully all images"
echo "######################################"


echo $out_UBIQUITY_IMAGE_MULTIARCH
echo $out_UBIQUITY_DB_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH


