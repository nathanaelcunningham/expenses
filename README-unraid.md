# unRAID Deployment Guide

## One-Click Stack Deployment

### Installation:

1. **Apps** → **Install** (or **Docker** → **Add Container**)
2. **Template URL:**
   ```
   https://raw.githubusercontent.com/nathanaelcunningham/expenses/main/unraid-templates/expenses-stack.xml
   ```
3. **Configure:**
   - **Frontend Port:** Set your desired port (default: 3000)
   - **Database Password:** Set secure password
   - **Database Path:** Storage location (default: `/mnt/user/appdata/expenses/database`)
4. **Apply** - All services deploy together automatically

## What Gets Deployed:

- ✅ **PostgreSQL Database** - Data persistence with configurable password
- ✅ **Go Backend** - API service with automatic database connection
- ✅ **React Frontend** - Web UI with nginx serving static files and API proxy
- ✅ **Automatic Networking** - All services can communicate internally

## Access

- **Application:** `http://unraid-ip:YOUR_PORT`
- **WebUI button** available in unRAID Docker tab

## Updating Versions

1. **Edit** the stack in unRAID Docker tab
2. **Advanced View** → Change image tags from `:latest` to `:1.2.3`
3. **Apply** changes

## Stack Management

- **Start/Stop:** Use the stack controls in Docker tab
- **Logs:** Click on individual services to view logs
- **Remove:** Delete the entire stack as one unit

## GitHub Registry Authentication

For private repositories, add to Extra Parameters:

```
--login ghcr.io -u nathanaelcunningham -p YOUR_GITHUB_TOKEN
```

