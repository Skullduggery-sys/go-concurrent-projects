package main

import (
	"encoding/json"
	"net/http"
	"time"

	"sharded-cache/internal/cache/sharded_cache"
)

func main() {
	cache := sharded_cache.New(32)

	mux := http.NewServeMux()

	// GET /get?key=somekey - получить значение по ключу
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key parameter is required", http.StatusBadRequest)
			return
		}

		value, ok := cache.Get(key)
		if !ok {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"key":   key,
			"value": value,
		})
	})

	// POST /set - установить значение
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			TTL   int64  `json:"ttl,omitempty"` // опциональный TTL в наносекундах
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Key == "" {
			http.Error(w, "key is required", http.StatusBadRequest)
			return
		}

		cache.Set(req.Key, req.Value, time.Duration(req.TTL))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"key":    req.Key,
		})
	})

	// Запускаем сервер
	port := ":8080"
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	println("Cache server starting on http://localhost" + port)
	println("Endpoints:")
	println("  GET    /get?key={key}     - get value")
	println("  POST   /set               - set value (JSON: {\"key\":\"k\",\"value\":\"v\", \"ttl\":\"600000000\"})")

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

	cache.Stop()
}
