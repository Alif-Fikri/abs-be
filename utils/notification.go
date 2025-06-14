package utils

import (
	"context"
	"log"

	"firebase.google.com/go/messaging"
)

func SendNotification(token string, title string, body string) error {
	client, err := GetMessagingClient()
	if err != nil {
		return err
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token,
	}

	response, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}

	log.Printf("successfully : %v\n", response)
	return nil
}

func SendNotificationToUser(userID uint, title string, body string) {
	var sessions []models.Session
	if err := database.DB.Where("user_id = ? AND fcm_token IS NOT NULL", userID).Find(&sessions).Error; err != nil {
		log.Printf("gagal mendapatkan FCM token: %v", err)
		return
	}

	for _, session := range sessions {
		if err := SendNotification(session.FCMToken, title, body); err != nil {
			log.Printf("gagal mengirim notifikasi: %v", err)
		}
	}
}