output "public_ip" {
  description = "The public IP of the EC2 instance"
  value       = aws_eip.k3s.public_ip
}

output "private_key_pem" {
  description = "The private key PEM of the SSH key pair"
  value       = tls_private_key.pk.private_key_pem
  sensitive   = true
}
