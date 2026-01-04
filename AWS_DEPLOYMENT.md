# AWS Deployment Guide

This guide covers deploying the Coach Calendar application on AWS with SES email integration.

## Prerequisites

1. AWS Account
2. AWS CLI configured
3. Domain or verified email for SES

## AWS SES Setup

### 1. Verify Your Email/Domain

**Option A: Verify Single Email (Quick Start)**
```bash
aws ses verify-email-identity --email-address your-email@yourdomain.com --region us-east-1
```
Then check your email and click the verification link.

**Option B: Verify Domain (Production)**
1. Go to AWS SES Console → Verified identities
2. Click "Create identity"
3. Select "Domain"
4. Enter your domain
5. Add the provided DNS records to your domain

### 2. Move Out of Sandbox Mode (For Production)

By default, SES is in sandbox mode and can only send to verified emails.

To send to any email address:
1. Go to SES Console → Account dashboard
2. Click "Request production access"
3. Fill out the form explaining your use case
4. Wait for AWS approval (usually 24 hours)

### 3. Create IAM Policy for SES

Create a file `ses-policy.json`:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ses:SendEmail",
        "ses:SendRawEmail"
      ],
      "Resource": "*"
    }
  ]
}
```

Create the policy:
```bash
aws iam create-policy \
  --policy-name CoachCalendarSESPolicy \
  --policy-document file://ses-policy.json
```

## Deployment Options

### Option 1: EC2 Instance

#### Step 1: Create IAM Role
```bash
# Create trust policy
cat > trust-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

# Create role
aws iam create-role \
  --role-name CoachCalendarEC2Role \
  --assume-role-policy-document file://trust-policy.json

# Attach SES policy
aws iam attach-role-policy \
  --role-name CoachCalendarEC2Role \
  --policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/CoachCalendarSESPolicy

# Create instance profile
aws iam create-instance-profile \
  --instance-profile-name CoachCalendarEC2Role

# Add role to instance profile
aws iam add-role-to-instance-profile \
  --instance-profile-name CoachCalendarEC2Role \
  --role-name CoachCalendarEC2Role
```

#### Step 2: Launch EC2 Instance

```bash
# Create user data script
cat > user-data.sh << 'EOF'
#!/bin/bash
# Update system
yum update -y

# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile

# Clone and build application
cd /home/ec2-user
git clone YOUR_REPO_URL coach-calendar
cd coach-calendar
/usr/local/go/bin/go build -o coach-calendar

# Set environment variables
cat > /etc/systemd/system/coach-calendar.env << ENVEOF
USE_AWS_SES=true
SMTP_FROM=verified@yourdomain.com
AWS_REGION=us-east-1
PORT=8080
ENVEOF

# Create systemd service
cat > /etc/systemd/system/coach-calendar.service << SERVICEEOF
[Unit]
Description=Coach Calendar Application
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/home/ec2-user/coach-calendar
EnvironmentFile=/etc/systemd/system/coach-calendar.env
ExecStart=/home/ec2-user/coach-calendar/coach-calendar
Restart=always

[Install]
WantedBy=multi-user.target
SERVICEEOF

# Start service
systemctl daemon-reload
systemctl enable coach-calendar
systemctl start coach-calendar
EOF

# Launch instance
aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.micro \
  --iam-instance-profile Name=CoachCalendarEC2Role \
  --user-data file://user-data.sh \
  --security-group-ids sg-YOUR_SECURITY_GROUP \
  --subnet-id subnet-YOUR_SUBNET \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=CoachCalendar}]'
```

### Option 2: ECS (Elastic Container Service)

#### Step 1: Create Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o coach-calendar

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/coach-calendar .
EXPOSE 8080
CMD ["./coach-calendar"]
```

#### Step 2: Build and Push to ECR
```bash
# Create ECR repository
aws ecr create-repository --repository-name coach-calendar

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# Build and push
docker build -t coach-calendar .
docker tag coach-calendar:latest YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/coach-calendar:latest
docker push YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/coach-calendar:latest
```

#### Step 3: Create ECS Task Definition
```json
{
  "family": "coach-calendar",
  "taskRoleArn": "arn:aws:iam::YOUR_ACCOUNT_ID:role/CoachCalendarTaskRole",
  "executionRoleArn": "arn:aws:iam::YOUR_ACCOUNT_ID:role/ecsTaskExecutionRole",
  "networkMode": "awsvpc",
  "containerDefinitions": [
    {
      "name": "coach-calendar",
      "image": "YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/coach-calendar:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "USE_AWS_SES",
          "value": "true"
        },
        {
          "name": "SMTP_FROM",
          "value": "verified@yourdomain.com"
        },
        {
          "name": "AWS_REGION",
          "value": "us-east-1"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/coach-calendar",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ],
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512"
}
```

### Option 3: Elastic Beanstalk

```bash
# Install EB CLI
pip install awsebcli

# Initialize EB application
eb init -p go coach-calendar --region us-east-1

# Create environment
eb create coach-calendar-env \
  --instance-profile CoachCalendarEC2Role \
  --envvars USE_AWS_SES=true,SMTP_FROM=verified@yourdomain.com,AWS_REGION=us-east-1

# Deploy
eb deploy
```

## Testing Email Functionality

After deployment, test the email functionality:

```bash
# Check application logs
# EC2:
sudo journalctl -u coach-calendar -f

# ECS:
aws logs tail /ecs/coach-calendar --follow

# Look for:
# "Email service enabled using AWS SES in region: us-east-1"
# "Confirmation email sent successfully via AWS SES to..."
```

## Troubleshooting

### Email not sending

1. **Check SES verified identities**:
   ```bash
   aws ses list-verified-email-addresses --region us-east-1
   ```

2. **Check IAM role permissions**:
   ```bash
   aws iam list-attached-role-policies --role-name CoachCalendarEC2Role
   ```

3. **Check SES sending limits**:
   ```bash
   aws ses get-send-quota --region us-east-1
   ```

4. **Check application logs** for error messages

### Common Issues

- **"MessageRejected: Email address is not verified"**: Verify the sender email in SES
- **"AccessDenied"**: Check IAM role has SES permissions
- **"Daily sending quota exceeded"**: Request limit increase in SES console
- **"Account is in sandbox mode"**: Request production access

## Cost Estimation

- **SES**: $0.10 per 1,000 emails (first 62,000 emails/month free with EC2)
- **EC2 t3.micro**: ~$7.50/month (1-year reserved: ~$4/month)
- **ECS Fargate**: ~$15/month for 0.25 vCPU, 0.5 GB
- **Data transfer**: Usually negligible for small apps

## Security Best Practices

1. **Use IAM roles** instead of access keys
2. **Enable CloudWatch logs** for monitoring
3. **Set up billing alerts**
4. **Use VPC** for network isolation
5. **Enable AWS WAF** if exposing publicly
6. **Rotate credentials regularly**
7. **Monitor SES reputation dashboard**

## Next Steps

1. Set up domain and SSL certificate (ACM)
2. Configure Application Load Balancer
3. Set up CloudWatch alarms
4. Configure auto-scaling
5. Set up backup strategy for SQLite database
