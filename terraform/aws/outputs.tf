output "bore_server_public_ip" {
  description = "Public IP address of the bore server"
  value       = aws_eip.bore_eip.public_ip
}

output "bore_server_instance_id" {
  description = "Instance ID of the bore server"
  value       = aws_instance.bore_server.id
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.bore_vpc.id
}

output "security_group_id" {
  description = "Security group ID for bore server"
  value       = aws_security_group.bore_server_sg.id
}