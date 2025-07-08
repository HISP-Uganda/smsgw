package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"smsgw/config"
	"smsgw/utils"
	"strconv"
	"strings"
	"time"
)

type NotificationController struct{}

func (n *NotificationController) NotificationHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload map[string]interface{}
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		// Parse query params
		templateID := c.Query("template_id")
		trigger := c.Query("trigger")
		daysOffsetStr := c.Query("days_offset")

		// Find template by ID (primary approach)
		var tmpl *config.ProgramNotificationTemplate
		if templateID != "" {
			tmpl = config.FindTemplateByID(cfg, templateID)
			if tmpl == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
				return
			}
		} else {
			// Optionally, search by trigger and days_offset if templateID not given
			var daysOffset *int
			if daysOffsetStr != "" {
				if v, err := strconv.Atoi(daysOffsetStr); err == nil {
					daysOffset = &v
				}
			}
			for _, t := range cfg.Templates.ProgramNotificationTemplates {
				if (trigger == "" || strings.EqualFold(t.NotificationTrigger, trigger)) &&
					(daysOffset == nil || t.RelativeScheduledDays == *daysOffset) {
					tmpl = &t
					break
				}
			}
			if tmpl == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Template not found by trigger/days_offset"})
				return
			}
		}

		// Compute due_date if SCHEDULED_DAYS_DUE_DATE
		currentDateStr, _ := payload["CURRENT_DATE"].(string)
		currentDate, _ := time.Parse("2006-01-02", currentDateStr)
		dueDate := ""
		if tmpl.NotificationTrigger == "SCHEDULED_DAYS_DUE_DATE" {
			due := currentDate.AddDate(0, 0, -tmpl.RelativeScheduledDays)
			dueDate = due.Format("2006-01-02")
		}

		// Consent logic via helper
		consentValue := ""
		consentAttr := cfg.Templates.ConsentAttribute
		if consentAttr != "" {
			if v, ok := payload[consentAttr]; ok && v != nil {
				consentValue = fmt.Sprintf("%v", v)
			}
		}

		recipientAttrs := utils.FilterRecipientAttributes(
			tmpl.RecipientAttributes,
			cfg.Templates.ConsentIgnoreAttributes,
			consentValue,
		)

		// Extract recipient numbers using RecipientAttributes
		phoneNumbers := utils.ExtractUniquePhones(payload, recipientAttrs, "UG")
		if len(phoneNumbers) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid recipient phone numbers"})
			return
		}

		// Detect language for message template
		lang := utils.DetectLanguage(payload, cfg.Templates.LanguageAttribute)
		messageTemplate := tmpl.MessageTemplates[lang]
		if messageTemplate == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":               "No message template found for requested language",
				"template_id":         tmpl.ID,
				"requested_language":  lang,
				"available_languages": utils.MapKeys(tmpl.MessageTemplates),
			})
			return
		}

		// Prepare the payload for template substitution
		msgPayload := make(map[string]interface{})
		for k, v := range payload {
			msgPayload[k] = v
		}
		if dueDate != "" {
			msgPayload["due_date"] = dueDate
		}

		message := config.SubstituteTemplate(messageTemplate, msgPayload)

		// Business rules
		//if !config.TemplateIsAllowedToSend(*tmpl, payload, dueDate) {
		//	c.JSON(http.StatusOK, gin.H{"status": "Not allowed to send, per business rules"})
		//	return
		//}

		// Send SMS
		//if err := sendSMS(phoneNumbers, message); err != nil {
		//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send SMS"})
		//	return
		//}
		if config.AppConfig.Server.InTestMode {
			successful, failed := SendTestSMSorTelegram(
				phoneNumbers,
				message,
				config.AppConfig.Telegram.TelegramBots,
				defaultTelegramBot,
				sendTelegramMessage, // your function (e.g., func sendTelegramMessage(chatID int64, token, msg string) error)
			)

			c.JSON(http.StatusOK, gin.H{
				"status":      "Test mode: Sent via Telegram",
				"recipients":  successful,
				"failed":      failed,
				"template_id": tmpl.ID,
				"lang":        lang,
			})
			return
		}

		log.Printf("Sending SMS to %v: %s", phoneNumbers, message)
		successfulSends, failedSends := SendBulkSMSWithSMSOne(
			phoneNumbers,
			message,
			config.AppConfig.SMSOne,
			smsOneClient,
			sendUsingSMSOne,
		)

		c.JSON(http.StatusOK, gin.H{
			"status":      "Notification sent",
			"recipients":  successfulSends,
			"failed":      failedSends,
			"template_id": tmpl.ID,
			"lang":        lang,
		})
	}
}
