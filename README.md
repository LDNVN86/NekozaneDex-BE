# 🚀 Nekozanedex Backend - High-Performance Web Novel API

**English** | [Tiếng Việt](./README_VI.md)

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-Gonic-008ECF?style=for-the-badge&logo=gin&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-Secure-000000?style=for-the-badge&logo=json-web-tokens&logoColor=white)

Nekozanedex Backend is a robust, production-ready RESTful API built with **Go (Golang)** and the **Gin Gonic** framework. It provides a secure and scalable foundation for the Nekozanedex web novel platform.

---

## 🔥 Key Features

### 🔐 Advanced Security & Auth

- **JWT Authentication**: Secure Access and Refresh token lifecycle.
- **Refresh Token Rotation (RTR)**: Issues new tokens on every refresh to prevent theft.
- **Token Reuse Detection**: Automatically revokes all sessions if a stolen token is detected.
- **CSRF Protection**: Integrated CSRF token management for web safety.
- **Bcrypt Hashing**: Secure password storage using adaptive hashing.
- **Rate Limiting**: Configurable rate limits to prevent abuse.
- **Security Headers**: XSS protection, content-type sniffing prevention.

### 📚 Content Management

- **Story & Chapter CRUD**: Full administrative control over content.
- **Genre Management**: Flexible categorization system.
- **Comments & Replies**: Nested comment system with moderation.
- **Bookmarks**: User bookmark management.
- **Notifications**: Real-time notification system.
- **Advanced Pagination**: Optimized database queries for large data sets.
- **Relational Integrity**: Powered by **GORM** and **PostgreSQL**.

### ⚙️ System Operations

- **Background Cleanup Job**: Autonomous goroutine that purges expired and revoked tokens every 6 hours.
- **Swagger Documentation**: Interactive API documentation for easy integration.
- **Environment-based Config**: Fully configurable via environment variables.
- **WebSocket Support**: Real-time features ready (Centrifugo integration).
- **Image Upload**: Cloudinary integration for image storage.

---

## 🛠️ Tech Stack

| Category           | Technology                                      |
| ------------------ | ----------------------------------------------- |
| **Language**       | [Go (Golang)](https://go.dev/) 1.22+            |
| **Web Framework**  | [Gin Gonic](https://gin-gonic.com/)             |
| **ORM**            | [GORM](https://gorm.io/)                        |
| **Database**       | [PostgreSQL](https://www.postgresql.org/)       |
| **Authentication** | [golang-jwt](https://github.com/golang-jwt/jwt) |
| **Image Storage**  | [Cloudinary](https://cloudinary.com/)           |
| **Realtime**       | [Centrifugo](https://centrifugal.dev/)          |
| **Documentation**  | [swaggo/swag](https://github.com/swaggo/swag)   |

---

## 🚀 Getting Started

### Prerequisites

- Go 1.22 or higher
- PostgreSQL
- Optional: Cloudinary Account (for image uploads)
- Optional: Centrifugo (for real-time features)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/nekozanedex-backend.git
   cd nekozanedex-backend
   ```

2. Download dependencies:

   ```bash
   go mod download
   ```

3. Configure environment variables:

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

The API will be available at [http://localhost:9091](http://localhost:9091).

---

## 📁 Project Structure

```plaintext
nekozanedex-backend/
├── cmd/
│   └── server/           # Application entry point (main.go)
├── docs/                 # Swagger documentation (auto-generated)
├── internal/
│   ├── config/           # Configuration loading from environment
│   ├── database/         # Database connection and initialization
│   ├── handlers/         # HTTP request handlers (Controllers)
│   │   ├── auth_handler.go
│   │   ├── story_handler.go
│   │   ├── chapter_handler.go
│   │   ├── bookmark_handler.go
│   │   ├── comment_handler.go
│   │   ├── notification_handler.go
│   │   ├── upload_handler.go
│   │   └── csrf_handler.go
│   ├── middleware/       # HTTP middleware
│   │   ├── auth.go       # JWT authentication
│   │   ├── cors.go       # CORS configuration
│   │   ├── csrf.go       # CSRF protection
│   │   ├── logger.go     # Request logging
│   │   ├── rate_limit.go # Rate limiting
│   │   └── security.go   # Security headers
│   ├── models/           # Database entities (GORM models)
│   ├── repositories/     # Data access layer
│   ├── routes/           # API route definitions
│   ├── services/         # Business logic layer
│   ├── utils/            # Common utilities (JWT, Bcrypt, Result pattern)
│   └── websocket/        # WebSocket handlers (Centrifugo)
├── pkg/                  # Shared packages
├── .env.example          # Environment variables template
├── go.mod
└── go.sum
```

---

## ⚙️ Environment Variables

| Variable                    | Description                                          | Default                           |
| --------------------------- | ---------------------------------------------------- | --------------------------------- |
| **App**                     |                                                      |                                   |
| `APP_ENV`                   | Environment (`development`, `staging`, `production`) | `development`                     |
| `PORT`                      | Server port                                          | `9091`                            |
| `GIN_MODE`                  | Gin mode (`debug`, `release`)                        | `debug`                           |
| **Database**                |                                                      |                                   |
| `DB_HOST`                   | PostgreSQL host                                      | `localhost`                       |
| `DB_PORT`                   | PostgreSQL port                                      | `5432`                            |
| `DB_USER`                   | Database user                                        | `postgres`                        |
| `DB_PASSWORD`               | Database password                                    | -                                 |
| `DB_NAME`                   | Database name                                        | `nekozanedex`                     |
| **JWT**                     |                                                      |                                   |
| `JWT_ACCESS_SECRET`         | Access token secret                                  | -                                 |
| `JWT_REFRESH_SECRET`        | Refresh token secret                                 | -                                 |
| `JWT_ACCESS_EXPIRE_MINUTES` | Access token expiry (minutes)                        | `30`                              |
| `JWT_REFRESH_EXPIRE_DAYS`   | Refresh token expiry (days)                          | `7`                               |
| **Cookie**                  |                                                      |                                   |
| `JWT_COOKIE_DOMAIN`         | Cookie domain                                        | -                                 |
| `JWT_COOKIE_SAME_SITE`      | SameSite policy                                      | `lax`                             |
| `JWT_COOKIE_MAX_AGE`        | Cookie max age (seconds)                             | `604800`                          |
| **CORS**                    |                                                      |                                   |
| `CORS_DEV_ORIGINS`          | Development allowed origins                          | `http://localhost:3000,...`       |
| `CORS_PROD_ORIGINS`         | Production allowed origins                           | `https://nekozanedex.com,...`     |
| `CORS_STAGING_ORIGINS`      | Staging allowed origins                              | `https://staging.nekozanedex.com` |
| **Security**                |                                                      |                                   |
| `CSRF_SECRET_KEY`           | CSRF token secret                                    | -                                 |
| `FRAME_ANCESTORS`           | frame-ancestors CSP directive                        | `'self'`                          |
| **Cloudinary**              |                                                      |                                   |
| `CLOUDINARY_CLOUD_NAME`     | Cloudinary cloud name                                | -                                 |
| `CLOUDINARY_API_KEY`        | Cloudinary API key                                   | -                                 |
| `CLOUDINARY_API_SECRET`     | Cloudinary API secret                                | -                                 |
| **Centrifugo**              |                                                      |                                   |
| `CENTRIFUGO_URL`            | Centrifugo server URL                                | `http://localhost:8000`           |
| `CENTRIFUGO_API_KEY`        | Centrifugo API key                                   | -                                 |

---

## 📖 API Documentation

Once the server is running, access the Swagger UI at:

```
http://localhost:9091/swagger/index.html
```

### Main API Endpoints

| Method | Endpoint                              | Description          |
| ------ | ------------------------------------- | -------------------- |
| `POST` | `/api/auth/register`                  | User registration    |
| `POST` | `/api/auth/login`                     | User login           |
| `POST` | `/api/auth/refresh`                   | Refresh access token |
| `POST` | `/api/auth/logout`                    | User logout          |
| `GET`  | `/api/stories`                        | List stories         |
| `GET`  | `/api/stories/:slug`                  | Get story details    |
| `GET`  | `/api/stories/:slug/chapters`         | List chapters        |
| `GET`  | `/api/stories/:slug/chapters/:number` | Get chapter content  |
| `GET`  | `/api/genres`                         | List genres          |
| `GET`  | `/api/csrf-token`                     | Get CSRF token       |

---

## 🛡️ Security Implementation

### Refresh Token Rotation (RTR) with Reuse Detection

1. When a client refreshes an `access_token`, the `refresh_token` is also rotated.
2. If an old `refresh_token` is used again (potential replay attack), the system identifies the reuse.
3. Upon detection, **all active tokens** for that user are immediately revoked, forcing a full re-authentication.

### CORS Configuration

CORS origins are configurable via environment variables:

- **Development**: Allows localhost origins
- **Staging**: Includes staging domain
- **Production**: Restricts to production domains only

---

## 🤝 Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## 📄 License

This project is licensed under the MIT License.
