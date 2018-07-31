#!/usr/bin/env bash
REQUEST=$1
echo $1

set -e
set -o pipefail
MAPBOX_ACCESS_TOKEN="pk.eyJ1IjoicmNjbCIsImEiOiJjamRhcThyd3A3aWcwMzNvNXV1aTU4ZzBtIn0.1zaneG6GaAy44AvmfULNiw"

OFFLINE=./mbgl-offline

$OFFLINE --north 41.4664 --west 2.0407 --south 41.2724 --east 2.2680 --output barcelona.db --token $MAPBOX_ACCESS_TOKEN
# All
# $OFFLINE --north ${N-BOUND} --west ${W-BOUND} --south ${S-BOUND} --east ${E-BOUND} --output $OUTDPUT_DIR/$MAP_NAME --style $STYLE --token $MAPBOX_ACCESS_TOKEN
