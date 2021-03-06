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

export CA_CERT=$(cat cert-generator/ca.crt | $ONESSL base64)
export CLIENT_CERT=$(cat cert-generator/client.crt | $ONESSL base64)
export CLIENT_KEY=$(cat cert-generator/client.key | $ONESSL base64)

cat ./hack/deploy/deployment-client.yaml | $ONESSL envsubst | kubectl apply -f -

pushd $REPO_ROOT/cert-generator
# rm -rf ca.crt ca.key srv.crt srv.key cl.crt cl.key
popd
rm -rf hack/docker-client/https-client
popd