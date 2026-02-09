# fauxfile

In-memory download/upload server for benchmarking. No real files are read or written; downloads are generated on the fly and uploads are hashed in memory then discarded.

## Build

```bash
go build -o fauxfile ./cmd/fauxfile
```

## Run

```bash
./fauxfile
```

Listens on `:8080` by default. Override with `-addr` or the `FAUXFILE_ADDR` environment variable (flag overrides env).

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | Listen address |
| `-max-size` | (none) | Max download size, e.g. `1g`, `100m`; empty = no limit |
| `-max-upload` | (none) | Max upload body size, e.g. `100m`; empty = no limit |
| `-hash` | `sha256` | Default hash algorithm: `sha256`, `sha512`, `sha1`, `md5` |
| `-response-type` | `text` | Default upload response body: `text` or `json` |

## Download (GET)

Request a stream of random bytes by size. Size can be in the path or in the query; **path wins** if both are present.

- **Path:** `/download/<size>.bin` — e.g. `/download/10mb.bin`, `/download/10m.bin` (case insensitive).
- **Query:** `/download?size=<size>` — e.g. `?size=10mb`, `?size=1024` (bytes if no unit).

**Size format:** Integer or decimal, optional unit: `b`, `k`/`kb`, `m`/`mb`, `g`/`gb` (case insensitive). Examples: `1024`, `10mb`, `1.5g`.

**Hash:** The server streams random data and sends the content hash in **trailers**: `X-Content-Hash` and `X-Hash-Algorithm`. Read the body to completion to receive trailers.

**Hash algorithm:** Query `?hash=sha512` or header `X-Hash-Algorithm: sha512`; query overrides header; both override the default from `-hash`.

Examples:

```bash
# 10 MiB from path
curl -o /dev/null -D - "http://localhost:8080/download/10mb.bin"

# 1 KiB from query
curl "http://localhost:8080/download?size=1k" | wc -c

# SHA-512 hash (trailers appear after body is consumed)
curl -o /dev/null -D - "http://localhost:8080/download/1k.bin?hash=sha512"
```

## Upload (POST)

`POST /upload` with the body as raw bytes. The server hashes the body in memory and does not write to disk. Response includes the hash in headers and in the body.

**Headers:** `X-Content-Hash`, `X-Hash-Algorithm`.

**Body:** Plain hash (default) or JSON. Use `?type=text` for plain hash or `?type=json` for e.g. `{"hash":"...","algorithm":"sha256"}`. Default is controlled by `-response-type`.

**Hash algorithm:** Same as download (query `hash=`, header `X-Hash-Algorithm`, default from `-hash`).

**Limit:** If `-max-upload` is set, the server reads only up to that many bytes from the body.

Examples:

```bash
# Upload and get hash in header + plain body
echo -n "hello" | curl -X POST -d @- "http://localhost:8080/upload"

# JSON response
echo -n "hello" | curl -X POST -d @- "http://localhost:8080/upload?type=json"

# SHA-512
echo -n "hello" | curl -X POST -H "X-Hash-Algorithm: sha512" -d @- "http://localhost:8080/upload"
```

## Benchmarking tips

- Use path-based sizes (e.g. `/download/100mb.bin`) for repeatable URLs.
- Enforce `-max-size` and `-max-upload` in shared environments to avoid abuse.
- For download throughput tests, consume the body fully so the server can send trailers; the hash is only available in trailers.
- Run the server and client on different machines to measure network throughput without disk I/O.

## License

See repository.
