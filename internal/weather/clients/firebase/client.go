package firebase

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"google.golang.org/api/option"
)

type Client struct {
	*messaging.Client
}

func New() (*Client, error) {
	opt := option.WithCredentialsFile("./creds.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error creating firebase client: %w", err)
	}

	m, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error creating messaging client: %w", err)
	}

	return &Client{
		m,
	}, nil
}
