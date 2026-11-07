data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_iam_role" "ec2_role" {
  name = "${var.project_name}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name           = "${var.project_name}-ec2-role"
    "cakd:managed" = "true"
    "cakd:project" = var.project_name
  }
}

resource "aws_iam_role_policy_attachment" "ssm" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "ec2_profile" {
  name = "${var.project_name}-ec2-profile"
  role = aws_iam_role.ec2_role.name

  tags = {
    Name           = "${var.project_name}-ec2-profile"
    "cakd:managed" = "true"
    "cakd:project" = var.project_name
  }
}

resource "aws_instance" "k8s_node" {
  ami                  = data.aws_ami.ubuntu.id
  instance_type        = var.instance_type
  subnet_id            = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.ec2.id]
  key_name             = aws_key_pair.kp.key_name
  iam_instance_profile = aws_iam_instance_profile.ec2_profile.name

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }

  root_block_device {
    volume_size           = var.disk_size
    volume_type           = "gp3"
    encrypted             = true
    delete_on_termination = true
  }

  tags = {
    Name           = "${var.project_name}-k8s-node"
    "cakd:managed" = "true"
    "cakd:project" = var.project_name
  }
}

resource "aws_eip" "k3s" {
  instance = aws_instance.k8s_node.id

  tags = {
    Name           = "${var.project_name}-eip"
    "cakd:managed" = "true"
    "cakd:project" = var.project_name
  }
}
