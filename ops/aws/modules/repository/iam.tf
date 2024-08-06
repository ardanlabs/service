resource "aws_iam_policy" "ecr_access_policy" {
  name        = "${var.name}-policy"
  description = "Policy to allow pushing images to ECR"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "ecr:*"
        ],
        Resource = [
          for repo in aws_ecr_repository.repository : repo.arn
        ],
      },
      {
        Effect = "Allow",
        Action = [
          "ecr:GetAuthorizationToken"
        ],
        Resource = [
          "*"
        ]
      }
    ]
  })
}

resource "aws_iam_role" "github_action_role" {
  name = "${var.name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity",
        Effect = "Allow",
        Principal = {
          Federated = var.github_oidc_provider_arn
        },
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com",
            "token.actions.githubusercontent.com:sub" = [
              "repo:${var.github_owner}/${var.github_repo}:ref:refs/heads/main",
              "repo:${var.github_owner}/${var.github_repo}:pull_request"
            ]
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "github_action_policy_attachment" {
  role       = aws_iam_role.github_action_role.name
  policy_arn = aws_iam_policy.ecr_access_policy.arn
}
