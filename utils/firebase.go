package utils

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App

func InitializeFirebase() {
	opt := option.WithCredentialsFile("firebase/abs-be-firebase-adminsdk-fbsvc-8c806f8ecc.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	FirebaseApp = app
}

func GetMessagingClient() (*messaging.Client, error) {
	return FirebaseApp.Messaging(context.Background())
}