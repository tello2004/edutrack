# EduTrack

Este repositorio contiene el código fuente de EduTrack, un sistema Web
minimalista, para gestión de alumnos y evaluaciones. Utiliza PostgreSQL
como base de datos por defecto, con soporte opcional para SQLite.

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
- **PostgreSQL 14+** como base de datos (por defecto)
- **SQLite 3** (opcional, para desarrollo local)

### Junto usando Compose

Para correr los dos servidores de desarrollo (recarga ante cambios),
puede usar Docker/Podman Compose:

```bash
docker compose up --build
```

Esto levantará los servicios necesarios, incluyendo PostgreSQL.

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

# Conexión a PostgreSQL (por defecto)
export DATABASE_URL="host=localhost user=edutrack password=edutrack dbname=edutrack port=5432 sslmode=disable"
```

#### 3. Compilar y ejecutar

**Con PostgreSQL (por defecto)**

```bash
# Compilar (PostgreSQL es el driver por defecto)
go build -o edutrackd ./cmd/edutrackd

# Ejecutar el servidor
./edutrackd
```

**Con SQLite (opcional, para desarrollo local)**

```bash
# Configurar ruta de la base de datos SQLite
export DATABASE_URL="edutrack.db"

# Compilar con soporte SQLite
go build -tags sqlite -o edutrackd ./cmd/edutrackd

# Ejecutar el servidor
./edutrackd
```

El backend estará disponible en `http://localhost:8080`

#### Bases de datos soportadas

| Base de datos | Tag de compilación | Variable de entorno | Por defecto |
|---------------|-------------------|---------------------|-------------|
| PostgreSQL | (ninguno) | `DATABASE_URL` (connection string) | ✓ |
| SQLite | `-tags sqlite` | `DATABASE_URL` (ruta al archivo .db) | |

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
| GET | `/subjects/{id}/students` | Listar estudiantes inscritos en una materia |
| POST | `/subjects/{id}/students` | Agregar un estudiante a una materia |
| DELETE | `/subjects/{id}/students/{student_id}` | Remover un estudiante de una materia |
| GET/POST | `/topics` | Listar/Crear temas |
| GET/PUT/DELETE | `/topics/{id}` | Obtener/Actualizar/Eliminar tema |
| GET/POST | `/attendances` | Listar/Crear asistencias |
| GET/PUT/DELETE | `/attendances/{id}` | Obtener/Actualizar/Eliminar asistencia |
| GET/POST | `/grades` | Listar/Crear calificaciones |
| GET/PUT/DELETE | `/grades/{id}` | Obtener/Actualizar/Eliminar calificación |

Detalles de cada endpoint (parámetros / cuerpo)
> Nota: los siguientes esquemas de request están inferidos a partir de los modelos en `edutrack/api` y las rutas registradas en `api/http/server.go`. Para detalles exactos de validación/respuestas revise los handlers correspondientes.

- Autenticación
  - `POST /auth/login`
    - Auth: pública
    - Body (JSON):
      - `email` (string, requerido)
      - `password` (string, requerido)
      - opcional: `tenant_id` (string) — si la instalación es multi-tenant
    - Respuesta esperada: token JWT (cabecera `Authorization: Bearer <token>`) y datos del usuario.
  - `POST /auth/license`
    - Auth: pública
    - Body (JSON):
      - `license_key` (string, requerido)
      - `tenant_id` (string, requerido)
    - Uso: validar licencia institucional / activar tenant.

- Cuentas (`accounts`)
  - `GET /accounts`
    - Auth: requerida (Bearer token)
    - Query params comunes: `page`, `per_page`, `tenant_id`, `role`
    - Lista cuentas (filtrable por `tenant_id`).
  - `POST /accounts`
    - Auth: requerida (normalmente solo `secretary`)
    - Body (JSON):
      - `name` (string, requerido)
      - `email` (string, requerido, único por tenant)
      - `password` (string, requerido) — se almacenará hasheada
      - `role` (string: `secretary`|`teacher`|`student`, opcional)
      - `active` (bool, opcional)
      - `tenant_id` (string, requerido)
  - `GET /accounts/{id}`
    - Auth: requerida
    - Path params:
      - `id` (account id)
  - `PUT /accounts/{id}`
    - Auth: requerida
    - Body (JSON): campos actualizables (ej. `name`, `email`, `password`, `role`, `active`)
  - `DELETE /accounts/{id}`
    - Auth: requerida (normalmente solo `secretary`)
    - Elimina o desactiva la cuenta.

- Estudiantes (`students`)
  - `GET /students`
    - Auth: requerida
    - Query params: `page`, `per_page`, `tenant_id`, `career_id`, `semester`
  - `POST /students`
    - Auth: requerida
    - Body (JSON):
      - `student_id` (string, requerido, único por tenant)
      - `semester` (int, opcional, default 1)
      - `tenant_id` (string, requerido)
      - `account_id` (uint, opcional) — vincular a `Account`
      - `career_id` (uint, opcional)
      - `subjects` (array de uint opcional) — IDs de materias
  - `GET /students/{id}`
    - Auth: requerida
    - Path: `id` (student record id)
  - `PUT /students/{id}`
    - Auth: requerida
    - Body (JSON): campos actualizables similares a `POST /students`
  - `DELETE /students/{id}`
    - Auth: requerida

- Docentes (`teachers`)
  - `GET /teachers`
    - Auth: requerida
    - Query params: `page`, `per_page`, `tenant_id`
  - `POST /teachers`
    - Auth: requerida
    - Body (JSON):
      - `tenant_id` (string, requerido)
      - `account_id` (uint, requerido)
      - `subjects` (array de uint, opcional) — materias asignadas
  - `GET /teachers/{id}`, `PUT /teachers/{id}`, `DELETE /teachers/{id}` similares a estudiantes.

- Carreras (`careers`)
  - `GET /careers`
    - Auth: requerida
    - Query params: `tenant_id`, `active`
  - `POST /careers`
    - Auth: requerida
    - Body (JSON):
      - `name` (string, requerido)
      - `code` (string, requerido, único por tenant)
      - `description` (string, opcional)
      - `duration` (int, opcional)
      - `active` (bool, opcional)
      - `tenant_id` (string, requerido)
  - `GET /careers/{id}`, `PUT /careers/{id}`, `DELETE /careers/{id}`: path param `id`, body de `PUT` con campos actualizables.

- Materias (`subjects`)
  - `GET /subjects`
    - Auth: requerida
    - Query params: `tenant_id`, `career_id`, `semester`, `teacher_id`
  - `POST /subjects`
    - Auth: requerida
    - Body (JSON):
      - `name` (string, requerido)
      - `code` (string, requerido)
      - `description` (string, opcional)
      - `credits` (int, opcional)
      - `semester` (int, opcional)
      - `career_id` (uint, requerido)
      - `teacher_id` (uint, opcional)
      - `tenant_id` (string, requerido)
  - `GET /subjects/{id}`, `PUT /subjects/{id}`, `DELETE /subjects/{id}`: path param `id`, `PUT` acepta campos anteriores.
  - `GET /subjects/{id}/students`
    - Auth: requerida
    - Path param: `id` (subject id)
    - Lista los estudiantes inscritos en la materia.
  - `POST /subjects/{id}/students`
    - Auth: requerida
    - Path param: `id` (subject id)
    - Body (JSON):
      - `student_id` (uint, requerido) — ID del estudiante a agregar
    - Añade el estudiante a la materia (tabla many-to-many `student_subjects`).
  - `DELETE /subjects/{id}/students/{student_id}`
    - Auth: requerida
    - Path params: `id` (subject id), `student_id` (student id)
    - Remueve la relación estudiante–materia.

- Temas (`topics`)
  - `GET /topics`
    - Auth: requerida
    - Query params: `subject_id`, `tenant_id`
  - `POST /topics`
    - Auth: requerida
    - Body (JSON):
      - `name` (string, requerido)
      - `description` (string, opcional)
      - `subject_id` (uint, requerido)
      - `tenant_id` (string, requerido)
  - `GET /topics/{id}`, `PUT /topics/{id}`, `DELETE /topics/{id}`: path param `id`, `PUT` con campos actualizables.

- Asistencias (`attendances`)
  - `GET /attendances`
    - Auth: requerida
    - Query params: `date`, `student_id`, `subject_id`, `tenant_id`
  - `POST /attendances`
    - Auth: requerida
    - Body (JSON):
      - `date` (string ISO-8601, requerido)
      - `status` (string: `present`|`absent`|`late`|`excused`, requerido)
      - `notes` (string, opcional)
      - `student_id` (uint, requerido)
      - `subject_id` (uint, requerido)
      - `tenant_id` (string, requerido)
  - `GET /attendances/{id}`, `PUT /attendances/{id}`, `DELETE /attendances/{id}`: operan sobre `id`.

- Calificaciones (`grades`)
  - `GET /grades`
    - Auth: requerida
    - Query params: `student_id`, `topic_id`, `subject_id`, `tenant_id`, `page`, `per_page`
  - `POST /grades`
    - Auth: requerida
    - Body (JSON):
      - `value` (float, requerido)
      - `notes` (string, opcional)
      - `student_id` (uint, requerido)
      - `topic_id` (uint, requerido)
      - `tenant_id` (string, requerido)
  - `GET /grades/{id}`, `PUT /grades/{id}`, `DELETE /grades/{id}`: `PUT` permite actualizar `value`, `notes`.

Encabezados y autenticación
- Para rutas protegidas incluir cabecera:
  - `Authorization: Bearer <JWT>`
- Contenido JSON: usar `Content-Type: application/json`.

Paginación y filtros
- Las rutas de listado suelen aceptar:
  - `page` (int)
  - `per_page` (int)
  - filtros específicos de recurso como `tenant_id`, `career_id`, `subject_id`, etc.

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

### CLI de Administración

EduTrack incluye una herramienta CLI para administrar tenants, licencias y cuentas:

```bash
# Compilar el CLI
go build -o edutrack ./cmd/edutrack

# Ver comandos disponibles
./edutrack --help

# Ejemplos
./edutrack tenant add "Mi Institución"
./edutrack tenant list
./edutrack account add -tenant=abc123 -email=admin@example.com -name="Admin" -password=secret -role=secretary
```

El CLI también usa PostgreSQL por defecto. Configure `DATABASE_URL` para conectarse a su base de datos.
