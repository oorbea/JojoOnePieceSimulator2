---
name: local-up
description: >
  Levanta el stack completo del proyecto en local con Docker Compose
  (postgres, redis, backend, frontend). Comprueba que Docker Desktop esté
  encendido (lo arranca si puede, si no pide al usuario que lo encienda),
  crea la red externa `public-net` si falta, libera los puertos 8080/3000
  si están ocupados (matando contenedores huérfanos del propio proyecto
  sin preguntar; pidiendo confirmación si el proceso ocupante es ajeno),
  y ejecuta `docker compose up -d --build` con los ficheros
  deployments/docker-compose.yml + docker-compose.dev.yml.
  Trigger: "levanta el proyecto en local", "arranca todo con docker",
  "docker compose up", "run project locally", "start local stack".
---

Stack: postgres + redis + backend + frontend, definido en
`deployments/docker-compose.yml` (base) + `deployments/docker-compose.dev.yml`
(override dev: publica puertos `${PORT:-8080}` backend y
`${FRONTEND_PORT:-3000}` frontend). Red externa requerida: `public-net`.
`.env` ya vive en `deployments/.env` (copiado de `.env.example`) — si falta,
avisar al usuario y parar, no inventar secretos.

Ejecuta estos pasos en orden, en PowerShell (shell primario del proyecto).
Usa Bash si el usuario está en WSL/Linux, mismos comandos con sintaxis POSIX.

## 1. Docker Desktop encendido

```powershell
docker info 2>$null | Out-Null
$dockerUp = $?
```

Si `$dockerUp` es false:
- Intenta arrancar Docker Desktop automáticamente:
  ```powershell
  Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe"
  ```
- Espera a que el daemon responda, con timeout (~90s), sondeando cada 3s:
  ```powershell
  $deadline = (Get-Date).AddSeconds(90)
  do {
    Start-Sleep -Seconds 3
    docker info 2>$null | Out-Null
  } until ($? -or (Get-Date) -gt $deadline)
  ```
- Si sigue sin responder al timeout (o el .exe no existe en esa ruta — probar
  también buscar el proceso `Docker Desktop` o preguntar la ruta), **para y
  pide al usuario que encienda Docker Desktop manualmente**. No sigas sin
  daemon activo.

## 2. Red externa `public-net`

```powershell
docker network inspect public-net 2>$null | Out-Null
if (-not $?) { docker network create public-net }
```

## 3. Puertos libres (8080 backend, 3000 frontend — leer valores reales de
   `deployments/.env` si `PORT`/`FRONTEND_PORT` están seteados, si no usar
   default 8080/3000)

Para cada puerto:
```powershell
$conn = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue
if ($conn) {
  $proc = Get-Process -Id $conn.OwningProcess
  # ¿es un contenedor huérfano de este proyecto? (com.docker.compose label / nombre jojo-one-piece-simulator-*)
  $ownContainer = docker ps --filter "publish=$port" --format "{{.Names}}" | Select-String "jojo-one-piece-simulator"
}
```

- Si el puerto lo ocupa un contenedor Docker con nombre
  `jojo-one-piece-simulator-*` (contenedor huérfano de una ejecución previa
  de este mismo proyecto): pararlo/eliminarlo sin preguntar
  (`docker stop <name>`, opcionalmente `docker rm`) — es limpieza rutinaria
  del propio stack.
- Si lo ocupa cualquier otro proceso o contenedor ajeno al proyecto:
  **no lo mates sin confirmar**. Muestra qué proceso/contenedor es
  (nombre, PID) y pregunta al usuario si quiere que lo cierre, o si prefiere
  cambiar el puerto en `deployments/.env` (`PORT=`/`FRONTEND_PORT=`).

## 4. Levantar el stack

Desde la raíz del repo:
```powershell
docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.dev.yml up -d --build
```
(mismo comando que usa `apps/backend/Makefile` vía `$(COMPOSE)`, añadiendo
`--build` para recoger cambios de código locales).

## 5. Verificación

- `docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.dev.yml ps`
  — confirmar que los 4 servicios están `running`/`healthy`.
- Si algún servicio no arranca sano, mostrar sus logs:
  `docker compose ... logs <service> --tail=50`.
- Reportar al usuario las URLs:
  - Frontend: http://localhost:<FRONTEND_PORT o 3000>
  - Backend health: http://localhost:<PORT o 8080>/health

## Notas

- No matar procesos/contenedores ajenos sin confirmación explícita (regla de
  seguridad general del proyecto — acciones destructivas/irreversibles).
- No tocar `deployments/.env` (secretos) salvo que el usuario pida
  explícitamente cambiar un puerto.
- Si `deployments/.env` no existe, parar y decir al usuario que lo cree desde
  `.env.example` (no hay entorno de dev sin él, ver comentario en ese fichero).
