# 🚀 Nekozanedex Backend - API Đọc Truyện Hiệu Suất Cao

[English version](./README.md) | **Tiếng Việt**

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-Gonic-008ECF?style=for-the-badge&logo=gin&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-Secure-000000?style=for-the-badge&logo=json-web-tokens&logoColor=white)

Nekozanedex Backend là một RESTful API mạnh mẽ, chuẩn production được xây dựng bằng **Go (Golang)** và framework **Gin Gonic**. Dự án cung cấp nền tảng bảo mật và khả năng mở rộng tốt cho hệ thống đọc truyện Nekozanedex.

---

## 🔥 Tính năng Chính

### 🔐 Bảo mật & Xác thực Nâng cao

- **JWT Authentication**: Quản lý vòng đời Access và Refresh token một cách an toàn.
- **Refresh Token Rotation (RTR)**: Cấp token mới sau mỗi lần refresh để ngăn chặn việc đánh cắp session.
- **Phát hiện Tái sử dụng Token**: Tự động hủy toàn bộ các phiên đăng nhập nếu phát hiện token bị đánh cắp.
- **Bảo vệ CSRF**: Tích hợp cơ chế quản lý CSRF token an toàn.
- **Bcrypt Hashing**: Lưu trữ mật khẩu an toàn với thuật toán băm thích ứng.
- **Rate Limiting**: Giới hạn request để chống spam và abuse.
- **Security Headers**: Bảo vệ XSS, ngăn chặn content-type sniffing.

### 📚 Quản lý Nội dung

- **CRUD Truyện & Chương**: Toàn quyền quản trị nội dung truyện.
- **Quản lý Thể loại**: Hệ thống phân loại linh hoạt.
- **Bình luận & Trả lời**: Hệ thống comment lồng nhau.
- **Bookmark**: Quản lý truyện yêu thích của người dùng.
- **Thông báo**: Hệ thống thông báo real-time.
- **Phân trang Nâng cao**: Tối ưu hóa truy vấn database cho các tập dữ liệu lớn.
- **Toàn vẹn Dữ liệu**: Được vận hành bởi **GORM** và **PostgreSQL**.

### ⚙️ Vận hành Hệ thống

- **Background Cleanup Job**: Goroutine chạy ngầm dọn dẹp tokens hết hạn mỗi 6 giờ.
- **Tài liệu Swagger**: API documentation tương tác.
- **Cấu hình từ Environment**: Toàn bộ config được đọc từ biến môi trường.
- **WebSocket Support**: Hỗ trợ tính năng real-time (tích hợp Centrifugo).
- **Upload Ảnh**: Tích hợp Cloudinary để lưu trữ ảnh.

---

## 🛠️ Công nghệ Sử dụng

| Loại               | Công nghệ                                       |
| ------------------ | ----------------------------------------------- |
| **Ngôn ngữ**       | [Go (Golang)](https://go.dev/) 1.22+            |
| **Web Framework**  | [Gin Gonic](https://gin-gonic.com/)             |
| **ORM**            | [GORM](https://gorm.io/)                        |
| **Database**       | [PostgreSQL](https://www.postgresql.org/)       |
| **Authentication** | [golang-jwt](https://github.com/golang-jwt/jwt) |
| **Image Storage**  | [Cloudinary](https://cloudinary.com/)           |
| **Realtime**       | [Centrifugo](https://centrifugal.dev/)          |
| **Documentation**  | [swaggo/swag](https://github.com/swaggo/swag)   |

---

## 🚀 Bắt đầu (Cài đặt)

### Yêu cầu hệ thống

- Go 1.22 trở lên
- PostgreSQL
- Tùy chọn: Tài khoản Cloudinary (để upload ảnh)
- Tùy chọn: Centrifugo (cho tính năng real-time)

### Các bước cài đặt

1. Clone dự án:

   ```bash
   git clone https://github.com/yourusername/nekozanedex-backend.git
   cd nekozanedex-backend
   ```

2. Tải các dependencies:

   ```bash
   go mod download
   ```

3. Cấu hình biến môi trường:

   ```bash
   cp .env.example .env
   # Chỉnh sửa file .env với thông tin của bạn
   ```

4. Chạy Server:
   ```bash
   go run cmd/server/main.go
   ```

API sẽ sẵn sàng tại [http://localhost:9091](http://localhost:9091).

---

## 📁 Cấu trúc Thư mục

```plaintext
nekozanedex-backend/
├── cmd/
│   └── server/           # Điểm bắt đầu ứng dụng (main.go)
├── docs/                 # Tài liệu Swagger (auto-generated)
├── internal/
│   ├── config/           # Load cấu hình từ environment
│   ├── database/         # Kết nối và khởi tạo Database
│   ├── handlers/         # Xử lý HTTP requests (Controllers)
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
│   │   ├── cors.go       # Cấu hình CORS
│   │   ├── csrf.go       # Bảo vệ CSRF
│   │   ├── logger.go     # Request logging
│   │   ├── rate_limit.go # Rate limiting
│   │   └── security.go   # Security headers
│   ├── models/           # Thực thể dữ liệu (GORM models)
│   ├── repositories/     # Lớp truy cập dữ liệu
│   ├── routes/           # Định nghĩa routes API
│   ├── services/         # Lớp xử lý nghiệp vụ
│   ├── utils/            # Công cụ hỗ trợ (JWT, Bcrypt, Result pattern)
│   └── websocket/        # WebSocket handlers (Centrifugo)
├── pkg/                  # Shared packages
├── .env.example          # Template biến môi trường
├── go.mod
└── go.sum
```

---

## ⚙️ Biến Môi trường

| Biến                        | Mô tả                                               | Mặc định                          |
| --------------------------- | --------------------------------------------------- | --------------------------------- |
| **App**                     |                                                     |                                   |
| `APP_ENV`                   | Môi trường (`development`, `staging`, `production`) | `development`                     |
| `PORT`                      | Port server                                         | `9091`                            |
| `GIN_MODE`                  | Gin mode (`debug`, `release`)                       | `debug`                           |
| **Database**                |                                                     |                                   |
| `DB_HOST`                   | PostgreSQL host                                     | `localhost`                       |
| `DB_PORT`                   | PostgreSQL port                                     | `5432`                            |
| `DB_USER`                   | Database user                                       | `postgres`                        |
| `DB_PASSWORD`               | Database password                                   | -                                 |
| `DB_NAME`                   | Tên database                                        | `nekozanedex`                     |
| **JWT**                     |                                                     |                                   |
| `JWT_ACCESS_SECRET`         | Secret key cho access token                         | -                                 |
| `JWT_REFRESH_SECRET`        | Secret key cho refresh token                        | -                                 |
| `JWT_ACCESS_EXPIRE_MINUTES` | Thời gian hết hạn access token (phút)               | `30`                              |
| `JWT_REFRESH_EXPIRE_DAYS`   | Thời gian hết hạn refresh token (ngày)              | `7`                               |
| **Cookie**                  |                                                     |                                   |
| `JWT_COOKIE_DOMAIN`         | Domain cho cookie                                   | -                                 |
| `JWT_COOKIE_SAME_SITE`      | SameSite policy                                     | `lax`                             |
| `JWT_COOKIE_MAX_AGE`        | Cookie max age (giây)                               | `604800`                          |
| **CORS**                    |                                                     |                                   |
| `CORS_DEV_ORIGINS`          | Allowed origins cho development                     | `http://localhost:3000,...`       |
| `CORS_PROD_ORIGINS`         | Allowed origins cho production                      | `https://nekozanedex.com,...`     |
| `CORS_STAGING_ORIGINS`      | Allowed origins cho staging                         | `https://staging.nekozanedex.com` |
| **Security**                |                                                     |                                   |
| `CSRF_SECRET_KEY`           | Secret key cho CSRF token                           | -                                 |
| `FRAME_ANCESTORS`           | CSP frame-ancestors directive                       | `'self'`                          |
| **Cloudinary**              |                                                     |                                   |
| `CLOUDINARY_CLOUD_NAME`     | Cloudinary cloud name                               | -                                 |
| `CLOUDINARY_API_KEY`        | Cloudinary API key                                  | -                                 |
| `CLOUDINARY_API_SECRET`     | Cloudinary API secret                               | -                                 |
| **Centrifugo**              |                                                     |                                   |
| `CENTRIFUGO_URL`            | URL server Centrifugo                               | `http://localhost:8000`           |
| `CENTRIFUGO_API_KEY`        | Centrifugo API key                                  | -                                 |

---

## 📖 Tài liệu API (Swagger)

Khi server đang chạy, truy cập Swagger UI tại:

```
http://localhost:9091/swagger/index.html
```

### Các Endpoint Chính

| Method | Endpoint                              | Mô tả                |
| ------ | ------------------------------------- | -------------------- |
| `POST` | `/api/auth/register`                  | Đăng ký tài khoản    |
| `POST` | `/api/auth/login`                     | Đăng nhập            |
| `POST` | `/api/auth/refresh`                   | Làm mới access token |
| `POST` | `/api/auth/logout`                    | Đăng xuất            |
| `GET`  | `/api/stories`                        | Danh sách truyện     |
| `GET`  | `/api/stories/:slug`                  | Chi tiết truyện      |
| `GET`  | `/api/stories/:slug/chapters`         | Danh sách chương     |
| `GET`  | `/api/stories/:slug/chapters/:number` | Nội dung chương      |
| `GET`  | `/api/genres`                         | Danh sách thể loại   |
| `GET`  | `/api/csrf-token`                     | Lấy CSRF token       |

---

## 🛡️ Cơ chế Bảo mật

### Refresh Token Rotation (RTR) với Reuse Detection

1. Khi client làm mới `access_token`, `refresh_token` cũng sẽ được đổi mới.
2. Nếu một `refresh_token` cũ bị sử dụng lại (dấu hiệu tấn công), hệ thống sẽ phát hiện ngay.
3. Khi phát hiện, **tất cả tokens** của người dùng đó sẽ bị vô hiệu hóa lập tức.

### Cấu hình CORS

CORS origins được cấu hình qua biến môi trường:

- **Development**: Cho phép localhost origins
- **Staging**: Bao gồm staging domain
- **Production**: Chỉ cho phép production domains

---

## 🤝 Đóng góp

Mọi đóng góp giúp hoàn thiện dự án đều được trân trọng! Vui lòng mở issue hoặc tạo PR.

## 📄 Bản quyền

Dự án được phát hành theo Giấy phép MIT.
