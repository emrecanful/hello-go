// A tiny web app for trying out CommyPanel.
//
// It listens on $PORT (default 8080) and answers:
//
//	/        a small HTML page
//	/about   a second page, added in 1.1.0 to try a redeploy
//	/zero-downtime  a live page that polls /info, added in 1.2.0, to watch a deploy switch
//	/info    the same facts as JSON
//	/health  "ok" (point CommyPanel's health check here)
//
// The page shows the container's hostname. Every deployment runs in a new
// container, so the hostname changes after a redeploy: that is how you can see
// CommyPanel switch from the old version to the new one.
package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"time"
)

// Change this, commit and push, then press Deploy to see a new release go live.
const version = "1.2.0"

var started = time.Now()

type info struct {
	App      string `json:"app"`
	Version  string `json:"version"`
	Greeting string `json:"greeting"`
	Hostname string `json:"hostname"`
	Uptime   string `json:"uptime"`
}

func current() info {
	host, _ := os.Hostname()
	greeting := os.Getenv("GREETING")
	if greeting == "" {
		greeting = "Hello from CommyPanel!"
	}
	return info{
		App:      "hello-go",
		Version:  version,
		Greeting: greeting,
		Hostname: host,
		Uptime:   time.Since(started).Round(time.Second).String(),
	}
}

func main() {
	// Set FAIL_STARTUP=1 as an environment variable in CommyPanel to test a
	// failing deployment: the new version never answers, the deploy fails and
	// the previous version keeps serving.
	if os.Getenv("FAIL_STARTUP") != "" {
		log.Fatal("FAIL_STARTUP is set: refusing to start (this is how you test a failed deploy)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		i := current()
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>%[1]s</title></head>
<body style="font-family: system-ui, sans-serif; max-width: 40rem; margin: 4rem auto; padding: 0 1rem">
<h1>%[1]s</h1>
<p>Version <strong>%[2]s</strong> of <code>%[3]s</code>.</p>
<p>Container: <code>%[4]s</code><br>Up for: %[5]s</p>
<p>New in 1.1.0: <a href="/about">the about page</a>.<br>New in 1.2.0: <a href="/zero-downtime">the zero downtime page</a>.</p>
</body>
</html>
`, html.EscapeString(i.Greeting), html.EscapeString(i.Version), html.EscapeString(i.App), html.EscapeString(i.Hostname), html.EscapeString(i.Uptime))
	})
	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		i := current()
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>About %[1]s</title></head>
<body style="font-family: system-ui, sans-serif; max-width: 40rem; margin: 4rem auto; padding: 0 1rem">
<h1>About</h1>
<p><code>%[1]s</code> is a tiny Go app used to try out CommyPanel. You are looking at version <strong>%[2]s</strong>,
served by container <code>%[3]s</code>.</p>
<ul>
<li><a href="/">Home</a></li>
<li><a href="/zero-downtime">Zero downtime deployment</a></li>
<li><a href="/info">/info</a> (JSON)</li>
<li><a href="/health">/health</a></li>
</ul>
</body>
</html>
`, html.EscapeString(i.App), html.EscapeString(i.Version), html.EscapeString(i.Hostname))
	})
	mux.HandleFunc("GET /zero-downtime", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// No Cache-Control games: the page only polls /info, which is never cached.
		_, _ = w.Write([]byte(zeroDowntimePage))
	})
	mux.HandleFunc("GET /info", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(current())
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	log.Printf("hello-go %s listening on :%s", version, port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// zeroDowntimePage polls /info a few times a second and shows which container
// answered, so you can press Deploy in CommyPanel and watch the switch: the
// container name changes once, and "failed" stays at 0 if nothing went down.
const zeroDowntimePage = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Zero downtime deployment</title></head>
<body style="font-family: system-ui, sans-serif; max-width: 40rem; margin: 4rem auto; padding: 0 1rem">
<h1>Zero downtime deployment</h1>
<p>This page asks <code>/info</code> every 300 ms. Press <strong>Deploy</strong> in CommyPanel (change <code>version</code> in
<code>main.go</code> first, push, then deploy) and watch it: the container changes once, and <em>failed</em> stays at 0
if the switch really had no downtime.</p>
<table style="border-collapse: collapse; width: 100%">
<tr><td>Answering now</td><td><strong id="host">-</strong> (version <span id="ver">-</span>)</td></tr>
<tr><td>Requests</td><td id="ok">0</td></tr>
<tr><td style="color: #b91c1c">Failed</td><td id="bad" style="color: #b91c1c">0</td></tr>
<tr><td>Container switches seen</td><td id="sw">0</td></tr>
</table>
<h2 style="font-size: 1.1rem">Switches</h2>
<ol id="log" style="padding-left: 1.2rem"><li>Waiting for the first answer...</li></ol>
<p><a href="/">Home</a> · <a href="/about">About</a></p>
<script>
let ok = 0, bad = 0, switches = 0, last = null;
const $ = (id) => document.getElementById(id);
async function poll() {
  try {
    const r = await fetch('/info', { cache: 'no-store' });
    if (!r.ok) throw new Error(String(r.status));
    const i = await r.json();
    ok++;
    $('host').textContent = i.hostname;
    $('ver').textContent = i.version;
    if (last !== null && i.hostname !== last) {
      switches++;
      const li = document.createElement('li');
      li.textContent = new Date().toLocaleTimeString() + ': ' + last + ' -> ' + i.hostname + ' (version ' + i.version + ')';
      $('log').appendChild(li);
    } else if (last === null) {
      $('log').innerHTML = '';
    }
    last = i.hostname;
  } catch (e) {
    bad++;
  }
  $('ok').textContent = ok;
  $('bad').textContent = bad;
  $('sw').textContent = switches;
}
setInterval(poll, 300);
poll();
</script>
</body>
</html>
`
