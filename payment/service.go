package payment

import (
	"errors"
	"fmt"
	"time"
)

type Middleware func(Service) Service

type Service interface {
	Authorise(total float32) (Authorisation, error)
	Health() []Health
}

type Authorisation struct {
	Authorised bool   `json:"authorised"`
	Message    string `json:"message"`
}

type Health struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Time    string `json:"time"`
}

type service struct {
	declineOverAmount float32
}

func NewAuthorisationService(declineOverAmount float32) Service {
	return &service{declineOverAmount}
}

func (s *service) Authorise(amount float32) (Authorisation, error) {
	if amount <= 0 {
		return Authorisation{}, errors.New("Invalid payment amount")
	}

	authorised := amount <= s.declineOverAmount
	var message string
	if authorised {
		message = "Payment authorised"
	} else {
		message = fmt.Sprintf("Payment declined: amount exceeds %.2f", s.declineOverAmount)
	}

	return Authorisation{
		Authorised: authorised,
		Message:    message,
	}, nil
}

func (s *service) Health() []Health {
	var health []Health
	app := Health{"payment", "OK", time.Now().String()}
	health = append(health, app)
	return health
}
