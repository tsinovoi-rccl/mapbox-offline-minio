FROM ubuntu:xenial
ENV MAPBOX_ACCESS_TOKEN "pk.eyJ1IjoicmNjbCIsImEiOiJjamRhcThyd3A3aWcwMzNvNXV1aTU4ZzBtIn0.1zaneG6GaAy44AvmfULNiw"
ENV APP_DIRECTORY /mapbox-gl-native
ENV APP_LIB_DIR /offline
LABEL NAME Mapbox-Offline
LABEL VERSION 1.0.0
RUN apt-get update \
    ## code dependencies
    && apt-get install -y --fix-missing git ca-certificates build-essential libxrandr-dev x11proto-randr-dev libgl1-mesa-dev gdb file ruby ccache cmake curl \
    && curl -sL https://deb.nodesource.com/setup_6.x | bash - \
    && apt-get install -y nodejs \
    && apt-get install -y --fix-missing zlib1g-dev golang-1.10-go libx11-dev libcurl4-openssl-dev xvfb libegl1-mesa-dev libgles2-mesa-dev libxcursor-dev libxinerama-dev libxi-dev dbus lcov \
    && apt-get clean \
    ## build mbgl-offline
    && git clone --progress https://github.com/mapbox/mapbox-gl-native.git \
    && sed -ie 's/handleError(curl::easy_setopt(handle, CURLOPT_CAINFO, "ca-bundle.crt"));/\/\/handleError(curl::easy_setopt(handle, CURLOPT_CAINFO, "ca-certificates.crt"));/g' /mapbox-gl-native/platform/default/http_file_source.cpp \
    && cd /mapbox-gl-native \
    && git submodule init \
    && git submodule update \
    && make offline \
    ## cleanup
    && mkdir /offline \
    && cp /mapbox-gl-native/build/linux-x86_64/Debug/mbgl-offline /offline/mbgl-offline \
    && rm -Rf /mapbox-gl-native/ \
    && apt-get remove -y build-essential cmake nodejs \
    zlib1g-dev libx11-dev xvfb \
    && apt autoremove -y
#Offline
COPY / /offline
WORKDIR /offline
#Add Hapttic
RUN export PATH="$PATH:/usr/lib/go-1.10/bin" \
export GOPATH="/offline" \
&& go get -u github.com/minio/minio-go \
&& go build -o hapttic .
EXPOSE 8080


ENTRYPOINT ["./hapttic","-file", "./mbgl-offline", "-logErrors=true", "-minioEndpoint", "127.0.0.1:9000", "-minioBucket", "maps", "-minioAccessSecret", "minio123", "-minioAccessID", "minio", "-minioSSL=false"]
CMD ["-help"]