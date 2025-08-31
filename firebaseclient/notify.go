package firebaseclient

import (
	"abs-be/database"
	"abs-be/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"firebase.google.com/go/v4/messaging"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func SendToTokens(ctx context.Context, title, body string, data map[string]string, tokens []string) error {
	client := MessagingClient()
	if client == nil {
		return fmt.Errorf("firebase client not initialized")
	}
	if len(tokens) == 0 {
		return nil
	}

	chunkSize := 500
	for i := 0; i < len(tokens); i += chunkSize {
		end := i + chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}
		chunk := tokens[i:end]

		msg := &messaging.MulticastMessage{
			Tokens: chunk,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Webpush: &messaging.WebpushConfig{
				Notification: &messaging.WebpushNotification{
					Title: title,
					Body:  body,
				},
			},
		}

		resp, err := client.SendEachForMulticast(ctx, msg)
		if err != nil {
			log.Printf("SendMulticast error: %v", err)
			if s, ok := status.FromError(err); ok {
				log.Printf("gRPC status: code=%v message=%q", s.Code(), s.Message())
			}

			log.Printf("SendMulticast failed, falling back to individual sends for %d tokens", len(chunk))
			fallbackSendPerToken(ctx, client, title, body, data, chunk)
			continue
		}

		for idx, r := range resp.Responses {
			if !r.Success {
				tok := chunk[idx]
				if r.Error != nil {
					handleFCMErrorForToken(tok, r.Error)
				}
			}
		}
	}

	return nil
}

func fallbackSendPerToken(ctx context.Context, client *messaging.Client, title, body string, data map[string]string, tokens []string) {
	for _, tok := range tokens {
		m := &messaging.Message{
			Token: tok,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Webpush: &messaging.WebpushConfig{
				Notification: &messaging.WebpushNotification{
					Title: title,
					Body:  body,
				},
			},
		}
		_, err := client.Send(ctx, m)
		if err != nil {
			log.Printf("Send to token failed: %s -> %v", tok, err)
			handleFCMErrorForToken(tok, err)
			continue
		}
		// time.Sleep(10 * time.Millisecond)
	}
}

func handleFCMErrorForToken(tok string, err error) {
	errStr := err.Error()
	log.Printf("FCM error for token %s: %v", tok, errStr)
	if strings.Contains(errStr, "registration token not registered") ||
		strings.Contains(errStr, "invalid registration token") ||
		strings.Contains(errStr, "not registered") {
		if err := database.DB.Where("token = ?", tok).Delete(&models.DeviceToken{}).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Printf("failed delete device token %s: %v", tok, err)
		} else {
			log.Printf("removed invalid token: %s", tok)
		}
	}
	if s, ok := status.FromError(err); ok {
		log.Printf("FCM grpc status: code=%v message=%q", s.Code(), s.Message())
	}
}

var SendNotify = NotifyUsers

func NotifyUsers(ctx context.Context, typeStr, title, body string, payload map[string]interface{}, userIDs []uint) error {
	uniqMap := make(map[uint]struct{}, len(userIDs))
	var recipients []uint
	for _, id := range userIDs {
		if id == 0 {
			continue
		}
		if _, ok := uniqMap[id]; !ok {
			uniqMap[id] = struct{}{}
			recipients = append(recipients, id)
		}
	}

	if len(recipients) == 0 {
		return nil
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("notify: gagal marshal payload: %v", err)
		payloadBytes = []byte("{}")
	}

	now := time.Now()
	notifs := make([]models.Notification, 0, len(recipients))
	for _, rid := range recipients {
		notifs = append(notifs, models.Notification{
			Title:     title,
			Body:      body,
			Type:      typeStr,
			Payload:   string(payloadBytes),
			Recipient: rid,
			Read:      false,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	if err := database.DB.Create(&notifs).Error; err != nil {
		log.Printf("notify: gagal menyimpan notifications: %v", err)
	}

	var tokens []string
	if err := database.DB.
		Model(&models.DeviceToken{}).
		Where("user_id IN ?", recipients).
		Pluck("token", &tokens).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("notify: gagal ambil device tokens: %v", err)
		return fmt.Errorf("gagal ambil device tokens: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	dataMap := map[string]string{
		"type":    typeStr,
		"payload": string(payloadBytes),
	}

	if err := SendToTokens(ctx, title, body, dataMap, tokens); err != nil {
		log.Printf("notify: SendToTokens error: %v", err)
		return err
	}

	return nil
}
