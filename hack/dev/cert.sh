#!/usr/bin/env bash

# https://stackoverflow.com/a/677212/244009
if [ -x "$(command -v onessl)" ]; then
    export ONESSL=onessl
else
    # ref: https://stackoverflow.com/a/27776822/244009
    case "$(uname -s)" in
        Darwin)
            curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-darwin-amd64
            chmod +x onessl
            export ONESSL=./onessl
            ;;

        Linux)
            curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-linux-amd64
            chmod +x onessl
            export ONESSL=./onessl
            ;;

        CYGWIN*|MINGW32*|MSYS*)
            curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-windows-amd64.exe
            chmod +x onessl.exe
            export ONESSL=./onessl.exe
            ;;
        *)
            echo 'other OS'
            ;;
    esac
fi

echo "creating necessary certificate-key pairs"

# create necessary TLS certificates:
# - a local CA key and cert
# - a webhook server key and cert signed by the local CA
$ONESSL create ca-cert --cert-dir=pki/scanner
$ONESSL create server-cert server --cert-dir=pki/scanner --domains=scanner.kube-system.svc --ips="192.168.99.100,127.0.0.1"

# In the clair notifier part, server=scanner-server, client=clair
# create necessary TLS certificates:
# - a client key and cert signed by the above local CA for clair notifier
$ONESSL create client-cert client --cert-dir=pki/scanner

# In the clair api part: server=clair, client=scanner-server
# create necessary TLS certificates:
# - a CA key and cert for clair api
# - a server key and cert signed by this CA for clair api
# - a client key and cert signed by this CA for clair api
$ONESSL create ca-cert --cert-dir=pki/clair
$ONESSL create server-cert server --cert-dir=pki/clair --domains="clairsvc.kube-system.svc,localhost" --ips="192.168.99.100,127.0.0.1"
$ONESSL create client-cert client --cert-dir=pki/clair