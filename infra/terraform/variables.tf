variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Deployment environment (prod, staging, dev)"
  type        = string
  default     = "prod"
}

variable "eks_node_min_size" {
  description = "Minimum number of nodes in EKS cluster"
  type        = number
  default     = 3
}

variable "eks_node_max_size" {
  description = "Maximum number of nodes in EKS cluster"
  type        = number
  default     = 10
}

variable "eks_instance_types" {
  description = "Instance types for EKS node group"
  type        = list(string)
  default     = ["t3.large", "t3.xlarge"]
}
