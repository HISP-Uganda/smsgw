package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"smsgw/config"
	"smsgw/controllers"
	"smsgw/db"
)

var splash = `
┏━┓┏┳┓┏━┓   ┏━╸┏━┓╺┳╸┏━╸╻ ╻┏━┓╻ ╻
┗━┓┃┃┃┗━┓   ┃╺┓┣━┫ ┃ ┣╸ ┃╻┃┣━┫┗┳┛
┗━┛╹ ╹┗━┛   ┗━┛╹ ╹ ╹ ┗━╸┗┻┛╹ ╹ ╹
`

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

	router.Run(":8383") // Listen and serve on 0.0.0.0:8080
}
