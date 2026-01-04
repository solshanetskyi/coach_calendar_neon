# AWS App Runner Deployment Guide

This guide covers deploying the Coach Calendar application using AWS App Runner with automatic GitHub integration.

## Prerequisites

1. AWS Account with appropriate permissions
2. GitHub repository with the Coach Calendar code
3. (Optional) Verified email in AWS SES for email confirmations

## Deployment Options

AWS App Runner supports two deployment methods:
1. **Source code from GitHub** (Recommended - easiest)
2. **Container image from ECR**

## Option 1: Deploy from GitHub (Recommended)

### Step 1: Prepare Your Repository

The repository already includes:
- `apprunner.yaml` - App Runner configuration file
- `Dockerfile` - Alternative containerized deployment
- All necessary source code

### Step 2: Create App Runner Service

#### Via AWS Console:

1. **Navigate to AWS App Runner Console**
   - Go to https://console.aws.amazon.com/apprunner

2. **Create Service**
   - Click "Create service"

3. **Source Configuration**
   - Repository type: **Source code repository**
   - Connect to GitHub:
     - Click "Add new"
     - Authorize AWS Connector for GitHub
     - Install in your GitHub account
   - Repository: Select `coach_calendar`
   - Branch: `first_usable_version`
   - Deployment trigger: **Automatic** (deploys on every push)

4. **Build Configuration**
   - Configuration source: **Use a configuration file**
   - Configuration file: `apprunner.yaml`
   - Runtime: Go 1
   - Build command: Will use apprunner.yaml settings
   - Start command: Will use apprunner.yaml settings

5. **Service Configuration**
   - Service name: `coach-calendar`
   - Port: `8080` (already configured in apprunner.yaml)
   - CPU: `1 vCPU`
   - Memory: `2 GB`
   - Environment variables (optional for email):
     ```
     USE_AWS_SES=true
     SMTP_FROM=verified@yourdomain.com
     AWS_REGION=us-east-1
     ```

6. **Auto Scaling** (optional)
   - Min instances: `1`
   - Max instances: `25`
   - Max concurrency: `100`

7. **Health Check**
   - Path: `/health`
   - Interval: `10 seconds`
   - Timeout: `5 seconds`
   - Unhealthy threshold: `3`
   - Healthy threshold: `3`

8. **Security - IAM Role** (for SES)
   - Create new service role or use existing
   - Add policy for SES (see below)

9. **Review and Create**
   - Review all settings
   - Click "Create & deploy"

#### Via AWS CLI:

```bash
# Create apprunner.json configuration file
cat > apprunner.json << 'EOF'
{
  "ServiceName": "coach-calendar",
  "SourceConfiguration": {
    "AuthenticationConfiguration": {
      "ConnectionArn": "YOUR_GITHUB_CONNECTION_ARN"
    },
    "AutoDeploymentsEnabled": true,
    "CodeRepository": {
      "RepositoryUrl": "https://github.com/solshanetskyi/coach_calendar",
      "SourceCodeVersion": {
        "Type": "BRANCH",
        "Value": "first_usable_version"
      },
      "CodeConfiguration": {
        "ConfigurationSource": "API",
        "CodeConfigurationValues": {
          "Runtime": "GO_1",
          "BuildCommand": "CGO_ENABLED=1 go build -o coach-calendar .",
          "StartCommand": "./coach-calendar",
          "Port": "8080",
          "RuntimeEnvironmentVariables": {
            "PORT": "8080"
          }
        }
      }
    }
  },
  "InstanceConfiguration": {
    "Cpu": "1 vCPU",
    "Memory": "2 GB",
    "InstanceRoleArn": "YOUR_INSTANCE_ROLE_ARN"
  },
  "HealthCheckConfiguration": {
    "Protocol": "HTTP",
    "Path": "/health",
    "Interval": 10,
    "Timeout": 5,
    "HealthyThreshold": 3,
    "UnhealthyThreshold": 3
  }
}
EOF

# Create the service
aws apprunner create-service --cli-input-json file://apprunner.json --region us-east-1
```

### Step 3: Configure IAM Role for SES (Optional)

If you want email confirmations:

1. **Create IAM Policy** for SES:
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

2. **Attach to App Runner Instance Role**:
```bash
# Create policy
aws iam create-policy \
  --policy-name CoachCalendarSESPolicy \
  --policy-document file://ses-policy.json

# Attach to App Runner instance role
aws iam attach-role-policy \
  --role-name AppRunnerInstanceRoleForCoachCalendar \
  --policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/CoachCalendarSESPolicy
```

3. **Update App Runner Service** with environment variables:
```bash
aws apprunner update-service \
  --service-arn YOUR_SERVICE_ARN \
  --source-configuration "CodeRepository={CodeConfiguration={CodeConfigurationValues={RuntimeEnvironmentVariables={USE_AWS_SES=true,SMTP_FROM=verified@yourdomain.com,AWS_REGION=us-east-1}}}}"
```

### Step 4: Verify Deployment

```bash
# Get service status
aws apprunner describe-service --service-arn YOUR_SERVICE_ARN

# Check service URL
# The URL will be something like: https://xxxxx.us-east-1.awsapprunner.com
```

## Option 2: Deploy from Container (ECR)

If you prefer to use Docker:

### Step 1: Build and Push to ECR

```bash
# Create ECR repository
aws ecr create-repository --repository-name coach-calendar --region us-east-1

# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# Build image
docker build -t coach-calendar .

# Tag image
docker tag coach-calendar:latest YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/coach-calendar:latest

# Push to ECR
docker push YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/coach-calendar:latest
```

### Step 2: Create App Runner Service from ECR

```bash
aws apprunner create-service \
  --service-name coach-calendar \
  --source-configuration '{
    "ImageRepository": {
      "ImageIdentifier": "YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/coach-calendar:latest",
      "ImageConfiguration": {
        "Port": "8080",
        "RuntimeEnvironmentVariables": {
          "PORT": "8080"
        }
      },
      "ImageRepositoryType": "ECR"
    },
    "AutoDeploymentsEnabled": true
  }' \
  --instance-configuration '{
    "Cpu": "1 vCPU",
    "Memory": "2 GB"
  }' \
  --region us-east-1
```

## Troubleshooting

### Build Failures

1. **"CGO not enabled" or SQLite errors**:
   - Ensure `apprunner.yaml` includes the pre-build step to install build dependencies
   - Verify CGO_ENABLED=1 is set in build command

2. **Module errors**:
   ```bash
   # Locally test the build
   CGO_ENABLED=1 go build -o coach-calendar .
   ```

3. **Check build logs**:
   - In App Runner console, go to your service → Logs
   - View deployment logs for detailed error messages

### Runtime Issues

1. **Service failing health checks**:
   - Verify the app is listening on the correct port (8080)
   - Check that `/health` endpoint returns 200 OK
   - Review application logs

2. **Database issues**:
   - SQLite database is created on first run
   - Database persists during the container lifecycle
   - **Important**: Database is lost when service scales down to 0 or redeployment
   - For production, consider using RDS or DynamoDB

3. **Email not sending**:
   - Verify SES sender email is verified
   - Check IAM role has SES permissions
   - Review application logs for SES errors

### Viewing Logs

```bash
# Via CLI
aws logs tail /aws/apprunner/coach-calendar/SERVICE_ID/application --follow

# Or in Console
# App Runner → Your Service → Logs tab
```

## Environment Variables

Set these in App Runner service configuration:

### Required
- `PORT=8080` (usually automatically set)

### Optional (Email via AWS SES)
- `USE_AWS_SES=true`
- `SMTP_FROM=verified@yourdomain.com`
- `AWS_REGION=us-east-1`

### Optional (Email via SMTP)
- `SMTP_HOST=smtp.gmail.com`
- `SMTP_PORT=587`
- `SMTP_FROM=your-email@gmail.com`
- `SMTP_PASSWORD=your-app-password`

## Cost Estimate

**App Runner Pricing** (us-east-1):
- Provisioned compute: $0.064/vCPU-hour + $0.007/GB-hour
- Active requests: $0.064/vCPU-hour + $0.007/GB-hour
- Build: $0.005/build minute

**Estimated monthly cost** (1 vCPU, 2 GB, always on):
- ~$46-50/month for always-on service
- Less if service scales to zero during low traffic

**Free Tier**:
- 2,000 build minutes per month
- Provisioned container instances included in compute pricing

## Custom Domain

1. **In App Runner Console**:
   - Go to your service → Custom domains
   - Click "Link domain"
   - Enter your domain (e.g., `calendar.yourdomain.com`)
   - Follow DNS validation steps

2. **Add DNS Records** in your domain registrar:
   ```
   Type: CNAME
   Name: calendar (or your subdomain)
   Value: xxxxx.us-east-1.awsapprunner.com

   Type: TXT (for validation)
   Name: _apprunner-xxxxx
   Value: provided by App Runner
   ```

## Automatic Deployments

With GitHub integration:
- **Push to branch** → Automatic deployment
- **Pull request merge** → Automatic deployment
- **Manual trigger** → Via App Runner console

To disable auto-deploy:
```bash
aws apprunner update-service \
  --service-arn YOUR_SERVICE_ARN \
  --auto-deployments-enabled false
```

## Best Practices

1. **Use GitHub integration** for easiest deployment
2. **Enable auto-deployment** for continuous delivery
3. **Set up health checks** at `/health`
4. **Use IAM roles** instead of access keys
5. **Monitor logs** in CloudWatch
6. **Set up alarms** for failed health checks
7. **Consider database persistence** - use RDS for production
8. **Verify SES sender** before enabling email

## Support

For issues:
- Check AWS App Runner documentation
- Review CloudWatch logs
- Test locally with Docker
- Verify GitHub connection is active
