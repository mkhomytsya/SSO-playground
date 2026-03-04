package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// Claims represents the OIDC claims forwarded by oauth2-proxy.
type Claims struct {
	User              string `json:"user,omitempty"`
	Email             string `json:"email,omitempty"`
	PreferredUsername  string `json:"preferred_username,omitempty"`
	Groups            string `json:"groups,omitempty"`
	AccessToken       string `json:"access_token,omitempty"`
}

// Response is the JSON body returned to the caller.
type Response struct {
	Message    string            `json:"message"`
	Claims     Claims            `json:"claims"`
	AllHeaders map[string]string `json:"all_forwarded_headers"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", claimsHandler)
	mux.HandleFunc("/healthz", healthHandler)

	log.Printf("demo-app listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func claimsHandler(w http.ResponseWriter, r *http.Request) {
	// Collect all X-Forwarded-* headers set by oauth2-proxy.
	forwarded := make(map[string]string)
	for name, values := range r.Header {
		if len(name) > 12 && name[:12] == "X-Forwarded-" {
			forwarded[name] = values[0]
		}
	}

	resp := Response{
		Message: "Authenticated via OIDC (Pocket ID → oauth2-proxy)",
		Claims: Claims{
			User:             r.Header.Get("X-Forwarded-User"),
			Email:            r.Header.Get("X-Forwarded-Email"),
			PreferredUsername: r.Header.Get("X-Forwarded-Preferred-Username"),
			Groups:           r.Header.Get("X-Forwarded-Groups"),
			AccessToken:      r.Header.Get("X-Forwarded-Access-Token"),
		},
		AllHeaders: forwarded,
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		http.Error(w, `{"error":"encoding failed"}`, http.StatusInternalServerError)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
