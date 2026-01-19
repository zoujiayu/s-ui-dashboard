# S-UI Dashboard

A lightweight web dashboard for **[S-UI](https://github.com/alireza0/s-ui)**, providing real-time traffic statistics, user usage overview, and system runtime information.

---

## Features

- User-based traffic statistics (Upload / Download)
- Remaining & Total traffic display (Unlimited supported)
- Real-time traffic chart (Chart.js)
- System runtime & online status
- Clean, compact, responsive UI
- No database required
- Supports direct run and Docker deployment

---

## Screenshots
![Screenshots-01](https://raw.githubusercontent.com/zoujiayu/s-ui-dashboard/master/screenshot/01.png)

## Requirements

- Go **1.20+**
- Running **S-UI** with API enabled
- Valid API Token

---

## Required Configuration (Before Running)

Before running the program, **you must edit `main.go`** and set the correct API endpoint and token.

```go
const (
    BaseURL = "http://127.0.0.1:2095/app/apiv2"
    Token   = "xxxxxxxxxxxxxxx"
)
```

### Configuration Description

| Field | Description |
|------|-------------|
| BaseURL | S-UI API address (usually port `2095`) |
| Token | API authentication token |

> The dashboard service itself listens on port **2097**.

---

## Run Locally (Normal Mode)

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Build

```bash
go build -o s-ui-dashboard
```

### 3. Run

```bash
./s-ui-dashboard
```

The service will listen on:

```text
0.0.0.0:2097
```

---

## Access URL

```text
http://127.0.0.1:2097/?user=USERNAME
```

Example:

```text
http://127.0.0.1:2097/?user=admin
```

Notes:

- `user` parameter is required
- If the user does not exist, the server returns **404**
- Traffic and runtime data are loaded dynamically

---

## Docker Compose Build & Run

From the **project root directory**, execute:

```bash
docker-compose build
docker-compose up -d
```

---

### Docker + Host API Notes

If S-UI API is running on the host machine, modify `BaseURL` in `main.go`:

```go
BaseURL = "http://host.docker.internal:2095/app/apiv2"
```

Or replace `host.docker.internal` with your host IP address.

---

## Access (Docker Mode)

```text
http://127.0.0.1:2097/?user=USERNAME
```

Example:

```text
http://127.0.0.1:2097/?user=admin
```

---

## Nginx Reverse Proxy (Optional)

```nginx
location / {
    proxy_pass http://127.0.0.1:2097;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

Access example:

```text
https://panel.example.com/?user=admin
```

---

## Security Notes

- This dashboard **does not provide authentication**
- Strongly recommended protections:
    - Nginx Basic Auth
    - IP whitelist
    - Firewall rules
- Avoid exposing directly to the public internet
- Users with `volume = 0` are automatically displayed as **Unlimited**
- It is recommended to block or hide sensitive users (e.g. `admin`) at proxy level

---

## Access Path Summary

```text
http://127.0.0.1:2097/?user=USERNAME
```

---

## License

MIT License

---

## Disclaimer

This project is intended for **personal or internal use only**.  
The author is not responsible for any misuse or data exposure caused by improper deployment.
