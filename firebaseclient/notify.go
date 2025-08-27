package firebaseclient

import (
	"abs-be/database"
	"abs-be/models"
	"context"
	"fmt"
	"log"
	"strings"

	"firebase.google.com/go/messaging"
	"gorm.io/gorm"
)

func SendToTokens(ctx context.Context, title, body string, data map[string]string, tokens []string) error {
    client := MessagingClient()
    if client == nil {
        return fmt.Errorf("firebase client not initialized")
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

        resp, err := client.SendMulticast(ctx, msg)
        if err != nil {
            log.Printf("FCM SendMulticast error: %v", err)
            continue
        }

        for idx, r := range resp.Responses {
            if !r.Success {
                tok := chunk[idx]
                if r.Error != nil {
                    errStr := r.Error.Error()
                    if strings.Contains(errStr, "registration-token-not-registered") ||
                        strings.Contains(errStr, "invalid-registration-token") ||
                        strings.Contains(errStr, "not-registered") {

                        if err := database.DB.Where("token = ?", tok).Delete(&models.DeviceToken{}).Error; err != nil && err != gorm.ErrRecordNotFound {
                            log.Printf("failed delete device token %s: %v", tok, err)
                        } else {
                            log.Printf("removed invalid token: %s", tok)
                        }
                    } else {
                        log.Printf("fcm send error for token %s: %v", tok, r.Error)
                    }
                }
            }
        }
    }

    return nil
}
