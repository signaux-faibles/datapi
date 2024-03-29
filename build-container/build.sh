#!/bin/sh

if [ "$#" -ne 1 ]; then
    echo "build.sh: construit l'application datapi dans une image"
    echo "usage: build.sh branch"
    echo "exemple: ./build.sh master"
    exit 255
fi

if [ -d workspace ]; then
    echo "supprimer le répertoire workspace avant de commencer"
    exit 1
fi
 
# Checkout git
mkdir workspace
cd workspace
curl -LOs "https://github.com/signaux-faibles/datapi/archive/$1.zip"

if [ $(openssl dgst -md5 "$1.zip" |awk '{print $2}') = '3be7b8b182ccd96e48989b4e57311193' ]; then
   echo "sources manquantes, branche probablement inexistante"
   exit
fi

# Unzip des sources et build
unzip "$1.zip"
cd "datapi-$1"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# Build docker
cd ../..
docker build --tag datapi --build-arg datapiDir="./workspace/datapi-$1/" .
docker save datapi | gzip > datapi.tar.gz

# Cleanup
rm -rf workspace
