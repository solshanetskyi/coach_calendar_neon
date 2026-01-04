# AWS App Runner Deployment Steps

This guide walks you through deploying the Go web application to AWS App Runner.

## Prerequisites

- AWS Account
- AWS CLI installed and configured
- Docker installed (for local testing)
- GitHub account (for source code repository)

## Option 1: Deploy from Source Code Repository (Recommended)

### Step 1: Push Code to GitHub

```bash
git init
git add .
git commit -m "Initial commit"
git remote add origin <your-github-repo-url>
git push -u origin main
```

### Step 2: Create App Runner Service via AWS Console

1. Navigate to AWS App Runner in the AWS Console
2. Click "Create service"
3. Choose "Source code repository"
4. Click "Add new" to connect your GitHub account
5. Select your repository and branch (e.g., `main`)

### Step 3: Configure Build Settings

- **Runtime**: Managed runtime
- **Runtime**: Go 1
- **Build command**: `go build -o main .`
- **Start command**: `./main`
- **Port**: 8080

### Step 4: Configure Service Settings

- **Service name**: Choose a name (e.g., `go-web-app`)
- **Virtual CPU**: 1 vCPU (or as needed)
- **Memory**: 2 GB (or as needed)
- **Environment variables**: Add if needed
- **Auto scaling**: Configure as needed (default: 1-25 instances)

### Step 5: Health Check

- **Health check protocol**: HTTP
- **Health check path**: `/health`
- **Interval**: 5 seconds
- **Timeout**: 2 seconds
- **Healthy threshold**: 1
- **Unhealthy threshold**: 5

### Step 6: Review and Create

1. Review all settings
2. Click "Create & deploy"
3. Wait for deployment to complete (5-10 minutes)

## Option 2: Deploy from Container Registry (ECR)

### Step 1: Create ECR Repository

```bash
aws ecr create-repository --repository-name go-web-app --region us-east-1
```

### Step 2: Build and Push Docker Image

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <aws-account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build the image
docker build -t go-web-app .

# Tag the image
docker tag go-web-app:latest <aws-account-id>.dkr.ecr.us-east-1.amazonaws.com/go-web-app:latest

# Push to ECR
docker push <aws-account-id>.dkr.ecr.us-east-1.amazonaws.com/go-web-app:latest
```

### Step 3: Create App Runner Service from ECR

1. Navigate to AWS App Runner in the AWS Console
2. Click "Create service"
3. Choose "Container registry"
4. Select "Amazon ECR"
5. Click "Browse" and select your image
6. Choose "Automatic" deployment trigger

### Step 4: Configure Deployment Settings

- **ECR access role**: Create new role or use existing
- **Service name**: Choose a name
- **Virtual CPU & Memory**: Configure as needed
- **Port**: 8080

### Step 5: Configure Health Check

Same as Option 1, Step 5

### Step 6: Create Service

Click "Create & deploy"

## Option 3: Deploy using AWS CLI

### Create App Runner Service from ECR

```bash
aws apprunner create-service \
  --service-name go-web-app \
  --source-configuration '{
    "ImageRepository": {
      "ImageIdentifier": "<aws-account-id>.dkr.ecr.us-east-1.amazonaws.com/go-web-app:latest",
      "ImageRepositoryType": "ECR",
      "ImageConfiguration": {
        "Port": "8080"
      }
    },
    "AutoDeploymentsEnabled": true
  }' \
  --instance-configuration '{
    "Cpu": "1 vCPU",
    "Memory": "2 GB"
  }' \
  --health-check-configuration '{
    "Protocol": "HTTP",
    "Path": "/health",
    "Interval": 5,
    "Timeout": 2,
    "HealthyThreshold": 1,
    "UnhealthyThreshold": 5
  }' \
  --region us-east-1
```

## Post-Deployment

### Access Your Application

After deployment completes, App Runner provides a default domain:
```
https://<random-id>.us-east-1.awsapprunner.com
```

### Monitor Your Application

1. Go to AWS App Runner Console
2. Select your service
3. View:
   - Service status
   - Logs (CloudWatch Logs)
   - Metrics
   - Activity history

### Configure Custom Domain (Optional)

1. In App Runner service, go to "Custom domains"
2. Click "Link domain"
3. Enter your domain name
4. Add the provided CNAME records to your DNS provider
5. Wait for validation

## Local Testing

Before deploying, test locally:

```bash
# Run directly
go run main.go

# Or build and run
go build -o main .
./main

# Or use Docker
docker build -t go-web-app .
docker run -p 8080:8080 go-web-app
```

Visit `http://localhost:8080` to verify the application works.

## Costs

AWS App Runner pricing:
- **Compute**: Pay per vCPU and memory used
- **Requests**: Pay per request
- Free tier: Limited compute hours per month

Check current pricing at: https://aws.amazon.com/apprunner/pricing/

## Troubleshooting

### Deployment Fails

- Check CloudWatch Logs for build/runtime errors
- Verify Dockerfile builds successfully locally
- Ensure PORT environment variable is handled correctly

### Health Check Fails

- Verify `/health` endpoint returns 200 OK
- Check that port 8080 is exposed and listening
- Increase timeout if application is slow to start

### Application Not Accessible

- Check service status in App Runner console
- Verify security settings allow public access
- Check CloudWatch Logs for runtime errors
