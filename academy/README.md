# Summer Academy Platform

A comprehensive educational platform for learning DSA, Linux commands, and software development through daily challenges and interactive terminal sessions.

## Features

- Daily challenges in three categories: DSA, Linux, and Build
- Interactive terminal sessions via WBFY integration
- Leaderboard and progress tracking
- Telegram-based authentication
- PostgreSQL database for storing user data, problems, and submissions
- Redis for OTP storage

## Architecture

- Backend: Go with Gin framework
- Frontend: HTML templates with HTMX for dynamic content
- Terminal: WBFY integration for interactive terminal sessions
- Database: PostgreSQL for data storage
- Authentication: Telegram bot with JWT tokens
- Caching: Redis for OTP storage

## Directory Structure

```
academy/
├── cmd/
│   └── main.go                    # Entry point
├── internal/
│   ├── auth/                      # Authentication logic
│   ├── config/                    # Configuration
│   ├── database/                  # Database connection and queries
│   ├── handlers/                  # HTTP handlers
│   ├── middleware/                # Middleware components
│   ├── models/                    # Data models
│   ├── problem/                   # Problem management
│   ├── submission/                # Submission processing
│   └── telegrambot/               # Telegram bot integration
├── web/
│   ├── static/                    # Static assets (CSS, JS)
│   └── templates/                 # HTML templates
└── problems/                      # Problem content and test cases
```

## Problem Structure

Each problem is stored in a markdown file with metadata and test cases:

```
problems/
└── day-1/
    ├── dsa.md                     # Problem description
    ├── linux.md                   # Problem description
    ├── build.md                   # Problem description
    ├── dsa.json                   # Test cases
    └── metadata.json              # Problem metadata
```

## Authentication Flow

The platform uses Telegram bot for authentication:

1. User visits the login page and is directed to the Telegram bot
2. The bot asks the user to share their phone number
3. The bot requests the user's full name
4. The bot generates a 6-digit OTP and sends it with a login link
5. User clicks the link or enters the code on the verification page
6. The server verifies the OTP against Redis
7. Upon successful verification, a JWT token is issued
8. The user is logged in and redirected to the platform

## Database Schema

- **users**: User information and authentication
- **problems**: Challenge details and metadata
- **submissions**: User submissions and results

## Integration with WBFY

The Summer Academy platform integrates with WBFY (Web Browser For Your CLI) to provide interactive terminal sessions:

1. User clicks "Open in Terminal" button on a problem page
2. Academy backend creates a new terminal session with appropriate environment
3. User is redirected to terminal page where they can interact with the CLI
4. Terminal session communicates with the WBFY backend via WebSockets

## Getting Started

### Prerequisites

- Go 1.24.4
- PostgreSQL
- Redis (for OTP storage)
- WBFY binary
- Telegram Bot Token

### Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/summer-academy.git
cd summer-academy
```

2. Install dependencies:
```bash
cd academy
go mod download
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env to set your Telegram bot token and other configurations
```

4. Set up Telegram Bot:
```bash
# Create a bot with @BotFather on Telegram
# Set the following commands for your bot:
/start - Start the bot
/login - Begin the login process
```

5. Create database:
```bash
createdb academy
psql -d academy -f scripts/schema.sql
```

6. Run the application:
```bash
go run cmd/main.go
```

The application will be available at `http://localhost:8080`.

## API Endpoints

### Public Routes
- `GET /` - Home page
- `GET /login` - Login page
- `GET /verify` - OTP verification page
- `POST /login` - Process login with OTP
- `GET /leaderboard` - Public leaderboard

### Authenticated Routes
- `GET /days` - List all days with problems
- `GET /days/:day` - Details for a specific day
- `GET /problems/:slug` - Problem detail
- `GET /submit/:slug` - Submission page
- `POST /submit/:slug` - Process submission
- `POST /test/:slug` - Test submission
- `GET /profile` - User profile
- `POST /profile` - Update profile
- `POST /terminal/:slug` - Create terminal session
- `GET /terminal/:id` - Terminal session page

### Admin Routes
- `GET /admin` - Admin dashboard
- `GET /admin/users` - Manage users
- `GET /admin/problems` - Manage problems
- `POST /admin/problems` - Create problem
- `PUT /admin/problems/:id` - Update problem

## License

[MIT License](LICENSE)
