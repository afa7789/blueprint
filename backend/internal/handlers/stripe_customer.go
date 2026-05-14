package handlers

import (
	"context"
	"fmt"

	"github.com/afa/blueprint/backend/internal/domain"
	stripe "github.com/stripe/stripe-go/v82"
	stripecustomer "github.com/stripe/stripe-go/v82/customer"
)

// GetOrCreateStripeCustomer returns the user's Stripe customer ID, creating
// a new Customer in Stripe and persisting it on the user record if one
// does not yet exist. The provided user struct is mutated to reflect the
// persisted customer ID so the caller sees the new value.
func GetOrCreateStripeCustomer(ctx context.Context, users domain.UserRepository, u *domain.User) (string, error) {
	if u.StripeCustomerID != nil && *u.StripeCustomerID != "" {
		return *u.StripeCustomerID, nil
	}

	params := &stripe.CustomerParams{Email: stripe.String(u.Email)}
	if u.Name != nil && *u.Name != "" {
		params.Name = stripe.String(*u.Name)
	}
	customer, err := stripecustomer.New(params)
	if err != nil {
		return "", fmt.Errorf("create stripe customer: %w", err)
	}

	if err := users.UpdateStripeCustomerID(ctx, u.ID, customer.ID); err != nil {
		return "", fmt.Errorf("persist stripe customer id: %w", err)
	}
	u.StripeCustomerID = &customer.ID
	return customer.ID, nil
}
