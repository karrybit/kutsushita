package catalogue

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

func MakeHTTPHandler(service Service, imagePath string) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/catalogue", func(r chi.Router) {
		r.Get("/", list(service))
		r.Get("/size", size(service))
		r.Get("/{id}", id(service))
	})
	r.Get("/tags", tags(service))
	r.Get("/health", health(service))
	return r
}

func list(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context()
		req, _ := decodeListRequest(r)
		socks, err := service.List(req.Tags, req.Order, req.PageNum, req.PageSize)
		resp := listResponse{socks, err}
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
}

func size(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context()
		req, _ := decodeCountRequest(r)
		n, err := service.Count(req.Tags)
		resp := countResponse{n, err}
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
}

func id(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, ok := ctx.Value("id").(string)
		if !ok {
			http.NotFound(w, r)
			return
		}
		sock, err := service.Get(id)
		resp := getResponse{sock, err}
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
}

func tags(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context()
		tags, err := service.Tags()
		resp := tagsResponse{Tags: tags, Err: err}
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
}

func health(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context()
		health := service.Health()
		resp := healthResponse{health}
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
}
