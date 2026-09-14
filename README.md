<p align="center">
  <img src="assets/logo.jpg" alt="Mimisbrunnr — el pozo de Mímir" width="280">
</p>

# Mimisbrunnr

Servidor MCP en Go para gestionar memoria persistente de agentes.

## Requisitos

- Go 1.26+

## Layout

```
cmd/mimisbrunnr/        binario stdio
internal/config/        variables de entorno
internal/mcp/           servidor MCP (SDK oficial)
internal/storage/       puerto Store + factory
internal/storage/sqlite/  adaptador SQLite (único implementado)
```

## Uso

```bash
go run ./cmd/mimisbrunnr
```

El proceso habla MCP por stdin/stdout. Un cliente (Cursor, Claude, VS Code) lo lanza como subprocess.

Tools MCP con prefijo `mimir_` (el agente actúa en el pozo de Mímir). Implementados: `mimir_ping`, `mimir_save`, `mimir_get`, `mimir_list`, `mimir_search`, `mimir_session_start`, `mimir_session_end`, `mimir_context`.

### Sesiones y contexto

1. `mimir_session_start` → obtienes `session_id`
2. `mimir_save` con `session_id` opcional para vincular observaciones
3. `mimir_context` devuelve título + snippet (sin `content` completo) y `context_text` legible
4. `mimir_session_end` cierra la sesión y guarda un resumen (`type=session_summary`, `topic_key=sessions/{id}/summary`)

## Cursor (este repo)

Config en [`.cursor/mcp.json`](.cursor/mcp.json). **Rutas con espacios** (`proyectos personales`) rompen el spawn de Cursor; usa el launcher sin espacios:

```bash
go build -o bin/mimisbrunnr ./cmd/mimisbrunnr
chmod +x scripts/mimisbrunnr-mcp.sh
ln -sf "$(pwd)/scripts/mimisbrunnr-mcp.sh" ~/.local/bin/mimisbrunnr-mcp
```

El launcher apunta al repo y deja la DB en `.mimisbrunnr.db` (gitignored). Luego **Settings → MCP** → activa `mimisbrunnr` → recarga → `mimir_ping`.

## Configuración

| Variable | Default | Descripción |
|----------|---------|-------------|
| `MIMISBRUNNR_DRIVER` | `sqlite` | Driver de almacenamiento. Solo `sqlite` está implementado. |
| `MIMISBRUNNR_SQLITE_PATH` | `~/.config/mimisbrunnr/mimisbrunnr.db` | Ruta del archivo SQLite (o `mimisbrunnr.db` en el cwd si no hay config dir). |

Otros drivers (p. ej. `postgres`) devuelven error al arrancar hasta que exista un adaptador.

## Tests

```bash
go test ./...
```
