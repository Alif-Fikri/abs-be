package main

import (
	"abs-be/database"
	"abs-be/firebaseclient"
	"abs-be/routes"
	"context"
	"fmt"
	"io/ioutil"
	"log"

	// "log"
	"os"

	// "firebase.google.com/go/v4/messaging"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/google"
)

func main() {
	database.Konek()

	r := gin.Default()
	r.Use(cors.Default())
	routes.Api(r)
	if err := firebaseclient.InitFirebase(os.Getenv("FIREBASE_SA_FILE")); err != nil {
		log.Fatalf("firebase init error: %v", err)
	}
	if firebaseclient.MessagingClient() == nil {
		log.Fatalf("firebase messaging client is nil after init")
	}
	log.Println("Firebase initialized OK")
	ctx := context.Background()
	saPath := "C:/Users/Alif Fikri/abs-be/firebase/abs-be-firebase-adminsdk-fbsvc-18ec9aa5ce.json"
	b, err := ioutil.ReadFile(saPath)
	if err != nil {
		log.Fatalf("read sa file error: %v", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, b, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		log.Fatalf("credentials from json error: %v", err)
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		log.Fatalf("get token error: %v", err)
	}

	fmt.Println("Access token retrieved, expiry:", token.Expiry)
	fmt.Println("Access token (prefix):", token.AccessToken[:30], "...")

	
	// client := firebaseclient.MessagingClient()
	// if client == nil {
	// 	log.Fatal("messaging client nil")
	// }

	// token := "<DEVICE_FCM_TOKEN>" // ganti dengan token valid dari device (web/Android)
	// msg := &messaging.Message{
	// 	Token: token,
	// 	Notification: &messaging.Notification{
	// 		Title: "Test single",
	// 		Body:  "Hello dari test single Send",
	// 	},
	// 	Data: map[string]string{
	// 		"test": "1",
	// 	},
	// }
	// res, err := client.Send(context.Background(), msg)
	// if err != nil {
	// 	log.Fatalf("Send failed: %v", err)
	// }
	// fmt.Println("Send OK, response:", res)
	// os.Exit(0)

	if err := r.Run(":8080"); err != nil {
		panic("gagal menjalankan server: " + err.Error())
	}
}
