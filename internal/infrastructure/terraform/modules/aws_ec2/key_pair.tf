resource "tls_private_key" "pk" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "kp" {
  key_name   = "${var.project_name}-key"
  public_key = tls_private_key.pk.public_key_openssh

  tags = {
    Name           = "${var.project_name}-key"
    "cakd:managed" = "true"
    "cakd:project" = var.project_name
  }
}
