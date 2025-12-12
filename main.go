package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"net/url"
)

type PackageInfo struct {
	Name    string
	Version string
	Author  string
}

type CacheEntry struct {
	Data        []PackageInfo
	ExpiresAt   time.Time
	LastUpdated time.Time
}

type PageData struct {
	Packages    []PackageInfo
	LastUpdated string
}

var (
	cache      CacheEntry
	cacheMutex sync.RWMutex

	packages = []string{
		"expr-eval/latest",
		"@ng-select/ng-select/8.3.0",
		"sweetalert2/11.10.1",
	}
)

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/refresh", handleRefresh)
	log.Println("Server running on :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := getCachedData()
	cacheMutex.RLock()
	lastUpdatedStr := timeAgo(cache.LastUpdated)
	cacheMutex.RUnlock()

	pageData := PageData{
		Packages:    data,
		LastUpdated: lastUpdatedStr,
	}
	tmpl := `<!DOCTYPE html>
<html>
<head>
<title>📦 go-pkgspy</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font:14px -apple-system,sans-serif;background:#f8f9fa;padding:20px}
.header{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px}
h1{color:#333;margin:0}
button{background:#007bff;color:white;border:none;padding:8px 16px;border-radius:4px;cursor:pointer}
button:hover{background:#0056b3}
table{width:100%;background:white;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.1);margin-bottom:60px}
th,td{padding:12px;text-align:left;border-bottom:1px solid #eee}
th{background:#f8f9fa;font-weight:600;color:#495057}
tr:hover{background:#f8f9fa}
td button{background:#28a745;padding:4px 8px;font-size:12px}
td button:hover{background:#1e7e34}
</style>
</head>
<body>
<div class="header">
<h1>📦 go-pkgspy</h1>
<button onclick="copyAllDeps()">Copy All Dependencies</button>
</div>
<p style="font-size:12px;color:#666;margin:0">updated {{.LastUpdated}}</p>
<table>
<tr><th>Package</th><th>Version</th><th>Author</th><th>Install</th></tr>
{{ range .Packages  }}<tr>
<td>{{ .Name }}</td>
<td>{{ .Version }}</td>
<td>{{ .Author }}</td>
<td><button onclick="copyCmd('npm install {{ .Name }}@{{ .Version }}')">Copy</button></td>
</tr>{{ end }}
</table>
<footer style="position:fixed;bottom:0;left:0;right:0;text-align:center;padding:10px;background:linear-gradient(transparent,rgba(248,249,250,0.9))">
<a href="https://github.com/yrmsa/go-pkgspy" target="_blank" style="color:#666;text-decoration:none" title="View on GitHub">
<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
</svg>
</a>
</footer>
</body>
<script>
function copyCmd(cmd){
  if(navigator.clipboard && window.isSecureContext){
    navigator.clipboard.writeText(cmd);
  } else {
    var textarea = document.createElement('textarea');
    textarea.value = cmd;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
  }
}
function copyAllDeps(){
  var deps={};
  {{ range .Packages  }}deps["{{ .Name }}"]="{{ .Version }}";
  {{ end }}
  var json = JSON.stringify(deps,null,2);
  if(navigator.clipboard && window.isSecureContext){
    navigator.clipboard.writeText(json);
  } else {
    var textarea = document.createElement('textarea');
    textarea.value = json;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
  }
}
</script>
</html>`
	t, _ := template.New("page").Parse(tmpl)
	t.Execute(w, pageData)
}

func timeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshCache()
	w.Write([]byte("Cache refreshed"))
}

func getCachedData() []PackageInfo {
	cacheMutex.RLock() // read lock
	if time.Now().Before(cache.ExpiresAt) && cache.Data != nil {
		data := cache.Data
		cacheMutex.RUnlock()
		return data
	}
	cacheMutex.RUnlock()
	return refreshCache()
}

func refreshCache() []PackageInfo {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	var wg sync.WaitGroup
	results := make([]PackageInfo, len(packages))

	for i, p := range packages {
		wg.Add(1)
		go func(i int, pkg string) {
			defer wg.Done()
			info, err := fetchVersion(pkg)
			if err != nil {
				log.Println("Error fetching", pkg, err)
				return
			}
			results[i] = info
		}(i, p)
	}

	wg.Wait()

	// Keep original order, skip empty results
	var filtered []PackageInfo
	for _, r := range results {
		if r.Name != "" {
			filtered = append(filtered, r)
		}
	}

	cache = CacheEntry{
		Data:        filtered,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		LastUpdated: time.Now(),
	}
	return filtered
}

func fetchVersion(pkg string) (PackageInfo, error) {
	parts := strings.Split(pkg, "/")
	var packageName, tag string

	// Handle scoped packages like @abc/react/latest
	if strings.HasPrefix(pkg, "@") && len(parts) >= 3 {
		// @abc/react/latest -> packageName: @abc/react, tag: latest
		packageName = strings.Join(parts[:2], "/")
		tag = parts[2]
	} else if len(parts) >= 2 {
		// react/latest -> packageName: react, tag: latest
		packageName = strings.Join(parts[:len(parts)-1], "/")
		tag = parts[len(parts)-1]
	} else {
		// react -> packageName: react, tag: latest
		packageName = pkg
		tag = "latest"
	}

	client := &http.Client{}

	esc := url.PathEscape(packageName)
	req, err := http.NewRequest("GET", "https://registry.npmjs.org/"+esc+"/"+tag, nil)
	if err != nil {
		return PackageInfo{}, err
	}

	// Set headers if needed
	req.Header.Set("Authorization", "Bearer <YOUR_TOKEN_HERE>")

	resp, err := client.Do(req)
	if err != nil {
		return PackageInfo{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PackageInfo{}, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return PackageInfo{}, err
	}

	version, ok := raw["version"].(string)
	if !ok {
		return PackageInfo{}, fmt.Errorf("invalid version data")
	}

	author := ""
	if a, ok := raw["author"].(map[string]interface{}); ok {
		if name, ok := a["name"].(string); ok {
			author = name
		}
	}

	return PackageInfo{Name: packageName, Version: version, Author: author}, nil
}
