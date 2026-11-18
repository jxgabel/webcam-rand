# webcam-rand

`webcam-rand` is a learning project that uses visual data from one or more webcams as a source of real-world entropy. The system periodically captures frames, reduces them to 32-byte SHA-256-based samples, stores those samples in an API-side ring buffer, and uses the buffer to generate random integers through an HTMX-powered web interface.

The project is inspired by physical entropy sources such as Cloudflare's lava lamp wall. It demonstrates a small data pipeline across three separately runnable Go services:

```text
Webcam(s)
    │
    ▼
Harvester ──POST /hash──> API / ring buffer ──GET /random──> Web frontend
```

## Project layout

```text
webcam-rand/
├── harvester/
│   ├── main.go          # Capture loop, configuration loading, and API submission
│   ├── camera.go        # Opens V4L2 devices and captures frames
│   ├── entropy.go       # Hashes and combines source frame data
│   ├── config/
│   │   ├── config.go    # Configuration loading and validation
│   │   └── config.json  # Webcam sources, API URL, and capture interval
│   └── go.mod
├── api/
│   ├── main.go          # Gin HTTP server and /random, /status, and /hash routes
│   ├── models.go        # Request/response and ring-buffer types
│   ├── ringbuffer.go    # Thread-safe fixed-size byte buffer
│   └── go.mod
└── frontend/
    ├── main.go          # Gin server for the web page
    ├── templates/
    │   └── index.tmpl   # HTML and HTMX interface
    └── go.mod
```

Each directory is an independent Go module and service.

## How it works

1. The harvester loads `harvester/config/config.json`.
2. It opens each configured Linux V4L2 camera device at 320×240 JPEG format.
3. On the configured interval, it captures a frame from every available camera.
4. Each frame is passed through SHA-256, and the resulting 32-byte digests are XOR-combined into one 32-byte value.
5. The value is POSTed as a JSON hex string to `POST /hash` on the API.
6. The API validates that the submitted value is exactly 32 bytes and writes it into a 4096-byte ring buffer.
7. When `/random` is requested, the API consumes four bytes from the ring buffer and maps them to an integer in the requested inclusive range.
8. The frontend sends the request with HTMX and displays the result.

The ring buffer keeps only the newest data when it fills. Reading consumes four bytes at a time, so requesting random numbers also reduces the amount of available entropy in the pool.

## Services and ports

| Service | Port | Purpose |
|---|---:|---|
| Frontend | `8080` | Serves the HTML interface |
| API | `8888` | Stores entropy and provides random numbers/status |

By default, the services expect to run on the local machine and communicate over HTTP.

## API endpoints

### `POST /hash`

Adds one 32-byte SHA-256 value to the entropy pool.

```json
{
  "hash": "<64-character hexadecimal string>"
}
```

A successful request returns HTTP `201`.

### `GET /random`

Consumes 4 bytes and returns an integer in an inclusive range.

```text
/random?min=0&max=1000000
```

The response includes the number, hexadecimal bytes consumed, range, and remaining pool size.

### `GET /status`

Returns the number of available bytes, ring-buffer capacity, and percentage full.

## Configuration

The harvester is configured with `harvester/config/config.json`:

```json
{
  "app_name": "webcam-rand",
  "api": "http://localhost:8888",
  "sources": ["/dev/video0"],
  "capture_interval_milliseconds": 500
}
```

- `api` is the base URL of the API service.
- `sources` is a list of Linux V4L2 camera device paths.
- `capture_interval_milliseconds` controls how often frames are sampled.

## Requirements

- Go 1.27 or compatible
- Linux with V4L2 webcam support
- One or more `/dev/video*` camera devices
- Network access to the configured API and frontend ports

The camera integration uses `github.com/vladimirvivien/go4vl`, and the API and frontend use Gin with HTMX on the browser side.

## Running locally

Run each service from its own terminal. The API and frontend must be started before opening the web page.

Start the API:

```bash
cd api
go run .
```

Start the frontend:

```bash
cd frontend
go run .
```

Update `harvester/config/config.json` to identify your webcam devices, then start the harvester:

```bash
cd harvester
go run .
```

Open <http://localhost:8080> in a browser. Enter a minimum and maximum value, then use **Get a Random Number** to request a value from the API.

Stop any service with `Ctrl+C`.

## Notes

- This project is intended as an educational entropy-pipeline example, not a security-certified random-number generator.
- The API currently allows browser requests only from `http://localhost:8080` through its CORS middleware.
- The frontend calls `http://localhost:8888/random` directly, so the API and frontend URLs must remain aligned with the local configuration.
- Camera capture is Linux/V4L2-specific.
