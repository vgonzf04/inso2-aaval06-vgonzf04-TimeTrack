# TimeTrack App

Una solución full-stack para gestión de empleados, fichajes (timecards), solicitudes de vacaciones y métricas de productividad.

- **Backend**: Go + Gin + GORM + PostgreSQL  
- **Frontend**: Next.js + React + TailwindCSS  
- **Contenerización**: Docker & Docker Compose  
- **Producción**: Backend en Render · Frontend en Vercel

  ## 🛠 Prerrequisitos

- Docker ≥ 20.x  
- Docker Compose ≥ 1.29.x  
- Go ≥ 1.20 (si compilas el backend local sin Docker)  
- Node ≥ 16.x (si levantas el frontend local sin Docker) 

---

## 📁 Estructura del repositorio
├── backend/ # Código Go (API REST)
│ ├── controllers/
│ ├── models/
│ ├── main.go
│ ├── Dockerfile
│ └── init.sql # Creación de tablas iniciales
├── frontend/ # Next.js (React + Tailwind)
│ ├── app/
│ ├── components/
│ ├── public/
│ ├── package.json
│ └── Dockerfile
└── docker-compose.yml # Orquestación local

## 🚀 Desarrollo local con Docker Compose

1. Crea un fichero `.env` en la raíz con:
   env
   # PostgreSQL
   DB_HOST=db
   DB_PORT=5432
   DB_USER=admin
   DB_PASSWORD=admin123
   DB_NAME=empleadosdb
   DB_SSLMODE=disable
   DB_TIMEZONE=Europe/Madrid

   # JWT
   JWT_SECRET=tu_secreto_jwt

   # Google OAuth & Maps (opcional)
   GOOGLE_CLIENT_ID=…
   GOOGLE_CLIENT_SECRET=…
   GOOGLE_MAPS_API_KEY=…

  3. Inicia los servicios:

      docker-compose up --build
      db en localhost:5432

      backend en localhost:3000

     frontend en localhost:3001

   3. Para cerrar:
      docker-compose down

      
   5. USO

      Abre el navegador en http://localhost:3001

      Pulsa Login con Google para autenticarte.

      En el Dashboard:

      Botón Comenzar día / Parar día para fichajes.

      Vista de Timecards y Solicitudes de vacaciones.

      En la barra lateral:

      Pedir vacaciones con el calendario.

      Rol supervisor: en Admin podrás aprobar/rechazar vacaciones y gestionar empleados.



    
