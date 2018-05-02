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
#    5. Scripts input comes from environment variables, see ubiquity_*_envs
# ----------------------------------------------------------------

ubiquity_envs="in_UBIQUITY_IMAGE_AMD64 out_UBIQUITY_IMAGE_AMD64 in_UBIQUITY_IMAGE_PPC64LE out_UBIQUITY_IMAGE_PPC64LE out_UBIQUITY_IMAGE_MULTIARCH"
ubiquity_db_envs="in_UBIQUITY_DB_IMAGE_AMD64 out_UBIQUITY_DB_IMAGE_AMD64 in_UBIQUITY_DB_IMAGE_PPC64LE out_UBIQUITY_DB_IMAGE_PPC64LE out_UBIQUITY_DB_IMAGE_MULTIARCH"
ubiquity_provisioner_envs="in_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64 out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64 in_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH"
ubiquity_flex_envs="in_UBIQUITY_IMAGE_K8S_FLEX_AMD64 out_UBIQUITY_K8S_FLEX_IMAGE_AMD64 in_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH"

# Validations
[ -f ./helper_to_push_internal_images_2hub.sh -a -f ./helper_to_push_manifest.sh ]  && : || exit 1
for expected_env in $ubiquity_envs $ubiquity_db_envs $ubiquity_provisioner_envs $ubiquity_flex_envs; do
   [ -z "`printenv $expected_env`" ] && { echo "Error: expected env [$expected_env] does not exist. Please set it first."; exit 1; } || :
   echo "$expected_env=`printenv $expected_env`"
done

echo ""
echo "Start to push Ubiquity..."
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_IMAGE_AMD64                 $out_UBIQUITY_IMAGE_AMD64
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_IMAGE_PPC64LE               $out_UBIQUITY_IMAGE_PPC64LE
./helper_to_push_manifest.sh $out_UBIQUITY_IMAGE_MULTIARCH   $out_UBIQUITY_IMAGE_AMD64  $out_UBIQUITY_IMAGE_PPC64LE

echo ""
echo "Start to push Ubiquity DB :"
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_DB_IMAGE_AMD64                $out_UBIQUITY_DB_IMAGE_AMD64
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_DB_IMAGE_PPC64LE                 $out_UBIQUITY_DB_IMAGE_PPC64LE
./helper_to_push_manifest.sh $out_UBIQUITY_DB_IMAGE_MULTIARCH   $out_UBIQUITY_DB_IMAGE_AMD64  $out_UBIQUITY_DB_IMAGE_PPC64LE


echo ""
echo "Start to push Ubiquity provisioner :"
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64                 $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE               $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE
./helper_to_push_manifest.sh $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH   $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64  $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE

echo ""
echo "Start to push Ubiquity flex :"
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_K8S_FLEX_IMAGE_AMD64                 $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64
./helper_to_push_internal_images_2hub.sh $in_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE               $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE
./helper_to_push_manifest.sh $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH   $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64  $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE



echo "######################################"
echo "Finish to push successfully all images"
echo "######################################"


echo $out_UBIQUITY_IMAGE_MULTIARCH
echo $out_UBIQUITY_DB_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH


