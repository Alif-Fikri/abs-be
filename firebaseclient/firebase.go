package firebaseclient

import (
    "context"
    "sync"
    "os"
    "firebase.google.com/go"
    "firebase.google.com/go/messaging"
    "google.golang.org/api/option"
)

var (
    once sync.Once
    msgClient *messaging.Client
)

func InitFirebase(saFilePath string) error {
    var initErr error
    once.Do(func() {
        var opt option.ClientOption
        if saFilePath != "" {
            opt = option.WithCredentialsFile(saFilePath)
        } else if p := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); p != "" {
            opt = option.WithCredentialsFile(p)
        } else {
            opt = option.WithCredentialsJSON(nil)
        }

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
