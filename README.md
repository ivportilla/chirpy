# Chirpy - A Twitter-like Web Application

Chirpy is a lightweight social media web application that allows users to post short messages called "chirps" (similar to tweets) with a maximum length of 140 characters.
This is part of the Servers course in boot.dev

## Features

- User authentication using JWT tokens
- Create and manage user accounts
- Post and delete chirps
- View all chirps or specific chirps by ID
- User session management with refresh tokens
- Premium user upgrades via Polka webhooks
- Admin metrics and reset capabilities
- Static file serving for frontend application

## API Endpoints

### Authentication
- `POST /api/users` - Create a new user account
- `POST /api/login` - User login and receive JWT token
- `POST /api/refresh` - Refresh expired JWT tokens
- `POST /api/revoke` - Revoke refresh tokens
- `PUT /api/users` - Update user information (authenticated)

### Chirps
- `POST /api/chirps` - Create a new chirp (authenticated)
- `GET /api/chirps` - Get all chirps
  Query params:
  - sort - DESC or ASC (optional)
  - author_id - ID of the chirps author you wanna fetch (optional)
- `GET /api/chirps/{chirpID}` - Get a specific chirp
- `DELETE /api/chirps/{chirpID}` - Delete a specific chirp (authenticated)

### Admin & Metrics
- `GET /admin/metrics` - View application metrics
- `POST /admin/reset` - Reset application state
- `GET /api/healthz` - Health check endpoint

### Webhooks
- `POST /api/polka/webhooks` - Handle premium user upgrades

### Frontend
- `/app/*` - Serves static files for the web application

## Authentication

The application uses JWT (JSON Web Tokens) for authentication:
1. Users can obtain tokens via the login endpoint
2. Access tokens are required for protected endpoints
3. Refresh tokens are available for extended sessions
4. Tokens can be revoked for security purposes
