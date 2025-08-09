# SnapShell Deployment Guide 🚀

## Quick Deploy Options

### 🏆 **Option 1: Railway (Recommended)**

**Why Railway?**
- Built-in Redis (no extra cost)
- $5/month free credits
- Simple one-command deployment
- Easy migration to AWS later

**Deploy Steps:**
```bash
# 1. Install Railway CLI
npm install -g @railway/cli

# 2. Login and deploy
railway login
railway init
railway add redis  # Adds Redis service
railway deploy     # Deploys signaling server

# 3. Get your URL
railway domain     # Shows your deployment URL
```

**Environment Variables (Auto-configured):**
- `REDIS_URL` - Automatically set by Railway
- `PORT` - Automatically set by Railway

### 🥈 **Option 2: Render (Free)**

**Steps:**
1. Go to [render.com](https://render.com)
2. Connect your GitHub repository 
3. Create **Web Service** from your repo
4. Create **Redis** service
5. Link Redis URL to web service

**Manual Environment Variables:**
- `REDIS_URL` - Copy from Render Redis service
- `PORT` - Set to `10000` (Render default)

### 🥉 **Option 3: Heroku (Traditional)**

```bash
heroku create snapshell-signaler --stack container
heroku addons:create heroku-redis:mini  # $15/month
git push heroku main
```

## 🔧 Production Configuration

### Environment Variables Needed:
```bash
REDIS_URL=redis://user:pass@host:port/db  # Redis connection
PORT=8080                                  # Server port (auto-set by platforms)
```

### Health Check Endpoints:
- `GET /health` - Service health + Redis connectivity
- `GET /` - Service information and available endpoints

## 📱 Client Configuration

After deploying, users connect like this:

```bash
# Download client
go install github.com/saswatsam786/snapshell/cmd@latest

# Connect to your deployed signaling server
snapshell -signaled-o --room myroom --server https://your-app.railway.app
snapshell -signaled-a --room myroom --server https://your-app.railway.app
```

## 🔄 Migration Path to AWS

When ready for AWS:

1. **Deploy signaling server** to AWS ECS/Fargate
2. **Use AWS ElastiCache** for Redis
3. **Update environment variables**:
   ```bash
   REDIS_URL=redis://your-elasticache-endpoint:6379
   ```
4. **No client changes needed** - just update the `--server` URL

## 🎯 **My Recommendation**

**Start with Railway** because:
- ✅ Free tier covers your initial needs
- ✅ Built-in Redis (no separate setup)
- ✅ Excellent Go support
- ✅ Easy to migrate later
- ✅ Better performance than free alternatives

**Deploy Command:**
```bash
railway login
railway init
railway add redis
railway deploy
```

That's it! Your signaling server will be live at `https://your-app.railway.app`
