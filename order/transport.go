package order

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func MakeHTTPHandler(service Service) *fiber.App {
	app := fiber.New()
	app.Post("/orders", orders(service))
	app.Get("/health", health(service))
	return app
}

func orders(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		newOrderResource := new(NewOrderResource)
		if err := json.Unmarshal(c.Body(), newOrderResource); err != nil {
			return err
		}

		if newOrderResource.Address == "" ||
			newOrderResource.Customer == "" ||
			newOrderResource.Card == "" ||
			newOrderResource.Items == "" {
			return fiber.ErrBadRequest
		}

		itemsRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, newOrderResource.Items, nil)
		if err != nil {
			return err
		}
		itemsResponse, err := http.DefaultClient.Do(itemsRequest)
		if err != nil {
			return err
		}
		defer itemsResponse.Body.Close()

		items := new([]Item)
		if err := json.NewDecoder(itemsResponse.Body).Decode(items); err != nil {
			return err
		}

		addressRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, newOrderResource.Address, nil)
		if err != nil {
			return err
		}
		addressResponse, err := http.DefaultClient.Do(addressRequest)
		if err != nil {
			return err
		}
		defer addressResponse.Body.Close()
		address := new(Address)
		json.NewDecoder(addressResponse.Body).Decode(address)

		customerRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, newOrderResource.Customer, nil)
		if err != nil {
			return err
		}
		customerResponse, err := http.DefaultClient.Do(customerRequest)
		if err != nil {
			return err
		}
		defer customerResponse.Body.Close()
		customer := new(Customer)
		json.NewDecoder(customerResponse.Body).Decode(customer)

		cardRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, newOrderResource.Card, nil)
		if err != nil {
			return err
		}
		cardResponse, err := http.DefaultClient.Do(cardRequest)
		if err != nil {
			return err
		}
		defer cardResponse.Body.Close()
		card := new(Card)
		json.NewDecoder(cardResponse.Body).Decode(card)

		amount := calculateTotal(items)
		paymentRequestBody := PaymentRequest{
			Address:  *address,
			Customer: *customer,
			Card:     *card,
			Amount:   amount,
		}
		b, err := json.Marshal(paymentRequestBody)
		if err != nil {
			return err
		}
		paymentRequest, err := http.NewRequest(http.MethodPost, "", bytes.NewReader(b))
		if err != nil {
			return err
		}
		paymentResponse, err := http.DefaultClient.Do(paymentRequest)
		if err != nil {
			return err
		}
		defer paymentResponse.Body.Close()
		paymentResponseBody := new(PaymentResponse)
		if err := json.NewDecoder(paymentResponse.Body).Decode(paymentResponseBody); err != nil {
			return err
		}

		if !paymentResponseBody.Authorised {
			return fiber.ErrUnauthorized
		}

		shipment := Shipment{ID: customer.ID}
		customerOrder := CustomerOrder{
			ID:         "",
			CustomerID: customer.ID,
			Customer:   *customer,
			Address:    *address,
			Card:       *card,
			Items:      *items,
			Shipment:   shipment,
			Date:       time.Now(),
			Total:      amount,
		}

		if err := db.CreateOrder(ctx, customerOrder); err != nil {
			return err
		}

		return c.JSON(customerOrder)
	}
}

func health(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		health := service.Ping(ctx)
		response := struct {
			Health []HealthCheck `json:"health"`
		}{
			Health: health,
		}
		return c.JSON(response)
	}
}
