# Floe CMS

![Floe CMS Logo](https://via.placeholder.com/200x80?text=Floe+CMS)

Floe CMS is a lightweight, production-ready Content Management System written in Go. It's designed to be deployed as a single binary with all features included, making it easy to install and maintain.

## Features

- **Single Binary Deployment**: All components, including the admin UI, are embedded in a single executable.
- **Multiple Database Support**: SQLite (default, embedded), MySQL, and PostgreSQL.
- **RESTful API**: Complete API for managing content and media.
- **Markdown-Based Content**: Store and serve content in Markdown format.
- **Role-Based Access Control**: Admin, Editor, and Viewer roles with appropriate permissions.
- **Multi-Tenancy**: Support for multiple workspaces/organizations.
- **Media Management**: Upload, store, and serve images and other media.
- **Modern Admin UI**: Built with React and TailwindCSS.
- **Extensible**: Designed to be extended with additional features.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Binary Installation

Download the latest release from the [Releases page](https://github.com/randilt/floe-cms/releases).

```bash
# Download the binary
curl -L https://github.com/randilt/floe-cms/releases/latest/download/floe-cms-linux-amd64 -o floe-cms

# Make it executable
chmod +x floe-cms

# Run it
./floe-cms
```

### Docker Installation

```bash
docker pull randilt/floe-cms:latest
docker run -p 8080:8080 -v floe-data:/app/data randilt/floe-cms:latest
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/randilt/floe-cms.git
cd floe-cms

# Build the frontend
cd web/admin
npm install
npm run build
cd ../..

# Build the Go binary
go build -o floe-cms

# Run it
./floe-cms
```

## Quick Start

1. **Start the CMS**:

   ```bash
   ./floe-cms
   ```

2. **Access the Admin UI**:
   Open your browser and navigate to `http://localhost:8080`

3. **Login**:

   - Email: `admin@floe.cms`
   - Password: `adminpassword`

4. **Create Content**:

   - Create a content type
   - Add content items
   - Publish your content

5. **Access Your Content**:
   - Via API: `http://localhost:8080/api/content/{workspace}/{slug}`
   - Or build a frontend that consumes the API

## Configuration

Floe CMS can be configured using a configuration file, environment variables, or command-line flags.

### Configuration File

Create a `config.yaml` file in the same directory as the binary:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  graceful_shutdown: 30
  timeouts:
    read: 15
    write: 15
    idle: 60

database:
  type: sqlite # sqlite, mysql, or postgres
  url: floe.db # Used for SQLite
  host: localhost
  port: 5432
  username: floe
  password: floe
  name: floe
  ssl_mode: disable

auth:
  jwt_secret: "" # Will be generated if empty
  access_token_expiry: 900 # 15 minutes
  refresh_token_expiry: 604800 # 7 days
  admin_email: admin@floe.cms
  admin_password: adminpassword
  password_min_length: 8
  rate_limit_requests: 60
  rate_limit_expiry: 60

storage:
  type: local
  uploads_dir: ./uploads

cache:
  type: memory # memory or redis
  redis_url: redis://localhost:6379/0
  ttl: 300 # 5 minutes
```

### Environment Variables

Environment variables override settings in the configuration file. All variables are prefixed with `FLOE_`.

```bash
# Example
export FLOE_SERVER_PORT=9000
export FLOE_DATABASE_TYPE=postgres
export FLOE_DATABASE_URL=postgres://user:pass@localhost:5432/floe
export FLOE_AUTH_JWT_SECRET=your-secret-key
```

### Command Line Flags

Command line flags override both environment variables and configuration file settings.

```bash
# Example
./floe-cms --port 9000 --db-url "postgres://user:pass@localhost:5432/floe" --reset-admin
```

Available flags:

| Flag            | Description                                         |
| --------------- | --------------------------------------------------- |
| `--config`      | Path to configuration file (default: "config.yaml") |
| `--port`        | Override port defined in configuration              |
| `--reset-admin` | Reset admin credentials to those in config          |
| `--db-url`      | Override database URL defined in configuration      |

## Deployment

### Docker Deployment

```bash
# Create a Docker volume for persistent data
docker volume create floe-data

# Run the container
docker run -d \
  --name floe-cms \
  -p 8080:8080 \
  -v floe-data:/app/data \
  -e FLOE_DATABASE_URL=/app/data/floe.db \
  -e FLOE_STORAGE_UPLOADS_DIR=/app/data/uploads \
  randilt/floe-cms:latest
```

### Systemd Service

Create a systemd service file at `/etc/systemd/system/floe-cms.service`:

```ini
[Unit]
Description=Floe CMS
After=network.target

[Service]
Type=simple
User=floe
WorkingDirectory=/opt/floe-cms
ExecStart=/opt/floe-cms/floe-cms
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable floe-cms
sudo systemctl start floe-cms
```

### Nginx Reverse Proxy

Example Nginx configuration:

```nginx
server {
    listen 80;
    server_name cms.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### HTTPS with Let's Encrypt

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d cms.example.com

# Certbot will update your Nginx configuration
```

## API Documentation

### Authentication

#### Login

```
POST /api/auth/login
```

Request:

```json
{
  "email": "admin@floe.cms",
  "password": "adminpassword"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

#### Refresh Token

```
POST /api/auth/refresh
```

Request:

```json
{
  "refresh_token": "eyJhbGc..."
}
```

Response:

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc..."
  }
}
```

#### Logout

```
POST /api/auth/logout
```

Request:

```json
{
  "refresh_token": "eyJhbGc..."
}
```

Response:

```json
{
  "success": true,
  "data": {
    "message": "Logged out successfully"
  }
}
```

### Content

#### List Content

```
GET /api/content?workspace_id=1&limit=10&offset=0
```

Response:

```json
{
  "success": true,
  "data": {
    "contents": [
      {
        "id": 1,
        "title": "Example Post",
        "slug": "example-post",
        "body": "# Example\n\nThis is an example post.",
        "status": "published",
        "author": {
          "id": 1,
          "email": "admin@floe.cms",
          "first_name": "Admin",
          "last_name": "User"
        },
        "content_type": {
          "id": 1,
          "name": "Blog Post",
          "slug": "blog-post"
        },
        "created_at": "2023-01-01T00:00:00Z",
        "updated_at": "2023-01-01T00:00:00Z",
        "published_at": "2023-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "limit": 10,
    "offset": 0
  }
}
```

#### Create Content

```
POST /api/content
```

Request:

```json
{
  "workspace_id": 1,
  "content_type_id": 1,
  "title": "New Post",
  "slug": "new-post",
  "body": "# New Post\n\nThis is a new post.",
  "status": "draft"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": 2,
    "title": "New Post",
    "slug": "new-post",
    "body": "# New Post\n\nThis is a new post.",
    "status": "draft",
    "created_at": "2023-01-02T00:00:00Z",
    "updated_at": "2023-01-02T00:00:00Z"
  }
}
```

### Media

#### Upload Media

```
POST /api/media
```

Request (Form Data):

```
workspace_id: 1
file: [file upload]
name: example.jpg
```

Response:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "example.jpg",
    "file_name": "example.jpg",
    "file_path": "/uploads/2023/01/01/example.jpg",
    "mime_type": "image/jpeg",
    "size": 12345,
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

For complete API documentation, see [API.md](API.md) or the Swagger documentation at `/swagger/index.html` when running the CMS.

## Development

### Prerequisites

- Go 1.19 or higher
- Node.js 16 or higher
- npm or yarn

### Project Structure

```
floe-cms/
├── cmd/
│   └── floe-cms/           # Command line entry point
├── internal/               # Internal packages
│   ├── api/                # API router and handlers
│   ├── auth/               # Authentication and authorization
│   ├── config/             # Configuration management
│   ├── db/                 # Database management
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Data models
│   ├── storage/            # Storage management
│   └── utils/              # Utility functions
├── web/
│   └── admin/              # Admin UI (React)
├── config.yaml             # Sample configuration
├── Dockerfile              # Docker build file
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
└── README.md               # This file
```

### Development Workflow

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/randilt/floe-cms.git
   cd floe-cms
   ```

2. **Set Up the Frontend**:

   ```bash
   cd web/admin
   npm install
   npm run dev
   ```

   This will start the frontend development server at `http://localhost:5173`.

3. **Run the Backend**:
   In a separate terminal:

   ```bash
   go run cmd/floe-cms/main.go
   ```

   This will start the backend at `http://localhost:8080`.

4. **Building for Production**:

   ```bash
   # Build the frontend
   cd web/admin
   npm run build
   cd ../..

   # Build the backend with the embedded frontend
   go build -o floe-cms
   ```

## Troubleshooting

### Common Issues

#### "Too Many Requests" Error

This is caused by the rate limiting middleware. You can increase the limit in the configuration:

```yaml
auth:
  rate_limit_requests: 100 # Increase this value
  rate_limit_expiry: 60
```

#### Database Connection Issues

If you're having trouble connecting to a database, check:

- Database credentials in your configuration
- That the database server is running
- Network connectivity to the database server

For SQLite, ensure the directory where the database file will be created is writable.

#### Missing Frontend Assets

If the admin UI isn't loading, ensure:

- The frontend has been built (`cd web/admin && npm run build`)
- The go:embed directive is correctly pointing to the frontend build directory
- You've rebuilt the Go binary after building the frontend

#### Logging and Debugging

To enable more verbose logging:

```yaml
# In config.yaml
logging:
  level: debug # debug, info, warn, error
```

Or set the environment variable:

```bash
export FLOE_LOGGING_LEVEL=debug
```

## Contributing

We welcome contributions to Floe CMS! Here's how to get started:

1. **Fork the Repository**:
   Fork the repository on GitHub.

2. **Create a Branch**:

   ```bash
   git checkout -b feature/my-feature
   ```

3. **Make Your Changes**:
   Implement your feature or fix.

4. **Run the Tests**:

   ```bash
   go test ./...
   ```

5. **Submit a Pull Request**:
   Push to your fork and submit a pull request to the main repository.

### Code Standards

- Follow Go best practices and coding standards
- Write tests for new features
- Document your code and update documentation as needed

## License

Floe CMS is licensed under the MIT License. See [LICENSE](LICENSE) for more information.

---

## Contact

For questions, issues, or support, please [open an issue](https://github.com/randilt/floe-cms/issues) on GitHub.

---

_Floe CMS - Seamless content management flow_
