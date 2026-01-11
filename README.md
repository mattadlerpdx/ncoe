# NCOE Case Management System

Nevada Commission on Ethics - Case Management System

## Overview

This is a web-based case management system for the Nevada Commission on Ethics, built with:

- **Backend:** Go 1.21+ (standard library)
- **Frontend:** HTMX + Bootstrap 5.3
- **Database:** PostgreSQL (with mock mode for demos)

## Features

### Public Interface (No Login Required)
- Submit Advisory Opinion Requests
- File Ethics Complaints
- File Ethics Acknowledgments
- Submit Public Records Requests
- Search Published Opinions & Orders

### Staff Portal (Login Required)
- Dashboard with KPIs
- Case management (view, assign, update status)
- Deadline tracking with reminders
- Document management
- Reporting and analytics

## Quick Start

### Demo Mode (No Database)

```bash
# Clone/copy the project
cd ncoe

# Run the server (uses mock data)
go run cmd/server/main.go

# Open browser to http://localhost:8080
```

In demo mode, any credentials work for staff login.

### Production Mode

```bash
# Set environment variables
export DATABASE_URL="postgres://user:pass@localhost/ncoe?sslmode=disable"
export ENVIRONMENT="production"
export SESSION_SECRET="your-32-char-secret-here"

# Run migrations (not yet implemented)
# psql $DATABASE_URL < migrations/001_initial.sql

# Run the server
go run cmd/server/main.go
```

## Project Structure

```
ncoe/
├── cmd/server/main.go          # Entry point
├── config/
│   └── branding.yaml           # Agency branding config
├── internal/
│   ├── config/                 # Configuration loading
│   ├── domain/                 # Domain models
│   ├── handler/                # HTTP handlers
│   ├── middleware/             # Auth, logging, recovery
│   ├── repository/
│   │   ├── mock/               # In-memory mock repos
│   │   └── postgres/           # PostgreSQL repos
│   └── service/                # Business logic
├── migrations/                 # SQL migrations
├── static/
│   ├── css/custom.css
│   ├── js/app.js
│   └── img/
├── templates/
│   ├── auth/                   # Login pages
│   ├── public/                 # Public submission forms
│   ├── staff/                  # Staff portal pages
│   └── partials/               # Reusable components
├── go.mod
├── go.sum
└── README.md
```

## Case Types

| Code | Type | Deadline |
|------|------|----------|
| AO | Advisory Opinion Request | 45 business days |
| EC | Ethics Complaint | Investigation timeline |
| EA | Ethics Acknowledgment | N/A |
| PRR | Public Records Request | 5 business days |

## User Roles

| Role | Access |
|------|--------|
| Admin | Full system access |
| Commission Counsel | All cases, publishing |
| Staff Attorney | Assigned cases |
| Investigator | Complaint cases |
| Admin Staff | Case intake, PRRs |
| Read Only | View only |

## Configuration

Branding is configured via `config/branding.yaml`:

```yaml
agency_name: "Nevada Commission on Ethics"
short_name: "NCOE"
primary_color: "#003366"
contact_email: "ncoe@ethics.nv.gov"
contact_phone: "(775) 687-5469"
```

## Development

```bash
# Install dependencies
go mod download

# Run with hot reload (using air)
air

# Build binary
go build -o ncoe ./cmd/server/main.go

# Run tests
go test ./...
```

## Technology Stack

Based on the MyFlow/GovFlow architecture:

- **No JavaScript frameworks** - Bootstrap + vanilla JS only
- **No build pipeline** - No webpack, no transpilers
- **HTMX for interactivity** - SPA-like navigation without complexity
- **Server-side rendering** - Go html/template
- **Session-based auth** - HTTPOnly cookies, no JWTs

## License

Proprietary - Nevada Commission on Ethics
