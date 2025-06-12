# unRAID Template Setup

## Option 1: Use Individual Templates (Recommended)

### Setup Steps:

1. **Create Custom Network:**

   ```bash
   docker network create expenses-network
   ```

2. **Install in Order:**

   - Install `expenses-database` first
   - Install `expenses-backend` second
   - Install `expenses-frontend` last

3. **Add Templates to Community Applications:**
   - Copy template URLs to unRAID template repository
   - Or manually add via Docker → Add Container → Template

### Template URLs:

- Database: `https://raw.githubusercontent.com/nathanaelcunningham/expenses/main/unraid-templates/expenses-database.xml`
- Backend: `https://raw.githubusercontent.com/nathanaelcunningham/expenses/main/unraid-templates/expenses-backend.xml`
- Frontend: `https://raw.githubusercontent.com/nathanaelcunningham/expenses/main/unraid-templates/expenses-frontend.xml`

## Option 2: Manual Docker Compose (Alternative)

If you prefer docker-compose, you can still use the provided `docker-compose.yml` file:

1. Copy `docker-compose.yml` to `/mnt/user/appdata/expenses/`
2. SSH into unRAID: `docker-compose up -d`

## Configuration

### Port Mapping:

- Frontend will be accessible on the port you specify (default: 3000)
- Backend and database communicate internally via the custom network

### Environment Variables:

All database credentials are configurable through the template variables.

### Data Persistence:

Database data persists in `/mnt/user/appdata/expenses/database`

## Updating Versions:

1. Edit container → Change repository tag from `:latest` to `:1.2.3`
2. Update container

