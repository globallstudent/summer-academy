# Summer Academy

A production-ready educational platform for self-learning Linux, Data Structures & Algorithms (DSA), and build skills. The platform includes a Go Gin backend, HTMX frontend, PostgreSQL database, Redis cache, and WBFY terminal integration.

## Features

- Daily unlocked challenges across multiple tracks (Linux, DSA, Build)
- Interactive in-browser terminal powered by WBFY
- Secure authentication via Telegram bot
- Leaderboard and progress tracking
- Isolated per-user environments for challenges
- Docker-based containerized architecture

## Tech Stack

- **Backend**: Go with Gin framework
- **Frontend**: HTML/CSS/JS with HTMX for dynamic content
- **Database**: PostgreSQL
- **Cache**: Redis
- **Terminal**: WBFY (Web-based Terminal) with Docker containers
- **Authentication**: Telegram Bot API
- **Deployment**: Docker and Docker Compose

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.24 or later (for development)
- PostgreSQL client (for development)
- Docker Hub account (for pushing images)

### Environment Setup

1. Copy the example environment file:

```bash
cp academy/.env.example academy/.env
```

2. Update the environment variables in `academy/.env` as needed.

### Running with Docker Compose

The easiest way to run the entire stack:

```bash
docker-compose up -d
```

This will start:
- The Academy backend server
- PostgreSQL database
- Redis cache
- WBFY terminal service

### Development Mode

For development, you can run:

```bash
scripts/run-dev.sh
```

This will start PostgreSQL and Redis in Docker, but run the Academy backend locally.

## Docker Images

The platform uses the following Docker images:

- `globalstudent/academy:latest` - Main application server
- `globalstudent/wbfy:latest` - Web-based terminal server
- `globalstudent/wbfy-base:latest` - Base image for terminal environments
- `globalstudent/wbfy-python:latest` - Python terminal environment
- `globalstudent/wbfy-golang:latest` - Go terminal environment
- `globalstudent/wbfy-node:latest` - Node.js terminal environment

### Building and Pushing Images

To build and push all images to Docker Hub:

```bash
scripts/build-images.sh
```

## Project Structure

```
academy/             # Main application directory
├── cmd/             # Entry points
├── internal/        # Internal packages
│   ├── auth/        # Authentication
│   ├── config/      # Configuration
│   ├── database/    # Database access
│   ├── handlers/    # HTTP handlers
│   ├── middleware/  # Middleware
│   ├── models/      # Data models
│   └── telegrambot/ # Telegram bot integration
├── web/             # Web assets
│   ├── static/      # Static assets
│   └── templates/   # HTML templates
docker/              # Docker image definitions
├── wbfy-base/       # Base terminal image
├── wbfy-python/     # Python environment
├── wbfy-golang/     # Go environment
└── wbfy-node/       # Node.js environment
problems/            # Challenge content
├── day1/            # Day 1 problems
├── day2/            # Day 2 problems
└── ...              # Additional days
scripts/             # Utility scripts
wbfy/                # Terminal service
```

## Authentication Flow

1. User enters phone number on login page
2. Telegram bot sends OTP to user
3. User enters OTP on verification page
4. On success, JWT token is issued and stored in a cookie
5. All authenticated routes check the JWT token

## Adding New Challenges

1. Create a new directory under `problems/` for the day
2. Add problem markdown files and any supporting files
3. Create a `metadata.json` file with problem details and unlock date

## Terminal Integration

The platform uses WBFY for terminal integration:

1. When a user opens a terminal, a Docker container is launched
2. Problem files are copied to the container
3. User interacts with the container via WebSocket
4. Session is cleaned up after user closes terminal or timeout

## License

This project is licensed under the MIT License - see the LICENSE file for details.