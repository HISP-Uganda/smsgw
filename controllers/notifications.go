package controllers

import (
	"encoding/json"
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
		payload, err := parsePayload(c)
		if err != nil {
			respondWithError(c, http.StatusBadRequest, "Invalid payload")
			return
		}

		if cfg.Server.Debug {
			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				log.Errorf("Failed to marshal payload to JSON: %v", err)
			} else {
				log.Debugf("Received payload: %s", jsonPayload)
			}
		}

		templateID, trigger, daysOffsetStr := c.Query("template_id"), c.Query("trigger"), c.Query("days_offset")
		tmpl, err := findTemplate(cfg, templateID, trigger, daysOffsetStr)
		if err != nil {
			respondWithError(c, http.StatusNotFound, err.Error())
			return
		}

		dueDate := computeDueDate(payload, tmpl)
		consentValue := extractConsentValue(payload, cfg.Templates.ConsentAttribute)

		if !isMessagingAllowed(payload, cfg) {
			respondWithStatus(c, "Not allowed to send due to ignore messaging attribute", tmpl.ID, cfg.Templates.AllowMessagingAttribute)
			return
		}

		phoneNumbers := extractRecipientNumbers(payload, tmpl, cfg, consentValue)
		if len(phoneNumbers) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"status":  "No recipients",
				"message": "No valid recipient phone numbers found",
			})
			return
		}

		message, err := craftMessage(payload, tmpl, cfg.Templates.LanguageAttribute, dueDate)
		if err != nil {
			respondWithError(c, http.StatusNotFound, err.Error())
			return
		}

		handleMessageSending(c, cfg, phoneNumbers, message, tmpl.ID, language(payload, cfg.Templates.LanguageAttribute))
	}
}

// Helper Functions Below

func parsePayload(c *gin.Context) (map[string]interface{}, error) {
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func findTemplate(cfg *config.Config, templateID, trigger, daysOffsetStr string) (*config.ProgramNotificationTemplate, error) {
	if templateID != "" {
		tmpl := config.FindTemplateByID(cfg, templateID)
		if tmpl == nil {
			return nil, fmt.Errorf("Template not found")
		}
		return tmpl, nil
	}

	var daysOffset *int
	if daysOffsetStr != "" {
		if v, err := strconv.Atoi(daysOffsetStr); err == nil {
			daysOffset = &v
		}
	}
	for _, t := range cfg.Templates.ProgramNotificationTemplates {
		if (trigger == "" || strings.EqualFold(t.NotificationTrigger, trigger)) &&
			(daysOffset == nil || t.RelativeScheduledDays == *daysOffset) {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("Template not found by trigger/days_offset")
}

func computeDueDate(payload map[string]interface{}, tmpl *config.ProgramNotificationTemplate) string {
	if tmpl.NotificationTrigger != "SCHEDULED_DAYS_DUE_DATE" {
		return ""
	}
	currentDateStr, _ := payload["CURRENT_DATE"].(string)
	currentDate, _ := time.Parse("2006-01-02", currentDateStr)
	return currentDate.AddDate(0, 0, -tmpl.RelativeScheduledDays).Format("2006-01-02")
}

func extractConsentValue(payload map[string]interface{}, consentAttr string) string {
	if consentAttr != "" {
		if v, ok := payload[consentAttr]; ok && v != nil {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func isMessagingAllowed(payload map[string]interface{}, cfg *config.Config) bool {
	if cfg.Templates.AllowMessagingAttribute != "" {
		if v, ok := payload[cfg.Templates.AllowMessagingAttribute]; ok && v != nil {
			allowMessaging := fmt.Sprintf("%v", v)
			return !strings.EqualFold(allowMessaging, "false") && !strings.EqualFold(allowMessaging, "no") && allowMessaging != ""
		} else {
			// If the attribute is not present, we assume messaging is not allowed
			return false
		}
	}
	return true
}

func extractRecipientNumbers(
	payload map[string]interface{},
	tmpl *config.ProgramNotificationTemplate,
	cfg *config.Config,
	consentValue string,
) []string {
	recipientAttrs := utils.FilterRecipientAttributes(
		tmpl.RecipientAttributes,
		cfg.Templates.ConsentIgnoreAttributes,
		consentValue,
	)

	return utils.ExtractUniquePhones(payload, recipientAttrs, "UG")
}

func craftMessage(payload map[string]interface{}, tmpl *config.ProgramNotificationTemplate, langAttr, dueDate string) (string, error) {
	lang := utils.DetectLanguage(payload, langAttr)
	messageTemplate := tmpl.MessageTemplates[lang]
	if messageTemplate == "" {
		return "", fmt.Errorf("No message template found for requested language: %s", lang)
	}

	msgPayload := make(map[string]interface{})
	for k, v := range payload {
		msgPayload[k] = v
	}
	if dueDate != "" {
		msgPayload["due_date"] = dueDate
	}

	return config.SubstituteTemplate(messageTemplate, msgPayload), nil
}

func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

func respondWithStatus(c *gin.Context, status, templateID, ignoreAttr string) {
	c.JSON(http.StatusOK, gin.H{
		"status":           status,
		"template_id":      templateID,
		"ignore_attribute": ignoreAttr,
	})
}

func handleMessageSending(c *gin.Context, cfg *config.Config, phoneNumbers []string, message, templateID, lang string) {
	if cfg.Server.InTestMode {
		successful, failed := SendTestSMSorTelegram(phoneNumbers, message, config.AppConfig.Telegram.TelegramBots, config.AppConfig.Telegram.DefaultBot, sendTelegramMessage)
		c.JSON(http.StatusOK, gin.H{
			"status":      "Test mode: Sent via Telegram",
			"recipients":  successful,
			"failed":      failed,
			"template_id": templateID,
			"lang":        lang,
		})
		return
	}

	successfulSends, failedSends := SendBulkSMSWithSMSOne(phoneNumbers, message, config.AppConfig.SMSOne, smsOneClient, sendUsingSMSOne)
	c.JSON(http.StatusOK, gin.H{
		"status":      "Notification sent",
		"recipients":  successfulSends,
		"failed":      failedSends,
		"template_id": templateID,
		"lang":        lang,
	})
}

func language(payload map[string]interface{}, langAttr string) string {
	return utils.DetectLanguage(payload, langAttr)
}
