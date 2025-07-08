package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
	"smsgw/client"
	"smsgw/config"
	"smsgw/utils"
	"strings"
)

type SMSOneRequest struct {
	ApiID       string `json:"api_id"`
	ApiPassword string `json:"api_password"`
	SmsType     string `json:"sms_type"`
	SenderID    string `json:"sender_id"`
	Encoding    string `json:"encoding"`
	PhoneNumber string `json:"phonenumber"`
	TextMessage string `json:"textmessage"`
	TemplateID  string `json:"template_id,omitempty"`
	V1          string `json:"V1,omitempty"`
	V2          string `json:"V2,omitempty"`
	V3          string `json:"V3,omitempty"`
	V4          string `json:"V4,omitempty"`
	V5          string `json:"V5,omitempty"`
}

// SendBulkSMSWithSMSOne sends a message to multiple numbers using SMSOne.
// Returns slices of successful and failed numbers.
func SendBulkSMSWithSMSOne(
	numbers []string,
	message string,
	cfg config.SMSOneConfig, // struct with ApiID, APIPassword, etc.
	smsOneClient *client.Client, // your HTTP client
	sendFunc func(SMSOneRequest, *client.Client) (*resty.Response, error),
) (successful, failed []string) {
	for _, number := range numbers {
		request := SMSOneRequest{
			ApiID:       cfg.SMSOneApiID,
			ApiPassword: cfg.SMSOneAPIPassword,
			SmsType:     cfg.SMSOneSmsType,
			SenderID:    cfg.SMSOneSenderID,
			Encoding:    cfg.SMSOneEncoding,
			PhoneNumber: number,
			TextMessage: message,
		}
		resp, err := sendFunc(request, smsOneClient)
		if err != nil {
			log.Printf("Failed to send SMS to %s: %v", number, err)
			failed = append(failed, number)
		} else {
			log.Printf("Successfully sent SMS to %s: %v", number, string(resp.Body()))
			successful = append(successful, number)
		}
	}
	return successful, failed
}

type SMSOneController struct{}

var smsOneClient = client.NewClient(config.AppConfig.SMSOne.SMSOneBaseURL, "", "", "")

func (s *SMSOneController) SMSOneHandler(c *gin.Context) {
	phoneNumbers := c.Query("to")
	message := c.Query("text")

	if phoneNumbers == "" || message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both phoneNumber and message are required"})
		return
	}

	numbers := strings.Split(phoneNumbers, ",")
	uniqueNumbers := utils.RemoveDuplicates(numbers)
	log.Printf("Numbers: %v", uniqueNumbers)

	// create a slice to keep track of failed sends
	//failedSends := make([]string, 0)
	//successfulSends := make([]string, 0)
	//for _, number := range uniqueNumbers {
	//	request := SMSOneRequest{
	//		ApiID:       config.AppConfig.SMSOne.SMSOneApiID,
	//		ApiPassword: config.AppConfig.SMSOne.SMSOneAPIPassword,
	//		SmsType:     config.AppConfig.SMSOne.SMSOneSmsType,
	//		SenderID:    config.AppConfig.SMSOne.SMSOneSenderID,
	//		Encoding:    config.AppConfig.SMSOne.SMSOneEncoding,
	//		PhoneNumber: number,
	//		TextMessage: message,
	//	}
	//	resp, err := sendUsingSMSOne(request, smsOneClient)
	//	if err != nil {
	//		failedSends = append(failedSends, number)
	//		log.Printf("Failed to send SMS to %s: %v", number, err)
	//		// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	} else {
	//		log.Printf("Successfully sent SMS to %s: %v", number, string(resp.Body()))
	//		successfulSends = append(successfulSends, number)
	//		// c.String(http.StatusOK, string(resp.Body()))
	//	}
	//}
	successfulSends, failedSends := SendBulkSMSWithSMSOne(
		uniqueNumbers,
		message,
		config.AppConfig.SMSOne,
		smsOneClient,
		sendUsingSMSOne, // your function: func(SMSOneRequest, *resty.Client) (*resty.Response, error)
	)
	c.JSON(http.StatusOK, gin.H{
		"failed_sends":     failedSends,
		"successful_sends": successfulSends,
	})

}

func sendUsingSMSOne(payload SMSOneRequest, c *client.Client) (*resty.Response, error) {
	return c.PostResource("/api/SendSMS", payload)
}
