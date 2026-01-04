# AWS App Runner - Quick Start Guide

This is a simplified guide to deploy Coach Calendar to AWS App Runner using Docker.

## What You Need

- AWS Account
- GitHub repository: `https://github.com/solshanetskyi/coach_calendar`
- Branch: `main`

## Deployment Steps (5 minutes)

### 1. Open AWS App Runner Console

Visit: https://console.aws.amazon.com/apprunner

### 2. Create Service

Click **"Create service"**

### 3. Configure Source

**Source:**
- Repository type: `Source code repository`
- Click **"Add new"** to connect GitHub
- Authorize AWS Connector for GitHub
- Select repository: `coach_calendar`
- Branch: `main`
- Deployment trigger: `Automatic`

### 4. Configure Build

**Build settings:**
- Configuration source: `Use a configuration file`
- Configuration file: `apprunner.yaml`

The apprunner.yaml is already configured to use Docker, which handles CGO and SQLite correctly.

### 5. Configure Service

**Service settings:**
- Service name: `coach-calendar` (or your choice)
- Port: `8080` (already set in apprunner.yaml)
- CPU: `1 vCPU`
- Memory: `2 GB`

**Auto scaling (optional):**
- Min instances: `1`
- Max instances: `25`
- Max concurrency: `100`

### 6. Configure Health Check

- Protocol: `HTTP`
- Path: `/health`
- Interval: `10` seconds
- Timeout: `5` seconds
- Unhealthy threshold: `3`
- Healthy threshold: `3`

### 7. Environment Variables (Optional - Email)

**For AWS SES:**
```
USE_AWS_SES=true
SMTP_FROM=your-verified-email@domain.com
AWS_REGION=us-east-1
```

**OR for SMTP:**
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_FROM=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### 8. Security - IAM Role (Optional - for SES)

If using AWS SES for emails:
- Create or select an instance role
- Attach this policy:

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

### 9. Review and Deploy

- Review all settings
- Click **"Create & deploy"**
- Wait 5-10 minutes for deployment

### 10. Access Your Application

Once deployed, you'll get a URL like:
```
https://xxxxx.us-east-1.awsapprunner.com
```

## How It Works

### apprunner.yaml
```yaml
version: 1.0
runtime: docker
build:
  dockerfile: Dockerfile
run:
  command: ./main
  network:
    port: 8080
```

### Why Docker?

- ✅ Handles CGO compilation correctly
- ✅ Installs build dependencies (gcc, musl-dev, sqlite-dev)
- ✅ Multi-stage build keeps final image small
- ✅ Reliable and consistent builds

### The Dockerfile

1. **Build stage**: Uses `golang:1.21-alpine` with build tools
2. **Compiles** with CGO enabled for SQLite
3. **Runtime stage**: Uses minimal `alpine:latest`
4. **Final image**: Only ~20MB with the compiled binary

## Testing Your Deployment

### Check Health
```bash
curl https://YOUR-APPRUNNER-URL/health
# Should return: OK
```

### View Application
Open in browser: `https://YOUR-APPRUNNER-URL/`

### View Admin Panel
Open in browser: `https://YOUR-APPRUNNER-URL/admin`

## Troubleshooting

### Build Failed
- Check build logs in App Runner console
- Verify Dockerfile is in repository
- Ensure apprunner.yaml points to Dockerfile

### Application Not Starting
- Check application logs
- Verify health check path is `/health`
- Ensure port is 8080

### Email Not Working
- Verify environment variables are set
- Check IAM role has SES permissions (if using SES)
- Verify sender email is verified in SES console
- Check application logs for email errors

## Updating the Application

App Runner automatically deploys when you push to the `main` branch:

```bash
git add .
git commit -m "Update application"
git push origin main
```

App Runner will:
1. Detect the new commit
2. Build the Docker image
3. Deploy the new version
4. Automatically handle traffic switching

## Cost Estimate

**App Runner pricing (us-east-1):**
- Compute: ~$46-50/month for always-on (1 vCPU, 2 GB)
- Build: $0.005/build minute (usually ~2-3 minutes)
- Free tier: 2,000 build minutes/month

## Custom Domain (Optional)

1. Go to App Runner → Your Service → Custom domains
2. Click "Link domain"
3. Enter your domain (e.g., `calendar.yourdomain.com`)
4. Add DNS records to your domain:
   - CNAME: `calendar` → `xxxxx.us-east-1.awsapprunner.com`
   - TXT: `_apprunner-xxxxx` → (validation value)

## Support

- App Runner logs: App Runner Console → Your Service → Logs
- Application works locally? Test with: `docker build -t coach-calendar . && docker run -p 8080:8080 coach-calendar`
- Check GitHub actions if you have CI/CD

## Summary

Your Coach Calendar application is now:
- ✅ Automatically deployed on every push
- ✅ Running with CGO/SQLite support
- ✅ Accessible via HTTPS
- ✅ Auto-scaling based on traffic
- ✅ Monitored with health checks

Enjoy your deployment! 🚀
