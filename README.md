# hello-go

A tiny Go web app for trying out [CommyPanel](https://github.com/emrecanful/commypanel). Standard library only; no dependencies.

| Path | What it returns |
|---|---|
| `/` | a small HTML page (greeting, version, container name, uptime) |
| `/about` | a second page (added in 1.1.0 to try a redeploy) |
| `/info` | the same facts as JSON |
| `/health` | `ok` |

Run it locally:

```bash
go run .            # http://localhost:8080
PORT=3000 go run .  # another port
```

## Deploy it with CommyPanel

1. Push this repository to GitHub (public, or private plus a deploy key from the app page).
2. In CommyPanel: **Apps → New app**
   - Git repository URL: your repo's `https://...` URL
   - Branch: `main`
   - App port: `8080`
   - Health check path: `/health`
3. Press **Deploy**. CommyPanel detects Go from `go.mod`, builds an image and starts it.

## Things worth trying

- **Redeploy:** change `version` in `main.go`, commit, push, press Deploy. The new version goes live and the page shows a new container name.
- **Environment variables:** add `GREETING` = `Hello, world!` on the app page, then deploy. The page headline changes.
- **A failing deploy:** add `FAIL_STARTUP` = `1`, then deploy. The new version never answers, the deploy fails, and the previous version keeps serving. Remove the variable and deploy again.
- **Rollback:** after two successful deploys, use "Roll back to this" on the older release.
- **A domain:** add a domain in CommyPanel, point its A record at your server, assign it to this app, and turn on automatic HTTPS.
