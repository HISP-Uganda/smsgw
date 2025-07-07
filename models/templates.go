package models

//import (
//	"fmt"
//	"strings"
//	"time"
//)
//

//// GetDueDateForTemplate returns the due date for a given template based on the current date and relative scheduled days
//func GetDueDateForTemplate(
//	tmpl ProgramNotificationTemplate,
//	currentDate time.Time,
//) (string, error) {
//	if tmpl.NotificationTrigger == "SCHEDULED_DAYS_DUE_DATE" {
//		dueDate := currentDate.AddDate(0, 0, -tmpl.RelativeScheduledDays)
//		return dueDate.Format("2006-01-02"), nil
//	} else if tmpl.NotificationTrigger == "ENROLLMENT" {
//		// For ENROLLMENT, we don't calculate a due date here
//		return "", nil
//	}
//	return "", fmt.Errorf("unknown notification trigger: %s", tmpl.NotificationTrigger)
//}
//
//func GetMatchingTemplates(
//	templates []ProgramNotificationTemplate,
//	payload map[string]interface{},
//) []struct {
//	Template ProgramNotificationTemplate
//	DueDate  string // in "YYYY-MM-DD", only for SCHEDULED_DAYS_DUE_DATE
//} {
//	currentDateStr, ok := payload["CURRENT_DATE"].(string)
//	if !ok {
//		return nil
//	}
//	currentDate, err := time.Parse("2006-01-02", currentDateStr)
//	if err != nil {
//		return nil
//	}
//
//	var result []struct {
//		Template ProgramNotificationTemplate
//		DueDate  string
//	}
//	for _, tmpl := range templates {
//		if tmpl.NotificationTrigger == "SCHEDULED_DAYS_DUE_DATE" {
//			// DUE_DATE = CURRENT_DATE - relativeScheduledDays
//			dueDate := currentDate.AddDate(0, 0, -tmpl.RelativeScheduledDays)
//			result = append(result, struct {
//				Template ProgramNotificationTemplate
//				DueDate  string
//			}{
//				Template: tmpl,
//				DueDate:  dueDate.Format("2006-01-02"),
//			})
//		} else if tmpl.NotificationTrigger == "ENROLLMENT" {
//			enrollDateStr, ok := payload["ENROLLMENT_DATE"].(string)
//			if !ok || enrollDateStr == "" {
//				continue
//			}
//			enrollDate, err := time.Parse("2006-01-02", enrollDateStr)
//			if err != nil {
//				continue
//			}
//			daysSinceEnroll := int(currentDate.Sub(enrollDate).Hours() / 24)
//			if daysSinceEnroll == tmpl.RelativeScheduledDays {
//				result = append(result, struct {
//					Template ProgramNotificationTemplate
//					DueDate  string
//				}{
//					Template: tmpl,
//					DueDate:  "", // Not needed
//				})
//			}
//		}
//	}
//	return result
//}
//
//func GetMatchingTemplates2(
//	templates []ProgramNotificationTemplate,
//	payload map[string]interface{},
//	templateID, trigger string,
//	daysOffset *int, // pointer, so nil if not provided
//) []struct {
//	Template ProgramNotificationTemplate
//	DueDate  string
//} {
//	var matches []struct {
//		Template ProgramNotificationTemplate
//		DueDate  string
//	}
//
//	currentDateStr, ok := payload["CURRENT_DATE"].(string)
//	if !ok {
//		return nil
//	}
//	currentDate, err := time.Parse("2006-01-02", currentDateStr)
//	if err != nil {
//		return nil
//	}
//
//	// First, if templateID is given, just get that one (if present)
//	if templateID != "" {
//		for _, tmpl := range templates {
//			if strings.EqualFold(tmpl.ID, templateID) {
//				switch tmpl.NotificationTrigger {
//				case "SCHEDULED_DAYS_DUE_DATE":
//					dueDate := currentDate.AddDate(0, 0, -tmpl.RelativeScheduledDays)
//					matches = append(matches, struct {
//						Template ProgramNotificationTemplate
//						DueDate  string
//					}{
//						Template: tmpl,
//						DueDate:  dueDate.Format("2006-01-02"),
//					})
//				case "ENROLLMENT":
//					enrollDateStr, ok := payload["ENROLLMENT_DATE"].(string)
//					if !ok || enrollDateStr == "" {
//						continue
//					}
//					enrollDate, err := time.Parse("2006-01-02", enrollDateStr)
//					if err != nil {
//						continue
//					}
//					daysSinceEnroll := int(currentDate.Sub(enrollDate).Hours() / 24)
//					if daysSinceEnroll == tmpl.RelativeScheduledDays {
//						matches = append(matches, struct {
//							Template ProgramNotificationTemplate
//							DueDate  string
//						}{
//							Template: tmpl,
//							DueDate:  "",
//						})
//					}
//				}
//				// Return immediately if ID matched
//				return matches
//			}
//		}
//		// No matching template ID
//		return nil
//	}
//
//	// Otherwise, filter by trigger/days_offset if given
//	for _, tmpl := range templates {
//		if trigger != "" && !strings.EqualFold(tmpl.NotificationTrigger, trigger) {
//			continue
//		}
//		if daysOffset != nil && tmpl.RelativeScheduledDays != *daysOffset {
//			continue
//		}
//		switch tmpl.NotificationTrigger {
//		case "SCHEDULED_DAYS_DUE_DATE":
//			dueDate := currentDate.AddDate(0, 0, -tmpl.RelativeScheduledDays)
//			matches = append(matches, struct {
//				Template ProgramNotificationTemplate
//				DueDate  string
//			}{
//				Template: tmpl,
//				DueDate:  dueDate.Format("2006-01-02"),
//			})
//		case "ENROLLMENT":
//			enrollDateStr, ok := payload["ENROLLMENT_DATE"].(string)
//			if !ok || enrollDateStr == "" {
//				continue
//			}
//			enrollDate, err := time.Parse("2006-01-02", enrollDateStr)
//			if err != nil {
//				continue
//			}
//			daysSinceEnroll := int(currentDate.Sub(enrollDate).Hours() / 24)
//			if daysSinceEnroll == tmpl.RelativeScheduledDays {
//				matches = append(matches, struct {
//					Template ProgramNotificationTemplate
//					DueDate  string
//				}{
//					Template: tmpl,
//					DueDate:  "",
//				})
//			}
//		}
//	}
//	return matches
//}
//
//
//
//// TemplateIsAllowedToSend checks business rules for notification delivery
//func TemplateIsAllowedToSend(
//	tmpl ProgramNotificationTemplate,
//	payload map[string]interface{},
//	dueDate string,
//) bool {
//	currentDateStr, ok := payload["CURRENT_DATE"].(string)
//	if !ok || currentDateStr == "" {
//		return false
//	}
//	currentDate, err := time.Parse("2006-01-02", currentDateStr)
//	if err != nil {
//		return false
//	}
//
//	switch tmpl.NotificationTrigger {
//	// SCHEDULED_DAYS_ENROLLMENT_DATE, SCHEDULED_DAYS_DUE_DATE, ENROLLMENT, COMPLETION, PROGRAM_RULE
//	case "SCHEDULED_DAYS_DUE_DATE":
//		// DUE_DATE must not be before INCIDENT_DATE; CURRENT_DATE must not be before INCIDENT_DATE
//		incidentDateStr, ok := payload["INCIDENT_DATE"].(string)
//		if !ok || incidentDateStr == "" {
//			return false
//		}
//		incidentDate, err := time.Parse("2006-01-02", incidentDateStr)
//		if err != nil {
//			return false
//		}
//		dueDateTime, err := time.Parse("2006-01-02", dueDate)
//		if err != nil {
//			return false
//		}
//		if dueDateTime.Before(incidentDate) {
//			return false
//		}
//		if currentDate.Before(incidentDate) {
//			return false
//		}
//		return true
//
//	case "ENROLLMENT":
//		enrollDateStr, ok := payload["ENROLLMENT_DATE"].(string)
//		if !ok || enrollDateStr == "" {
//			return false
//		}
//		enrollDate, err := time.Parse("2006-01-02", enrollDateStr)
//		if err != nil {
//			return false
//		}
//		if enrollDate.After(currentDate) {
//			return false
//		}
//		return true
//	}
//
//	return true
//}
