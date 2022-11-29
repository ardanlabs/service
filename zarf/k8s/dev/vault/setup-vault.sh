#!/bin/sh
kubectl -n vault exec vault-0 -- vault operator init -key-shares=1 -key-threshold=1 -format=json > keys.json

VAULT_UNSEAL_KEY=$(cat keys.json | jq -r ".unseal_keys_b64[]")
echo Vault Unseal Key: $VAULT_UNSEAL_KEY

VAULT_ROOT_KEY=$(cat keys.json | jq -r ".root_token")
echo Vault Root Key: $VAULT_ROOT_KEY

kubectl -n vault exec vault-0 -- vault operator unseal $VAULT_UNSEAL_KEY

kubectl -n vault exec vault-0 -- vault login $VAULT_ROOT_KEY

kubectl -n vault exec vault-0 -- vault secrets enable -version=2 -path="demo-app" kv

kubectl -n vault exec vault-0 -- vault kv put demo-app/user01 name=devopscube

kubectl -n vault exec vault-0 -- vault kv get demo-app/user01

kubectl -n vault exec vault-0 -- vault policy write demo-policy - <<EOH
path "demo-app/*" {
  capabilities = ["read"]
}
EOH

kubectl -n vault exec vault-0 -- vault policy list

kubectl -n vault exec vault-0 -- vault auth enable kubernetes

kubectl -n vault exec vault-0 -- vault write auth/kubernetes/config \
        token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
        kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443" \
        kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

kubectl -n vault exec vault-0 -- kubectl create serviceaccount vault-auth

kubectl -n vault exec vault-0 -- vault write auth/kubernetes/role/webapp \
        bound_service_account_names=vault-auth \
        bound_service_account_namespaces=default \
        policies=demo-policy \
        ttl=72h
