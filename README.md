# Tracked — Backend

API REST construida en Go con PostgreSQL como base de datos y Cloudinary para el almacenamiento de imágenes. Expone endpoints para autenticación de usuarios y gestión de series.

**Frontend repo:** [Frontend](https://github.com/alemanuel18/-Proyecto1_Full-Stack_Web_Frontend.git)

---

## Stack

| Tecnología | Uso |
|---|---|
| Go 1.22 | Lenguaje del servidor |
| `net/http` | Servidor HTTP estándar (sin frameworks) |
| PostgreSQL 16 | Base de datos relacional |
| `lib/pq` | Driver de PostgreSQL para Go |
| JWT (golang-jwt/jwt v5) | Autenticación stateless |
| bcrypt | Hash de contraseñas |
| Cloudinary | Almacenamiento de imágenes |
| Docker + Docker Compose | Contenedores |

---

## Estructura del proyecto

```
backend/
├── main.go                  # Entry point
├── go.mod                   # Módulo y dependencias
├── Dockerfile               # Build multi-stage (builder + alpine final)
├── docker-compose.yml       # Orquesta API + PostgreSQL
├── .env.example             # Template de variables de entorno
├── config/config.go         # Carga variables de entorno
├── db/db.go                 # Conexión y migraciones automáticas
├── models/models.go         # Structs de datos
├── middleware/middleware.go  # CORS + autenticación JWT
├── handlers/
│   ├── auth.go              # Register, Login, Me
│   ├── series.go            # CRUD + paginación + filtros
│   ├── upload.go            # Subida de imágenes a Cloudinary
│   └── helpers.go           # respondJSON / respondError
└── routes/routes.go         # Registro de rutas
```

---

## Requisitos

- [Docker](https://docs.docker.com/get-docker/) y Docker Compose
- Cuenta gratuita en [cloudinary.com](https://cloudinary.com)

---

## Correr con Docker

```bash
# 1. Clonar el repositorio
git clone https://github.com/tu-usuario/seriestracker-backend.git
cd seriestracker-backend

# 2. Crear el archivo de entorno
cp .env.example .env
# Editar .env con tus credenciales

# 3. Levantar la API y PostgreSQL
docker compose up --build
```

La API queda disponible en `http://localhost:8080`.
Las migraciones se ejecutan automáticamente al iniciar.

```bash
docker compose up --build -d   # Correr en background
docker compose logs -f api     # Ver logs
docker compose down            # Detener
docker compose down -v         # Detener y borrar la base de datos
```

---

## Variables de entorno

| Variable | Descripción |
|---|---|
| `POSTGRES_USER` | Usuario de PostgreSQL |
| `POSTGRES_PASSWORD` | Contraseña de PostgreSQL |
| `POSTGRES_DB` | Nombre de la base de datos |
| `JWT_SECRET` | Clave secreta para firmar tokens. Generar con `openssl rand -hex 32` |
| `CLOUDINARY_CLOUD_NAME` | Cloud Name de tu cuenta en Cloudinary |
| `CLOUDINARY_API_KEY` | API Key de Cloudinary |
| `CLOUDINARY_API_SECRET` | API Secret de Cloudinary |
| `PORT` | Puerto del servidor (default: `8080`) |

---

## API Reference

Todos los endpoints protegidos requieren el header:
```
Authorization: Bearer <token>
```

### Auth

| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| `POST` | `/auth/register` | Crear cuenta | ❌ |
| `POST` | `/auth/login` | Iniciar sesión, devuelve JWT | ❌ |
| `GET` | `/auth/me` | Perfil del usuario autenticado | ✅ |

### Series

| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| `GET` | `/series` | Listar series del usuario | ✅ |
| `GET` | `/series/:id` | Obtener una serie | ✅ |
| `POST` | `/series` | Crear serie | ✅ |
| `PUT` | `/series/:id` | Editar serie | ✅ |
| `DELETE` | `/series/:id` | Eliminar serie | ✅ |
| `POST` | `/series/:id/image` | Subir portada (multipart, campo `image`, máx 1MB) | ✅ |

### Query params — GET /series

| Param | Descripción | Ejemplo |
|---|---|---|
| `page` | Número de página (default: 1) | `?page=2` |
| `limit` | Resultados por página (default: 20, máx: 100) | `?limit=10` |
| `q` | Buscar por título | `?q=breaking` |
| `sort` | Ordenar por: `title`, `rating`, `status`, `created_at`, `updated_at` | `?sort=rating` |
| `order` | Dirección: `asc` o `desc` | `?order=desc` |
| `status` | Filtrar por estado: `watching`, `completed`, `dropped`, `plan_to_watch` | `?status=watching` |

### Códigos de respuesta

| Código | Situación |
|---|---|
| `200` | Consulta o actualización exitosa |
| `201` | Recurso creado |
| `204` | Eliminación exitosa |
| `400` | Datos inválidos o faltantes |
| `401` | Token ausente, inválido o expirado |
| `404` | Recurso no encontrado |
| `409` | Email o username ya registrado |
| `500` | Error interno del servidor |

---

## CORS

CORS es la política del navegador que bloquea peticiones `fetch()` entre orígenes distintos (puertos diferentes en localhost cuentan como orígenes distintos). El servidor configura los headers necesarios para permitir las peticiones del frontend.

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

En producción reemplazar `*` por el dominio real del frontend.

---

## Challenges implementados

- [x] Autenticación con JWT
- [x] Códigos HTTP correctos en toda la API (201, 204, 400, 401, 404, 409, 500)
- [x] Validación server-side con errores descriptivos en JSON
- [x] Paginación con `?page=` y `?limit=`
- [x] Búsqueda por título con `?q=`
- [x] Ordenamiento con `?sort=` y `?order=`
- [x] Subida real de imágenes a Cloudinary (multipart/form-data, máx 1MB)

---

## Reflexión

_[¿Usarías Go de nuevo? ¿Qué aprendiste sobre separar cliente y servidor? ¿Qué harías diferente?]_

---

## Screenshot

_[Agregar screenshot de la app funcionando]_
