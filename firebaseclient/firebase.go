package firebaseclient

import (
    "context"
    "sync"

    firebase "firebase.google.com/go/v4"
    "firebase.google.com/go/v4/messaging"
    "google.golang.org/api/option"
)

var (
    once      sync.Once
    msgClient *messaging.Client
)

func InitFirebase(saFilePath string) error {
    var initErr error
    once.Do(func() {
        opt := option.WithCredentialsFile(saFilePath)

        app, err := firebase.NewApp(context.Background(), nil, opt)
        if err != nil {
            initErr = err
            return
        }

        client, err := app.Messaging(context.Background())
        if err != nil {
            initErr = err
            return
        }
        msgClient = client
    })
    return initErr
}

func MessagingClient() *messaging.Client {
    return msgClient
}
