#!/bin/bash
# This script is to be called from ./build.sh before building the image.
# It backs up the install yaml files and updates the VERSION as defined in
# ./build.sh.

# backup the kappnav install yaml files
rm -rf backup
mkdir backup
cp kappnav.yaml backup/kappnav.yaml
cp kappnav-delete.yaml backup/kappnav-delete.yaml
cp kappnav-delete-CR.yaml backup/kappnav-delete-CR.yaml

# update the version numbers in the kappnav install yaml files
. ../build/version.sh

cat backup/kappnav.yaml \
| sed "s|KAPPNAV_VERSION|$VERSION|g" \
> kappnav.yaml

cat backup/kappnav-delete.yaml \
| sed "s|KAPPNAV_VERSION|$VERSION|g" \
> kappnav-delete.yaml

cat backup/kappnav-delete-CR.yaml \
| sed "s|KAPPNAV_VERSION|$VERSION|g" \
> kappnav-delete-CR.yaml
