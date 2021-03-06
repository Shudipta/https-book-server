#!/usr/bin/env bash
set -xe

GOPATH=$(go env GOPATH)
PACKAGE_NAME=https-book-server
REPO_ROOT="$GOPATH/src/github.com/shudipta/$PACKAGE_NAME"

pushd $REPO_ROOT
#go run cert-generator/certgen.go
if [ -x "$(command -v onessl)" ]; then
    export ONESSL=onessl
else
    curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-linux-amd64
    chmod +x onessl
    export ONESSL=./onessl
fi
#export certDir="/home/ac/go/src/github.com/soter/scanner/pki/clair"
export certDir="cert-generator"
export CA_CERT=$(cat $certDir/ca.crt | $ONESSL base64)
export SERVER_CERT=$(cat $certDir/server.crt | $ONESSL base64)
export SERVER_KEY=$(cat $certDir/server.key | $ONESSL base64)

cat ./hack/deploy/deployment.yaml | $ONESSL envsubst | kubectl apply -f -

pushd $REPO_ROOT/cert-generator
#rm -rf ca.crt ca.key srv.crt srv.key cl.crt cl.key
popd
popd