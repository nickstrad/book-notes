package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	var mu sync.Mutex
	mode := "ok"
	records := []map[string]any{}
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		secret = "report-secret"
	}
	m := http.NewServeMux()
	m.HandleFunc("/admin/mode", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Method == "POST" {
			var in struct {
				Mode string `json:"mode"`
			}
			json.NewDecoder(r.Body).Decode(&in)
			switch in.Mode {
			case "ok", "500", "slow", "timeout":
				mode = in.Mode
			default:
				http.Error(w, "unknown mode", 400)
				return
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"mode": mode})
	})
	m.HandleFunc("/admin/records", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Method == "DELETE" {
			records = []map[string]any{}
		}
		json.NewEncoder(w).Encode(map[string]any{"records": records})
	})
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(b)
		sig, e := hex.DecodeString(r.Header.Get("X-Queuelet-Signature"))
		valid := e == nil && hmac.Equal(sig, mac.Sum(nil))
		mu.Lock()
		current := mode
		status := 200
		if current == "500" {
			status = 500
		}
		var body any
		if json.Unmarshal(b, &body) != nil {
			body = string(b)
		}
		records = append(records, map[string]any{"at": time.Now(), "body": body, "raw_body": string(b), "signature_valid": valid, "status": status})
		mu.Unlock()
		if !valid {
			log.Printf("signature did not verify")
		}
		if current == "slow" {
			time.Sleep(3 * time.Second)
		}
		if current == "timeout" {
			time.Sleep(10 * time.Second)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		io.WriteString(w, `{"ok":true}`)
	})
	log.Fatal(http.ListenAndServe(":8201", m))
}
