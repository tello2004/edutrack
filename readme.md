# EduTrack

Este repositorio contiene el código fuente de EduTrack, un sistema Web
minimalista, para gestión de alumnos y evaluaciones.

Para más información como el stack tecnológico o guías de estílo, revise
[la guía de contribución](./contributing.md).

## Características

EduTrack es ideal para instituciones pequeñas o talleres de capacitación
donde los instructores no tienen un sistema central para registrar
asistencias, calificaciones y reportes.

- Gestión de alumnos.
- Registro de asistencias en tiempo real.
- Registro de evaluaciones.
- Dashboard con rendimiento por estudiante y por grupo.
- Exportación de reportes en CSV/PDF.

## Instalación

### Requisitos

- **Go 1.25+** para el backend
- **Node.js 20+** y npm para el frontend

### Junto usando Compose

Para correr los dos servidores de desarrollo (recarga ante cambios),
puede usar Docker/Podman Compose:

```bash
docker compose up --build
```

Esto levantará los servicios necesarios.

### Configuración del Backend (API)

#### 1. Navegar al directorio del backend

```bash
cd api
```

#### 2. Configurar variables de entorno

Crear un archivo `.env` o exportar las variables:

```bash
# Puerto del servidor (por defecto :8080)
export EDUTRACK_ADDR=":8080"

# Secreto JWT (CAMBIAR en producción)
export EDUTRACK_JWT_SECRET="tu-secreto-seguro-aqui"
```

#### 3. Compilar y ejecutar

```bash
# Compilar con soporte SQLite
go build -tags sqlite -o edutrackd ./cmd/edutrackd

# Ejecutar el servidor
./edutrackd
```

El backend estará disponible en `http://localhost:8080`

#### Endpoints de la API

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/auth/login` | Iniciar sesión con email/contraseña |
| POST | `/auth/license` | Validar licencia institucional |
| GET/POST | `/accounts` | Listar/Crear cuentas |
| GET/PUT/DELETE | `/accounts/{id}` | Obtener/Actualizar/Eliminar cuenta |
| GET/POST | `/students` | Listar/Crear estudiantes |
| GET/PUT/DELETE | `/students/{id}` | Obtener/Actualizar/Eliminar estudiante |
| GET/POST | `/teachers` | Listar/Crear docentes |
| GET/PUT/DELETE | `/teachers/{id}` | Obtener/Actualizar/Eliminar docente |
| GET/POST | `/careers` | Listar/Crear carreras |
| GET/PUT/DELETE | `/careers/{id}` | Obtener/Actualizar/Eliminar carrera |
| GET/POST | `/subjects` | Listar/Crear materias |
| GET/PUT/DELETE | `/subjects/{id}` | Obtener/Actualizar/Eliminar materia |
| GET/POST | `/attendances` | Listar/Crear asistencias |
| GET/PUT/DELETE | `/attendances/{id}` | Obtener/Actualizar/Eliminar asistencia |
| GET/POST | `/grades` | Listar/Crear calificaciones |
| GET/PUT/DELETE | `/grades/{id}` | Obtener/Actualizar/Eliminar calificación |

### Configuración del Frontend (Client)

#### 1. Navegar al directorio del frontend

```bash
cd client
```

#### 2. Instalar dependencias

```bash
npm install
```

#### 3. Ejecutar en modo desarrollo

```bash
npm run dev
```

El frontend estará disponible en `http://localhost:5173`

#### 4. Build para producción

```bash
npm run build
```
