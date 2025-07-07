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

		// Extract recipient numbers using RecipientAttributes
		phoneNumbers := utils.ExtractUniquePhones(payload, tmpl.RecipientAttributes, "UG")
		if len(phoneNumbers) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid recipient phone numbers"})
			return
		}

		langAttr := cfg.Templates.LanguageAttribute
		lang := "en"
		if langAttr != "" {
			if l, ok := payload[langAttr]; ok && l != nil && fmt.Sprintf("%v", l) != "" {
				lang = fmt.Sprintf("%v", l)
			}
		}
		messageTemplate := tmpl.MessageTemplates[lang]
		if messageTemplate == "" {
			// fallback to first message template if not found
			for _, v := range tmpl.MessageTemplates {
				messageTemplate = v
				break
			}
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
		log.Printf("Sending SMS to %v: %s", phoneNumbers, message)

		c.JSON(http.StatusOK, gin.H{
			"status":      "Notification sent",
			"recipients":  phoneNumbers,
			"template_id": tmpl.ID,
		})
	}
}
