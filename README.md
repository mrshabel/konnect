# Konnect

A location-based matchmaking platform for all personalities with similar interests as you

## Features

-   Location-based profile discovery
-   Real-time message
-   Match-making

## Prerequisites

-   Go 1.23 or higher
-   Docker
-   Cloudinary account (for photo uploads)
-   Google OAuth 2.0 credentials

## Installation

1. **Clone the repository**

```bash
git clone https://github.com/mrshabel/konnect.git
cd konnect
```

2. **Install dependencies**

```bash
go mod download
```

3. **Set up environment variables**

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```env
# database
DB_HOST=host.docker.internal
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=konnect


# auth
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_CALLBACK_URL=http://localhost:8080/api/auth/google/callback
JWT_EXPIRY_MINUTES=1440
JWT_SECRET= #generate with 'openssl rand -hex 32'
SESSION_SECRET= #generate with 'openssl rand -hex 32'

# cloudinary
CLOUDINARY_URL=cloudinary_url

# server
PORT=8080
```

4. **Start the application**

```bash
docker compose up --watch
```

## Usage

View the API documentation on:

```
http://localhost:8000/api/docs/index.html
```
