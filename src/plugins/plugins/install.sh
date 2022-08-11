#!/bin/sh
PLUGINURL=$(gum input --placeholder "plugin url")

wget --no-parent -P %USERPROFILE%/.yet-another-cli/plugins -r $YACPATH/plugins  