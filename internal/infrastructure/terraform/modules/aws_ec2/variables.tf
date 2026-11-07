variable "project_name" {
  description = "The name of the project"
  type        = string
}

variable "region" {
  description = "The AWS region to deploy resources in"
  type        = string
  default     = "ap-southeast-1"
}

variable "instance_type" {
  description = "The EC2 instance type"
  type        = string
  default     = "t3.medium"
}

variable "disk_size" {
  description = "The size of the EBS root volume in GB"
  type        = number
  default     = 30
}
