# GitHub Actions Workflows

This directory contains GitHub Actions workflows for CI/CD automation.

## Workflows

### 1. CI/CD Pipeline (`ci-cd.yml`)

**Triggers:**

- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Jobs:**

- **Test**: Runs tests, linting, and security checks
- **Build**: Builds and pushes Docker images to AWS ECR (only on push)

**Note:** Deployment jobs have been moved to separate workflow files for better modularity and slash command support.

### 2. Slash Commands (`slash-commands.yml`)

**Triggers:**

- Comments on pull requests containing slash commands

**Available Commands:**

- `/deploy staging` - Deploy to staging environment
- `/deploy production` - Deploy to production environment
- `/rollback staging` - Rollback staging deployment
- `/rollback production` - Rollback production deployment

**Features:**

- **Authorization Check**: Only authorized users can trigger deployments
- **Automatic Workflow Trigger**: Triggers appropriate deployment workflow
- **PR Integration**: Works directly in pull request comments
- **Status Updates**: Updates comments with deployment status

**Usage Example:**

```bash
/deploy staging
```

This will trigger the staging deployment workflow and update the comment with status.

### 3. Deploy Staging (`deploy-staging.yml`)

**Triggers:**

- Manual workflow dispatch
- Called by slash commands workflow
- Can be triggered programmatically

**Features:**

- Deploys to staging environment
- Supports custom image tags
- Rollback functionality
- Health checks and verification

### 4. Deploy Production (`deploy-production.yml`)

**Triggers:**

- Manual workflow dispatch
- Called by slash commands workflow
- Can be triggered programmatically

**Features:**

- Deploys to production environment
- Supports custom image tags
- Rollback functionality
- Health checks, verification, and smoke tests

### 5. Pull Request Validation (`pr-validation.yml`)

**Triggers:**

- Pull requests to `main` or `develop` branches

**Jobs:**

- **Validate**: Runs tests, linting, security checks, and vulnerability scanning
- **Build Test**: Tests Docker builds without pushing

### 6. Manual Deploy (`manual-deploy.yml`)

**Triggers:**

- Manual workflow dispatch

**Features:**

- Deploy to staging or production
- Specify custom image tag
- Rollback functionality
- Health checks after deployment

## Required Secrets

Configure these secrets in your GitHub repository settings:

### AWS Configuration

- `AWS_ACCESS_KEY_ID`: AWS access key for ECR and EKS access
- `AWS_SECRET_ACCESS_KEY`: AWS secret key
- `AWS_REGION`: AWS region (e.g., `us-west-2`)
- `AWS_ACCOUNT_ID`: Your AWS account ID

### EKS Clusters

- `EKS_STAGING_CLUSTER`: Name of staging EKS cluster
- `EKS_PRODUCTION_CLUSTER`: Name of production EKS cluster

### Service URLs

- `STAGING_URL`: URL for staging environment health checks
- `PRODUCTION_URL`: URL for production environment health checks

## Environment Protection

The workflows use GitHub Environments for staging and production deployments:

1. **Staging Environment**: Automatically deploys from `develop` branch
2. **Production Environment**: Automatically deploys from `main` branch

Configure environment protection rules in GitHub:

- Required reviewers for production deployments
- Wait timer for production deployments
- Branch restrictions

## Usage

### Automatic Deployments

- Push to `develop` → Deploys to staging
- Push to `main` → Deploys to production

### Slash Command Deployments (Recommended)

**In Pull Request Comments:**

- `/deploy staging` - Deploy current PR to staging
- `/deploy production` - Deploy current PR to production
- `/rollback staging` - Rollback staging deployment
- `/rollback production` - Rollback production deployment

**Benefits:**

- No need to leave the PR
- Automatic authorization checks
- Immediate feedback and status updates
- Integration with PR workflow

### Manual Deployments

1. Go to Actions tab in GitHub
2. Select "Manual Deploy" workflow
3. Click "Run workflow"
4. Choose environment, image tag, and rollback options

### Rollback

Use the manual deploy workflow with the rollback option checked, or use slash commands like `/rollback staging` in PR comments.

## Docker Images

The workflow builds and pushes these images to AWS ECR:

- `ardanlabs/service-sales`
- `ardanlabs/service-auth`
- `ardanlabs/service-metrics`

## Kubernetes Deployment

The workflows use Kustomize to manage Kubernetes manifests:

- `zarf/k8s/staging/` - Staging environment manifests
- `zarf/k8s/production/` - Production environment manifests

Each environment has separate directories for:

- `database/` - PostgreSQL deployment
- `auth/` - Authentication service
- `sales/` - Sales service

## Security Features

- Vulnerability scanning with Trivy
- Security results uploaded to GitHub Security tab
- Go vulnerability checks with govulncheck
- Static analysis with staticcheck
- Docker image scanning

## Monitoring

The workflows include:

- Health checks after deployment
- Deployment status verification
- Pod and service status checks
- Automatic rollback on health check failures
