#!/usr/bin/env bash
REQUEST=$1
echo $1

set -e
set -o pipefail
MAPBOX_ACCESS_TOKEN="pk.eyJ1IjoicmNjbCIsImEiOiJjams5MnRsa2YxcnV1M3BxbGw2ZTI1NW10In0.c56J36FCoHjnIM4UoIIWQQ"

OFFLINE=./mbgl-offline

$OFFLINE --north 41.4664 --west 2.0407 --south 41.2724 --east 2.2680 --output barcelona.db --token $MAPBOX_ACCESS_TOKEN
# All
# $OFFLINE --north ${N-BOUND} --west ${W-BOUND} --south ${S-BOUND} --east ${E-BOUND} --output $OUTDPUT_DIR/$MAP_NAME --style $STYLE --token $MAPBOX_ACCESS_TOKEN
