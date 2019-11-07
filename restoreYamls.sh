#!/bin/bash
# This script is to be called from ./build.sh after building the image.
# It restores the original install yaml files with unresolved values
# for the kAppNav version.  The updated yaml files used to build the
# image are preserved in subdirectory updatedYaml.

# Preserve the updated files for possible inspection
rm -rf updatedYaml
mkdir updatedYaml
mv kappnav.yaml updatedYaml
mv kappnav-delete.yaml updatedYaml
mv kappnav-delete-CR.yaml updatedYaml

# Now that we've built the image and saved the updated files we can restore the original files
mv backup/kappnav.yaml kappnav.yaml
mv backup/kappnav-delete.yaml kappnav-delete.yaml
mv backup/kappnav-delete-CR.yaml kappnav-delete-CR.yaml
rmdir backup
