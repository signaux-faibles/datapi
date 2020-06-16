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

echo 3be7b8b182ccd96e48989b4e57311193 $1.zip > "$1.zip.md5"
if md5sum -c $1.zip.md5 --quiet; then
   echo "sources manquantes, branche probablement inexistante"
   exit
fi

# Unzip des sources et build
unzip "$1.zip"
cd "datapi-$1"

CGO_ENABLED=0 go build

# Build docker
cd ../..
docker build -t datapi --build-arg binary="./workspace/datapi-$1/datapi" . 
docker save datapi | gzip > datapi.tar.gz

# Cleanup
rm workspace -rf