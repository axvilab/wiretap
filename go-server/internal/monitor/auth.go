package monitor

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/axvi/wiretap/internal/storage"
)

// applyAuth adds authentication to an HTTP request based on site config.
func applyAuth(req *http.Request, site storage.MonitorSite) {
	if site.AuthType == "" || site.AuthType == "none" {
		return
	}

	var auth storage.SiteAuth
	if err := json.Unmarshal([]byte(site.AuthJSON), &auth); err != nil {
		return
	}

	switch site.AuthType {
	case "bearer":
		if auth.Token != "" {
			req.Header.Set("Authorization", "Bearer "+auth.Token)
		}
	case "basic":
		if auth.Username != "" {
			encoded := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
			req.Header.Set("Authorization", "Basic "+encoded)
		}
	case "header":
		if auth.Header != "" && auth.Value != "" {
			req.Header.Set(auth.Header, auth.Value)
		}
	case "query":
		if auth.Param != "" {
			q := req.URL.Query()
			q.Set(auth.Param, auth.Value)
			req.URL.RawQuery = q.Encode()
		}
	}
}
