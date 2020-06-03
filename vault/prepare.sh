#!/usr/bin/env bash

set -e

vault_installation=${BASH_SOURCE%/*}/volumes

if [ ! -z "$1" ] && [ "$1" == "dev" ] ; then
  compose_file=${BASH_SOURCE%/*}/docker-compose.$1.yml
else
  compose_file=${BASH_SOURCE%/*}/docker-compose.yml
fi
function enable_kv {
  if [ -z "$(vault secrets list | grep $1/)" ] ; then
    vault secrets enable -max-lease-ttl=$2 -path=$1 -description="$3" kv-v2
  fi
}

function enable_userpass_auth {
  if [ -z "$(vault auth list | grep userpass/)" ] ; then
    vault auth enable userpass
  fi
}

function add_vault_user {
  vault write auth/userpass/users/$1 password=$2 token_policies=default,$3
}

function download_plugin_secrets_gen {
  mkdir -p $vault_installation/plugins

  plugin_version=0.0.6
  plugin=secrets-gen

  if [ ! -f $vault_installation/plugins/vault-$plugin ] ; then
    plugin_tar=vault-${plugin}_${plugin_version}_linux_amd64.tgz
    curl -LsSo ${plugin_tar} https://github.com/sethvargo/vault-secrets-gen/releases/download/v${plugin_version}/${plugin_tar}
    tar xzf ${plugin_tar} && rm ${plugin_tar}
    mv vault-${plugin} ${vault_installation}/plugins
  fi
}

function install_plugin_secrets_gen {
  sha256=$(shasum -a 256 "$vault_installation/plugins/vault-secrets-gen" | cut -d' ' -f1)

  vault plugin register -sha256="${sha256}" -command="vault-secrets-gen" secret vault-secrets-gen
  if [ -z "$(vault secrets list | grep vault-secrets-gen)" ] ; then
    vault secrets enable -path="gen" -plugin-name="vault-secrets-gen" plugin
  fi
}

download_plugin_secrets_gen
docker-compose -f $compose_file up -d
sleep 4
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_UNSEAL_KEY="$(docker-compose -f $compose_file logs | grep "Unseal Key" | cut -d ':' -f2 | xargs)"
export VAULT_TOKEN="$(docker-compose -f $compose_file logs | grep "Root Token" | cut -d ':' -f2 | xargs)"
env | grep VAULT
vault status

vault policy write dctna ${BASH_SOURCE%/*}/policies/dctna-policy.hcl
enable_kv dctna 720h "Docker Content Trust Notary Admin"
enable_userpass_auth
add_vault_user dctna topsecret dctna
install_plugin_secrets_gen

echo Add root token credential
vault kv put dctna/dev/760e57b96f72ed27e523633d2ffafe45ae0ff804e78dfc014a50f01f823d161d password=test1234 alias=root
vault kv put dctna/dev/4ea1fec36392486d4bd99795ffc70f3ffa4a76185b39c8c2ab1d9cf5054dbbc9 password=test1234 alias=localhost:5000/dct-notary-admin
