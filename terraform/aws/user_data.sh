#!/bin/bash
set -e

# Update system
yum update -y

# Install Docker
amazon-linux-extras install docker -y
systemctl start docker
systemctl enable docker
usermod -a -G docker ec2-user

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create bore directory
mkdir -p /opt/bore
cd /opt/bore

# Download latest bore release (replace with actual release URL)
# wget https://github.com/4cecoder/bore/releases/latest/download/bore-server-linux-amd64.tar.gz
# tar -xzf bore-server-linux-amd64.tar.gz

# For now, create a placeholder
echo "Bore server would be downloaded and installed here" > README.txt

# Create systemd service
cat > /etc/systemd/system/bore.service << EOF
[Unit]
Description=Bore Tunneling Server
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/bore
ExecStart=/opt/bore/server
Restart=always
RestartSec=5
Environment=API_KEY=${api_key}

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable bore
# systemctl start bore  # Uncomment when binary is available

# Install CloudWatch agent for logging
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U amazon-cloudwatch-agent.rpm

# Configure CloudWatch agent
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json << EOF
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/bore/server.log",
            "log_group_name": "/bore/server",
            "log_stream_name": "{instance_id}"
          }
        ]
      }
    }
  }
}
EOF

systemctl enable amazon-cloudwatch-agent
systemctl start amazon-cloudwatch-agent