package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"smsgw/config"
	"smsgw/utils"
	"strings"
)

func sendTelegramMessage(chatID int64, token, message string) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}
	bot.Debug = true

	msg := tgbotapi.NewMessage(chatID, message)
	_, err = bot.Send(msg)
	return err
}

func SendTestSMSorTelegram(
	uniqueNumbers []string,
	message string,
	bots map[string]config.TelegramBot,
	defaultBot config.TelegramBot,
	sendTelegramMessage func(chatID int64, token, message string) error,
) (successful, failed []string) {

	for _, number := range uniqueNumbers {
		fmt.Printf("Sending SMS to %s: %s\n", number, message)
		if bot, ok := bots[number]; ok {
			err := sendTelegramMessage(bot.ChatID, bot.Token, message)
			if err != nil {
				failed = append(failed, number)
				continue
			}
			successful = append(successful, number)
		} else {
			// send to default bot, include the intended number in the message
			testMsg := fmt.Sprintf("%s\nMeant For: %s", message, number)
			err := sendTelegramMessage(defaultBot.ChatID, defaultBot.Token, testMsg)
			if err != nil {
				failed = append(failed, number)
				continue
			}
			successful = append(successful, number)
		}
	}
	return
}

type TelegramController struct{}

func (t *TelegramController) SendSMS(c *gin.Context) {
	phoneNumbers := c.Query("to")
	message := c.Query("text")
	// from := c.Query("from")

	if phoneNumbers == "" || message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both phoneNumber and message are required"})
		return
	}

	// Split the phone numbers by comma
	numbers := strings.Split(phoneNumbers, ",")

	// Remove duplicates
	uniqueNumbers := utils.RemoveDuplicates(numbers)

	// create a slice to keep track of failed sends
	failedSends := make([]string, 0)
	successfulSends := make([]string, 0)

	// for each unique number print it
	for _, number := range uniqueNumbers {
		fmt.Printf("Sending SMS to %s: %s\n", number, message)
		if bot, ok := config.AppConfig.Telegram.TelegramBots[number]; ok {
			err := sendTelegramMessage(bot.ChatID, bot.Token, message)
			if err != nil {
				failedSends = append(failedSends, number)

				continue
			}
			successfulSends = append(successfulSends, number)
		} else {
			// send to default bot
			message = fmt.Sprintf("%s\n Meant For: %s", message, number)
			err := sendTelegramMessage(
				config.AppConfig.Telegram.DefaultBot.ChatID,
				config.AppConfig.Telegram.DefaultBot.Token, message)
			if err != nil {
				failedSends = append(failedSends, number)
				continue
			}
			successfulSends = append(successfulSends, number)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "SMS sent successfully",
		"phoneNumbers":    uniqueNumbers,
		"message":         message,
		"sentTo":          successfulSends,
		"failedSendingTo": failedSends,
	})
}
