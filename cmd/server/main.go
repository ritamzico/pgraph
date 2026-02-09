package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"

	pgraph "github.com/ritamzico/pgraph"
)

// store holds named graphs, protected by a read-write mutex.
type store struct {
	mu     sync.RWMutex
	graphs map[string]*pgraph.PGraph
}

func newStore() *store {
	return &store{graphs: make(map[string]*pgraph.PGraph)}
}

func (s *store) get(name string) (*pgraph.PGraph, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.graphs[name]
	return g, ok
}

func (s *store) set(name string, g *pgraph.PGraph) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graphs[name] = g
}

func (s *store) delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.graphs[name]; !ok {
		return false
	}
	delete(s.graphs, name)
	return true
}

func (s *store) names() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.graphs))
	for n := range s.graphs {
		names = append(names, n)
	}
	return names
}


func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// nameFromPath extracts the graph name from paths like /graphs/{name} or
// /graphs/{name}/query.
func nameFromPath(path string) string {
	// path is always rooted at /graphs/
	tail := strings.TrimPrefix(path, "/graphs/")
	if i := strings.Index(tail, "/"); i >= 0 {
		return tail[:i]
	}
	return tail
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	s := newStore()
	mux := http.NewServeMux()

	// GET /graphs — list graph names
	// POST /graphs/{name} — load graph from JSON body
	// DELETE /graphs/{name} — unload graph
	// POST /graphs/{name}/query — execute DSL query
	mux.HandleFunc("/graphs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string][]string{"graphs": s.names()})
	})

	mux.HandleFunc("/graphs/", func(w http.ResponseWriter, r *http.Request) {
		name := nameFromPath(r.URL.Path)
		if name == "" {
			writeError(w, http.StatusBadRequest, "graph name required")
			return
		}

		isQuery := strings.HasSuffix(r.URL.Path, "/query")

		switch {
		case r.Method == http.MethodPost && isQuery:
			pg, ok := s.get(name)
			if !ok {
				writeError(w, http.StatusNotFound, fmt.Sprintf("no graph %q", name))
				return
			}
			var body struct {
				DSL string `json:"dsl"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
			res, err := pg.Query(body.DSL)
			if err != nil {
				writeError(w, http.StatusUnprocessableEntity, err.Error())
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

		case r.Method == http.MethodPost:
			pg, err := pgraph.Load(r.Body)
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid graph JSON: %v", err))
				return
			}
			s.set(name, pg)
			writeJSON(w, http.StatusCreated, map[string]any{
				"name":  name,
				"nodes": len(pg.Graph.GetNodes()),
				"edges": len(pg.Graph.GetEdges()),
			})

		case r.Method == http.MethodDelete:
			if !s.delete(name) {
				writeError(w, http.StatusNotFound, fmt.Sprintf("no graph %q", name))
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("pgraph server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "server error: %v\n", err)
	}
}
