# Ardan Labs Service - Tilt Configuration
# Local development environment with Kind, Kustomize, and Helm
#
# Infrastructure (Database, Observability) deployed via Kustomize
# Application services (Auth, Sales) deployed via Helm
#
# Usage:
#   tilt up              Start all services
#   tilt down            Stop all services
#   tilt logs <service>  View logs for a specific service
#
# Services are organized by labels:
#   - data: Database (Kustomize)
#   - observability: Grafana, Prometheus, Tempo, Loki, Promtail (Kustomize)
#   - build: Local binary compilation
#   - services: Auth, Sales (Helm with hot-reload via live_update)

# Load extensions
load('ext://restart_process', 'docker_build_with_restart')

# =============================================================================
# Configuration
# =============================================================================

# Kubernetes namespace
k8s_namespace = 'sales-system'

# Allow Kubernetes context (set to your cluster name)
allow_k8s_contexts('kind-ardan-starter-cluster')

# Tilt settings
update_settings()

# =============================================================================
# Infrastructure (using Kustomize)
# =============================================================================
# Observability and Database deployed via kustomize for dev
# Production uses managed services (RDS, Grafana Cloud, etc.)
# Note: Namespace is created by the kustomize manifests

k8s_yaml(kustomize('./zarf/k8s/dev/database'))
k8s_resource('database', labels=['data'])

k8s_yaml(kustomize('./zarf/k8s/dev/grafana'), allow_duplicates=True)
k8s_resource('grafana', labels=['observability'], port_forwards=['3100:3100'])

k8s_yaml(kustomize('./zarf/k8s/dev/prometheus'), allow_duplicates=True)
k8s_resource('prometheus-deployment', new_name='prometheus', labels=['observability'], port_forwards=['9090:9090'])

k8s_yaml(kustomize('./zarf/k8s/dev/tempo'), allow_duplicates=True)
k8s_resource('tempo', labels=['observability'])

k8s_yaml(kustomize('./zarf/k8s/dev/loki'), allow_duplicates=True)
k8s_resource('loki', labels=['observability'])

k8s_yaml(kustomize('./zarf/k8s/dev/promtail'), allow_duplicates=True)
k8s_resource('promtail', labels=['observability'])

# =============================================================================
# Auth Service
# =============================================================================

# Compile binary locally for fast iteration
local_resource(
  'auth-compile',
  'CGO_ENABLED=0 GOOS=linux go build -o ./zarf/build/auth ./api/services/auth',
  deps=['./api/services/auth', './app', './business', './foundation'],
  labels=['build'],
  ignore=['**/*_test.go'],
)

# Build Docker image with production dockerfile + hot-reload
docker_build_with_restart(
  'localhost:5001/ardanlabs/auth',
  '.',
  entrypoint=['/service/auth'],
  dockerfile='./zarf/docker/dockerfile.auth',
  build_args={'BUILD_TAG': 'develop'},
  live_update=[
    sync('./zarf/build/auth', '/service/auth'),
  ],
)

# Deploy auth service via Helm
# Using k8s_yaml + k8s_resource pattern to ensure proper image build dependency
k8s_yaml(helm(
  './zarf/helm/charts/auth',
  namespace=k8s_namespace,
  values=['./zarf/helm/charts/auth/values-dev.yaml'],
))

k8s_resource(
  'auth',
  labels=['services'],
  port_forwards=['6000:6000', '6010:6010'],  # HTTP, Debug
  resource_deps=['auth-compile', 'database'],
)

# =============================================================================
# Sales Service
# =============================================================================

# Compile binaries locally for fast iteration
local_resource(
  'sales-compile',
  'CGO_ENABLED=0 GOOS=linux go build -o ./zarf/build/sales ./api/services/sales && CGO_ENABLED=0 GOOS=linux go build -o ./zarf/build/admin ./api/tooling/admin',
  deps=['./api/services/sales', './api/tooling/admin', './app', './business', './foundation'],
  labels=['build'],
  ignore=['**/*_test.go'],
)

# Build Docker image with production dockerfile + hot-reload
docker_build_with_restart(
  'localhost:5001/ardanlabs/sales',
  '.',
  entrypoint=['/service/sales'],
  dockerfile='./zarf/docker/dockerfile.sales',
  build_args={'BUILD_TAG': 'develop'},
  live_update=[
    sync('./zarf/build/sales', '/service/sales'),
    sync('./zarf/build/admin', '/service/admin'),
  ],
)

# Deploy sales service via Helm
# Using k8s_yaml + k8s_resource pattern to ensure proper image build dependency
k8s_yaml(helm(
  './zarf/helm/charts/sales',
  namespace=k8s_namespace,
  values=['./zarf/helm/charts/sales/values-dev.yaml'],
))

k8s_resource(
  'sales',
  labels=['services'],
  port_forwards=['3000:3000', '3010:3010', '4020:4020'],  # Sales, Debug, Metrics
  resource_deps=['sales-compile', 'auth', 'database'],
)

# Run database migrations automatically after sales pod is ready
# This runs once on startup and when you manually trigger it in Tilt
local_resource(
  'sales-migrate',
  'kubectl wait --for=condition=ready pod -l app=sales -n sales-system --timeout=120s && kubectl exec -n sales-system deployment/sales -c sales -- ./admin migrate-seed',
  resource_deps=['sales'],
  labels=['services'],
)

# =============================================================================
# Metrics Service
# =============================================================================

# Compile binary locally for fast iteration
local_resource(
  'metrics-compile',
  'CGO_ENABLED=0 GOOS=linux go build -o ./zarf/build/metrics ./api/services/metrics',
  deps=['./api/services/metrics', './app', './business', './foundation'],
  labels=['build'],
  ignore=['**/*_test.go'],
)

# Build Docker image with production dockerfile + hot-reload
docker_build_with_restart(
  'localhost:5001/ardanlabs/metrics',
  '.',
  entrypoint=['/service/metrics'],
  dockerfile='./zarf/docker/dockerfile.metrics',
  build_args={'BUILD_TAG': 'develop'},
  live_update=[
    sync('./zarf/build/metrics', '/service/metrics'),
  ],
)


