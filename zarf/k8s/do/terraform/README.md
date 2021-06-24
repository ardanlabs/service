# Terraform DigitalOcean Kubernetes

This project contains Terraform configuration that can be used to create a Kubernetes cluster on DigitalOcean.

## Layout

`main.tf`

This is the root Terraform module that is responsible for calling any child modules such as `modules/k8s-digitalocean`.  It is similar to a `main()` function in a Go program.

`modules/k8s-digitalocean`

This is the Terraform module that can be called to create the Kubernetes clusters on DigitalOcean. A Terraform module is similar to a library function that can be called to do the things that you want to do.

`modules/k8s-digitalocean/main.tf`

This is the main Terraform configuration for the `modules/k8s-digitalocean` module that creates the VPC and the Kubernetes cluster. It also defines an output for the raw `kubectl` configuration that can be read by the caller.

`modules/k8s-digitalocean/variables.tf`

These are the variables that the `modules/k8s-digitalocean` module can accept as inputs. Most of them are optional. Each variable has its own description as to what it is.

## Usage

First, [download Terraform](https://www.terraform.io/downloads.html).

Then export your DigitalOcean API token.

```
export DIGITALOCEAN_ACCESS_TOKEN='CHANGEME'
```

Initialize the Terraform configuration.

```
terraform init
```

Execute a Terraform plan to see what operations Terraform wants to perform.

```
terraform plan
```

Execute a Terraform apply to perform the desired operations and create the resources.

```
terraform apply
```

## Terraform State

Terraform tracks the resources it creates in a state file named `terraform.tfstate`. This state file may contain sensitive information and must not be checked into source control. Use the following `.gitignore` configuration to ignore Terraform files from being checked into source control.

```
**/*.tfstate
**/*.tfstate.*
**/.terraform/*
**/.terraformrc
**/terraform.rc
```

If you don't want to store the state locally, consider using an [alternative backend](https://www.terraform.io/docs/language/settings/backends/index.html).
