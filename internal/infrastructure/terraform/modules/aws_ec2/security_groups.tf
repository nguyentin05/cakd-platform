resource "aws_security_group" "ec2" {
  name        = "${var.project_name}-ec2-sg"
  description = "Security Group for CAKD single-node Kubernetes EC2 instance"
  vpc_id      = aws_vpc.main.id

  # SSH Access
  # trivy:ignore:AVD-AWS-0107
  ingress {
    description = "Allow SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    # trivy:ignore:AVD-AWS-0107
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Kubernetes API
  ingress {
    description = "Allow Kubernetes API"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTP
  ingress {
    description = "Allow HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS
  ingress {
    description = "Allow HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Outbound rules
  # trivy:ignore:AVD-AWS-0104
  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    # trivy:ignore:AVD-AWS-0104
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name           = "${var.project_name}-ec2-sg"
    "cakd:managed" = "true"
    "cakd:project" = var.project_name
  }
}
