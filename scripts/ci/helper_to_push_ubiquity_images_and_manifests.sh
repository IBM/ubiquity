#!/bin/bash -ex

# ----------------------------------------------------------------
# This script is for internal use in CI
# The script push all ubiquity images from internal registry to external registry.
# Images for amd64, ppc64le and s390x for each ubiquity image : ubiquity, ubiquity-db, flex and provisioner.
# It also creates and pushes relevant manifests per architecture into the external repository.
# The script validates the whole process. If something gets wrong the script will fail with error.
# Pre-requisites:
#    1. Run docker login to the external registry in advance.
#    2. The internal images should be exist in advance.
#    3. The external images should NOT be exist (the script will creates them).
#    4. Helper scripts should be accessible: ./helper_to_push_docker_image.sh and ./helper_to_push_docker_manifest.sh
#    5. Scripts input comes from environment variables, see ubiquity_*_envs and optional TAG_LATEST
# ----------------------------------------------------------------

function push_arch_images_and_create_manifest_for_app()
{
    # This function push <app> arch images(X, P and Z) from internal registry to external registry and then create manifest for them on the external registry.
    # Function arguments :
    #   <app name>
    #   <in_app_image_AMD64>   <out_app_image_AMD64>
    #   <in_app_image_PPC64LE> <out_app_image_PPC64LE>
    #   <in_app_image_S390X>   <out_app_image_S390X>
    #   <out_app_image_MULTIARCH>
    #   <tag_LATEST>


    app_name="$1"
    in_app_image_AMD64="$2"
    out_app_image_AMD64="$3"
    in_app_image_PPC64LE="$4"
    out_app_image_PPC64LE="$5"
    in_app_image_S390X="$6"
    out_app_image_S390X="$7"
    out_app_image_MULTIARCH="$8"
    tag_LATEST="$9"

    echo ""
    echo "Start to push $app_name images and manifest..."
    $HELPER_PUSH_IMAGE $in_app_image_AMD64                 $out_app_image_AMD64         $tag_LATEST
    $HELPER_PUSH_IMAGE $in_app_image_PPC64LE               $out_app_image_PPC64LE       $tag_LATEST
    $HELPER_PUSH_IMAGE $in_app_image_S390X                 $out_app_image_S390X         $tag_LATEST
    $HELPER_PUSH_MANIFEST $out_app_image_MULTIARCH   $out_app_image_AMD64  $out_app_image_PPC64LE  $out_app_image_S390X
    if [ -n "$tag_LATEST" ]; then
        latest_external_image=`echo $out_app_image_MULTIARCH | sed "s|^\(.*/.*:\)\(.*\)$|\1$tag_LATEST|"` # replace tag with $tag_LATEST
        $HELPER_PUSH_MANIFEST $latest_external_image   $out_app_image_AMD64  $out_app_image_PPC64LE  $out_app_image_S390X no
    fi
}

ubiquity_envs="in_UBIQUITY_IMAGE_AMD64 out_UBIQUITY_IMAGE_AMD64 in_UBIQUITY_IMAGE_PPC64LE out_UBIQUITY_IMAGE_PPC64LE in_UBIQUITY_IMAGE_S390X out_UBIQUITY_IMAGE_S390X out_UBIQUITY_IMAGE_MULTIARCH"
ubiquity_db_envs="in_UBIQUITY_DB_IMAGE_AMD64 out_UBIQUITY_DB_IMAGE_AMD64 in_UBIQUITY_DB_IMAGE_PPC64LE out_UBIQUITY_DB_IMAGE_PPC64LE in_UBIQUITY_DB_IMAGE_S390X out_UBIQUITY_DB_IMAGE_S390X out_UBIQUITY_DB_IMAGE_MULTIARCH"
ubiquity_provisioner_envs="in_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64 out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64 in_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE in_UBIQUITY_K8S_PROVISIONER_IMAGE_S390X out_UBIQUITY_K8S_PROVISIONER_IMAGE_S390X out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH"
ubiquity_flex_envs="in_UBIQUITY_K8S_FLEX_IMAGE_AMD64 out_UBIQUITY_K8S_FLEX_IMAGE_AMD64 in_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE in_UBIQUITY_K8S_FLEX_IMAGE_S390X out_UBIQUITY_K8S_FLEX_IMAGE_S390X out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH"
ubiquity_flex_sidecar_envs="in_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_AMD64 out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_AMD64 in_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_PPC64LE out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_PPC64LE in_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_S390X out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_S390X out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_MULTIARCH"
ubiquity_helm_utils_envs="in_UBIQUITY_K8S_HELM_UTILS_IMAGE_AMD64 out_UBIQUITY_K8S_HELM_UTILS_IMAGE_AMD64 in_UBIQUITY_K8S_HELM_UTILS_IMAGE_PPC64LE out_UBIQUITY_K8S_HELM_UTILS_IMAGE_PPC64LE in_UBIQUITY_K8S_HELM_UTILS_IMAGE_S390X out_UBIQUITY_K8S_HELM_UTILS_IMAGE_S390X out_UBIQUITY_K8S_HELM_UTILS_IMAGE_MULTIARCH"

HELPER_PUSH_IMAGE=./helper_to_push_docker_image.sh
HELPER_PUSH_MANIFEST=./helper_to_push_docker_manifest.sh

date
# Validations
[ -f $HELPER_PUSH_IMAGE -a -f $HELPER_PUSH_MANIFEST ]  && : || exit 1
for expected_env in $ubiquity_envs $ubiquity_db_envs $ubiquity_provisioner_envs $ubiquity_flex_envs $ubiquity_flex_sidecar_envs $ubiquity_helm_utils_envs; do
   [ -z "`printenv $expected_env`" ] && { echo "Error: expected env [$expected_env] does not exist. Please set it first."; exit 1; } || :
   echo "$expected_env=`printenv $expected_env`"
done

echo "TAG_LATEST=$TAG_LATEST"


push_arch_images_and_create_manifest_for_app "Ubiquity"               $in_UBIQUITY_IMAGE_AMD64                   $out_UBIQUITY_IMAGE_AMD64                          $in_UBIQUITY_IMAGE_PPC64LE                   $out_UBIQUITY_IMAGE_PPC64LE                          $in_UBIQUITY_IMAGE_S390X                   $out_UBIQUITY_IMAGE_S390X                          $out_UBIQUITY_IMAGE_MULTIARCH                    $TAG_LATEST
push_arch_images_and_create_manifest_for_app "Ubiquity DB"            $in_UBIQUITY_DB_IMAGE_AMD64                $out_UBIQUITY_DB_IMAGE_AMD64                       $in_UBIQUITY_DB_IMAGE_PPC64LE                $out_UBIQUITY_DB_IMAGE_PPC64LE                       $in_UBIQUITY_DB_IMAGE_S390X                $out_UBIQUITY_DB_IMAGE_S390X                       $out_UBIQUITY_DB_IMAGE_MULTIARCH                 $TAG_LATEST
push_arch_images_and_create_manifest_for_app "Ubiquity provisioner"   $in_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64   $out_UBIQUITY_K8S_PROVISIONER_IMAGE_AMD64          $in_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE   $out_UBIQUITY_K8S_PROVISIONER_IMAGE_PPC64LE          $in_UBIQUITY_K8S_PROVISIONER_IMAGE_S390X   $out_UBIQUITY_K8S_PROVISIONER_IMAGE_S390X          $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH    $TAG_LATEST
push_arch_images_and_create_manifest_for_app "Ubiquity flex"          $in_UBIQUITY_K8S_FLEX_IMAGE_AMD64          $out_UBIQUITY_K8S_FLEX_IMAGE_AMD64                 $in_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE          $out_UBIQUITY_K8S_FLEX_IMAGE_PPC64LE                 $in_UBIQUITY_K8S_FLEX_IMAGE_S390X          $out_UBIQUITY_K8S_FLEX_IMAGE_S390X                 $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH           $TAG_LATEST
push_arch_images_and_create_manifest_for_app "Ubiquity flex sidecar"  $in_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_AMD64  $out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_AMD64         $in_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_PPC64LE  $out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_PPC64LE         $in_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_S390X  $out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_S390X         $out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_MULTIARCH   $TAG_LATEST
push_arch_images_and_create_manifest_for_app "Ubiquity helm utils"    $in_UBIQUITY_K8S_HELM_UTILS_IMAGE_AMD64    $out_UBIQUITY_K8S_HELM_UTILS_IMAGE_AMD64           $in_UBIQUITY_K8S_HELM_UTILS_IMAGE_PPC64LE    $out_UBIQUITY_K8S_HELM_UTILS_IMAGE_PPC64LE           $in_UBIQUITY_K8S_HELM_UTILS_IMAGE_S390X    $out_UBIQUITY_K8S_HELM_UTILS_IMAGE_S390X           $out_UBIQUITY_K8S_HELM_UTILS_IMAGE_MULTIARCH     $TAG_LATEST

date
echo "######################################"
echo "Finish to push successfully all images"
echo "######################################"


echo $out_UBIQUITY_IMAGE_MULTIARCH
echo $out_UBIQUITY_DB_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_PROVISIONER_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_FLEX_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_FLEX_SIDECAR_IMAGE_MULTIARCH
echo $out_UBIQUITY_K8S_HELM_UTILS_IMAGE_MULTIARCH


