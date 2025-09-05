terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

# VPC Configuration
resource "aws_vpc" "bore_vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "bore-vpc"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "bore_igw" {
  vpc_id = aws_vpc.bore_vpc.id

  tags = {
    Name = "bore-igw"
  }
}

# Public Subnets
resource "aws_subnet" "bore_public_subnet" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.bore_vpc.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name = "bore-public-subnet-${count.index + 1}"
  }
}

# Route Table
resource "aws_route_table" "bore_public_rt" {
  vpc_id = aws_vpc.bore_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.bore_igw.id
  }

  tags = {
    Name = "bore-public-rt"
  }
}

resource "aws_route_table_association" "bore_public_rta" {
  count          = length(aws_subnet.bore_public_subnet)
  subnet_id      = aws_subnet.bore_public_subnet[count.index].id
  route_table_id = aws_route_table.bore_public_rt.id
}

# Security Groups
resource "aws_security_group" "bore_server_sg" {
  name_prefix = "bore-server-"
  vpc_id      = aws_vpc.bore_vpc.id

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP port for bore server"
  }

  ingress {
    from_port   = 8443
    to_port     = 8443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS port for bore server"
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH access"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "bore-server-sg"
  }
}

# EC2 Instance for Bore Server
resource "aws_instance" "bore_server" {
  ami           = var.ami_id
  instance_type = var.instance_type
  key_name      = var.key_name

  vpc_security_group_ids = [aws_security_group.bore_server_sg.id]
  subnet_id              = aws_subnet.bore_public_subnet[0].id

  user_data = templatefile("${path.module}/user_data.sh", {
    api_key = var.api_key
  })

  tags = {
    Name = "bore-server"
  }
}

# Elastic IP
resource "aws_eip" "bore_eip" {
  instance = aws_instance.bore_server.id
  vpc      = true

  tags = {
    Name = "bore-server-eip"
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "bore_logs" {
  name              = "/bore/server"
  retention_in_days = 30

  tags = {
    Name = "bore-logs"
  }
}