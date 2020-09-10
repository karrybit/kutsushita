package payment

import (
	"context"
	"log"
	"net/http"
)

func WireUp(ctx context.Context, declineAmount float32, serviceName string) (http.Handler, *log.Logger) {
	logger := log.Logger{}
	_ = NewAuthorisationService(declineAmount)

	// TODO endpoints
	// TODO router

	return nil, &logger
}
