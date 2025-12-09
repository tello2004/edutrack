# Guía de contribución

Para mantener la calidad académica del código, seguimos estrictamente las
reglas mencionadas en este archivo.

## Stack

- La aplicación está divida en dos partes: una API a través de HTTP escrita en
Go y un cliente Web escrito en TypeScript, usando React.
- Para la base de datos utilizamos `dbmate` para mantener las migraciones de la
base de datos (en este caso PostgreSQL.)

Para construir la aplicación completa (tanto backend como frontend), son
necesarias las siguientes herramientas:

- [Node.js](https://nodejs.org) (al menos v22)
- [Go](https://go.dev) (al menos v1.25)
- [PostgreSQL](https://www.postgresql.org) (al menos v16)
- [dbmate](https://github.com/amacneil/dbmate)
- [Terraform](https://developer.hashicorp.com/terraform)
- [Docker](https://www.docker.com) o [Podman](https://podman.io)

## Ramas de desarrollo

- El flujo de trabajo se maneja a través de ramas de desarrollo o
"features branches" que engloban una única característica nueva
(que puede modificar colateralmente múltiples módulos relacionados).
- La rama `main` siempre estará en estado de desarrollo. Para versiones
estables se encuentran etiquetadas siguiendo un
[versionado semántico](https://semver.org/lang/es/).

Para crear una rama de desarrollo prueba con:

```bash
git checkout -b [NOMBRE]

# por ejemplo:
git checkout -b student-pk
git checkout -b typescript-upgrade
```

## Commits

Los nombres de commits deben llevar un prefijo que haga referencia a qué
parte del proyecto se está contribuyendo. Además, el mensaje debe estar
escrito en infinitivo. Ejemplos correctos son:

- `app: Ajustar posición del botón de Inicio`
- `api: Validar cuerpo de estudiante antes de ingresar a BD`

## Guía de estílo

> [!IMPORTANT]
> La opción de formato al guardar en el editor de código debe estar activada.

### Go

- Seguimos la guía de estilo recomendado por el equipo. Si utiliza
Visual Studio Code, es necesario tener la
[extensión para Go](https://marketplace.visualstudio.com/items?itemName=golang.go)
instalada. Para formatear el código desde la terminal puede utilizar `go fmt`.

### JavaScript

- El proyecto del cliente utiliza TypeScript en lugar de JavaScript.
- Seguimos la guía de estilo recomendado por [Prettier](https://prettier.io).
Si utiliza Visual Studio Code, es necesario tener la
[extensión de Prettier](https://marketplace.visualstudio.com/items?itemName=esbenp.prettier-vscode)
instalada.

## Otro tipo de contribución

Para temas relacionados con ciberseguridad y de otra índole, contacte al [líder del proyecto](mailto:hu220111144@lahuerta.tecmm.edu.mx).
