#!/usr/bin/env bash

export MINIO_ACCESS_KEY="minio"
export MINIO_SECRET_KEY="minio123"



minio server --address 127.0.0.1:9000 /Users/tsinoel/mapbox-gl-native/minio

