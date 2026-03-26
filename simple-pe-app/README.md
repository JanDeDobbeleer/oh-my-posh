# simple-pe-app

Aplicación fintech peruana para gestión de cuentas, transferencias y pagos en soles (PEN) y dólares (USD).

## Stack

| Capa | Tecnologías |
|------|-------------|
| Frontend | React 18, TypeScript, Vite, React Router v6, Axios |
| Backend | Node.js, Express, TypeScript, Zod |
| Base de datos | PostgreSQL 15 |
| Autenticación | JWT + Refresh Tokens, bcrypt |
| DevOps | Docker, Docker Compose |

## Levantar el proyecto

```bash
# 1. Base de datos
docker compose up -d postgres

# 2. Variables de entorno
cp backend/.env.example backend/.env

# 3. Instalar dependencias
npm run install:all

# 4. Correr
npm run dev
```

- Frontend: http://localhost:5173
- Backend: http://localhost:4000/api/v1

## Usuarios de prueba

| Email | Password | Rol |
|-------|----------|-----|
| admin@simple-pe.com | Admin1234! | admin |
| ana.flores@gmail.com | User1234! | user |
| luis.quispe@hotmail.com | User1234! | user |
