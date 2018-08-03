# MapBox Offline Map Cache Image

## Running on Local without Server - Script Only

`./mbgl-offline --north 41.4664 --west 2.0407 --south 41.2724 --east2.2680 --output barcelona.db --style mapbox://styles/rccl/cjgs6mu3w000b2rptxpikfwuv --token mapboxToken`

## Running with Go Server Listening at "/"

To run server:
`./hapttic -file ./mbgl-offline -logErrors=true -minioEndpoint 127.0.0.1:9000 -minioAccessID minio -minioAccessSecret minio123 -minioBucket maps -minioLocation ship -minioSSL=false`

This is using the hapttic library to run shells triggered by API endpoint at PORT 8080
It takes in a *POST* request with a body like:
 ` {
        "n-bound": "41.4664",
        "s-bound": "41.2724",
        "w-bound": "2.0407",
        "e-bound": "2.2680",
        "mapStyle":"mapbox://styles/rccl/cjgs6mu3w000b2rptxpikfwuv",
        "minZoom":"4.0",
        "maxZoom":"10.0",
        "pixelRatio":"1.0",
        "mapName":"barcelona.db",
        "outputDir":"./barcelona.db",
        "mapboxAccessToken":"mapboxToken"
    }
`

## Docker Image

To build:
`docker build -t mapbox-offline:1.0 .`

To run:
`docker run -it -d  -p 8080:8080 mapbox-offline:1.0`
