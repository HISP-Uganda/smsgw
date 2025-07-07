package controllers

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"smsgw/client"
	"smsgw/config"
)

var DHIS2Clients = make(map[string]*client.Client)

func init() {
	// look through AppConfig.DHIS2Instances and initialize clients for each instance
	for _, instance := range config.AppConfig.DHISInstances {
		dClient := client.NewClient(instance.BaseURL, instance.Username, instance.Password, instance.Pat)
		if dClient == nil {
			log.Errorf("Failed to initialize Dhis2 client for instance: %s", instance.BaseURL)
			continue
		}
		DHIS2Clients[instance.ID] = dClient
		log.Infof("Initialized Dhis2 client for instance: %s", instance.ID)
	}
}

type Dhis2Controller struct{}

func (d *Dhis2Controller) SendToDhis2Handler(c *gin.Context) {
	phoneNumber := c.Query("originator")
	message := c.Query("message")
	instanceID := c.Param("instance_id")
	if instanceClient, ok := DHIS2Clients[instanceID]; !ok {
		log.Errorf("Dhis2 client not found for instance ID: %s", instanceID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Dhis2 client not found for instance ID"})
		return
	} else {
		if message != "" && phoneNumber != "" {
			params := map[string]string{"originator": phoneNumber, "text": message}
			resp, err := instanceClient.PostResource("sms/inbound", params)
			if err != nil {
				log.Printf("Error sending SMS to Dhis2: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			} else {
				log.Printf("SMS sent successfully to Dhis2: %v", string(resp.Body()))
			}
		}
	}

}
