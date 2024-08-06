resource "aws_ecr_repository" "repository" {
  for_each =  toset([ "auth", "metrics", "sales" ])
  name                 = format("%s-%s", var.name, each.value)
  image_tag_mutability = "MUTABLE"
  image_scanning_configuration {
    scan_on_push = true
  }
}
