package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"smsgw/config"
	"smsgw/controllers"
	"smsgw/db"
	"time"
)

var splash = `
┏━┓┏┳┓┏━┓   ┏━╸┏━┓╺┳╸┏━╸╻ ╻┏━┓╻ ╻
┗━┓┃┃┃┗━┓   ┃╺┓┣━┫ ┃ ┣╸ ┃╻┃┣━┫┗┳┛
┗━┛╹ ╹┗━┛   ┗━┛╹ ╹ ╹ ┗━╸┗┻┛╹ ╹ ╹
`

func init() {
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = time.RFC3339
	formatter.FullTimestamp = true

	if config.AppConfig.Server.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
}

func main() {
	fmt.Printf(splash)
	err := db.RunMigrations()
	if err != nil {
		fmt.Printf("Error running migrations: %v\n", err)
		return
	}
	router := gin.Default()

	smsOneController := &controllers.SMSOneController{}
	router.GET("/sms", smsOneController.SMSOneHandler)

	telegramController := &controllers.TelegramController{}
	router.GET("/telegram", telegramController.SendSMS)

	sendSMSController := &controllers.SendSMSController{}
	router.POST("/sendsms", sendSMSController.SendSMSHandler)

	dhis2Controller := &controllers.Dhis2Controller{}
	router.GET("/sendToDhis2/:instance_id", dhis2Controller.SendToDhis2Handler)

	testController := &controllers.TestController{}
	router.POST("/test", testController.TestHandler)

	router.GET("/test", func(c *gin.Context) {
		fmt.Println("Test Get endpoint hit")
		c.JSON(http.StatusOK, gin.H{"message": "Hello, this is a test"})
	})

	notificationControlle := &controllers.NotificationController{}
	router.POST("/notification", notificationControlle.NotificationHandler(&config.AppConfig))

	router.Run(fmt.Sprintf(":%d", config.AppConfig.Server.Port))
}
