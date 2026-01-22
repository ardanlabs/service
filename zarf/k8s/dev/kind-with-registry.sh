#!/bin/sh
set -o errexit

# Desired cluster name; default is "kind"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind-ardan-starter-cluster}"

reg_name='kind-registry'
reg_port='5001'

# 1. Create registry container unless it already exists
if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --network bridge --name "${reg_name}" \
    registry:2
fi

# 2. Create kind cluster with containerd registry config dir enabled
# TODO: kind will eventually enable this by default and this patch will
# be unnecessary.
#
# See:
# https://github.com/kubernetes-sigs/kind/issues/2875
# https://github.com/containerd/containerd/blob/main/docs/cri/config.md#registry-configuration
cat <<EOF | kind create cluster --name="${KIND_CLUSTER_NAME}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry]
    config_path = "/etc/containerd/certs.d"
nodes:
  - role: control-plane
    extraPortMappings:
      # Sales-Api
      - containerPort: 3000
        hostPort: 3000
      # Sales-Api debug
      - containerPort: 3010
        hostPort: 3010
      # Metrics
      - containerPort: 4000
        hostPort: 4000
      # Metrics debug
      - containerPort: 4010
        hostPort: 4010
      # Metrics Prometheus
      - containerPort: 4020
        hostPort: 4020
      # Auth
      - containerPort: 6000
        hostPort: 6000
      # Auth debug
      - containerPort: 6010
        hostPort: 6010
      # Grafana
      - containerPort: 3100
        hostPort: 3100
      # Postgres
      - containerPort: 5432
        hostPort: 5432
      # Prometheus
      - containerPort: 9090
        hostPort: 9090
      # Tempo (tracing)
      - containerPort: 9411
        hostPort: 9411
EOF

# 3. Add the registry config to the nodes
#
# This is necessary because localhost resolves to loopback addresses that are
# network-namespace local.
# In other words: localhost in the container is not localhost on the host.
#
# We want a consistent name that works from both ends, so we tell containerd to
# alias localhost:${reg_port} to the registry container when pulling images
REGISTRY_DIR="/etc/containerd/certs.d/localhost:${reg_port}"
for node in $(kind get nodes --name "${KIND_CLUSTER_NAME}"); do
  docker exec "${node}" mkdir -p "${REGISTRY_DIR}"
  cat <<EOF | docker exec -i "${node}" cp /dev/stdin "${REGISTRY_DIR}/hosts.toml"
[host."http://${reg_name}:5000"]
EOF
done

# 4. Connect the registry to the cluster network if not already connected
# This allows kind to bootstrap the network but ensures they're on the same network
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${reg_name}")" = 'null' ]; then
  docker network connect "kind" "${reg_name}"
fi

# 5. Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

echo "Registry is running at localhost:${reg_port}"
echo "Cluster ${KIND_CLUSTER_NAME} is ready with local registry!"
