package payment

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func MakeHTTPHandler(service Service, logger *log.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/paymentauth", auth(service))
	r.Get("/health", health(service))
	return r
}

func auth(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		amount, _ := decodeAuthoriseRequest(ctx, r)
		authorisation, err := service.Authorise(amount)
		resp := authoriseResponse{authorisation, err}
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
}

func decodeAuthoriseRequest(_ context.Context, r *http.Request) (float32, error) {
	var bodyBytes []byte
	if r.Body != nil {
		var err error
		bodyBytes, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return 0.0, err
		}
	}
	bodyString := string(bodyBytes)

	var amount float32
	if err := json.Unmarshal(bodyBytes, &amount); err != nil {
		return 0.0, err
	}
	if amount == 0.0 {
		return 0.0, &UnmarshalKeyError{Key: "amount", JSON: bodyString}
	}

	return amount, nil
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
