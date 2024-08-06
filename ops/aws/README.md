# Terraform stack

The subdirectory contains terraform sources to build the following infrastructure:
* EKS cluster
* ECR repository

## Terraform state (optional)

To setup the infrastructure, it is better to use an S3 bucket
to store the terraform state. The bucket name should be inserted 
in the `providers.tf` file, under `terraform.backend.s3.bucket` key.

## Usage
To build the infrastructure, run the following commands:
```bash
terraform init
```

```bash
terraform apply
```

After the infrastructure is built, you can use the following command to get the kubeconfig file:
```bash
aws eks --region $(terraform output -raw region) update-kubeconfig \
    --name eks-cluster
```
