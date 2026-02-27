package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	pgraph "github.com/ritamzico/pgraph"
)

var allowedOrigins = []string{
	"http://localhost:5173",
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func corsMiddleware(next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var body struct {
			Graph json.RawMessage `json:"graph"`
			DSL   string          `json:"dsl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if len(body.Graph) == 0 {
			writeError(w, http.StatusBadRequest, "missing field: graph")
			return
		}
		if body.DSL == "" {
			writeError(w, http.StatusBadRequest, "missing field: dsl")
			return
		}

		pg, err := pgraph.Load(bytes.NewReader(body.Graph))
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid graph: %v", err))
			return
		}

		res, err := pg.Query(body.DSL)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		// A nil result means the DSL statement was a mutation (CREATE/DELETE).
		// Serialize the updated graph so the client can persist the new state.
		if res == nil {
			var buf bytes.Buffer
			if err := pg.Save(&buf); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, struct {
				Kind string          `json:"kind"`
				Data json.RawMessage `json:"data"`
			}{Kind: "mutation", Data: json.RawMessage(buf.Bytes())})
			return
		}

		b, err := pgraph.MarshalResultJSON(res)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("pgraph server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "server error: %v\n", err)
	}
}
