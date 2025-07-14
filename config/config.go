package config

import (
	goflag "flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"regexp"
	"runtime"
	"strings"
)

type TelegramBot struct {
	ChatID      int64  `mapstructure:"chatID"`
	Token       string `mapstructure:"token"`
	Description string `mapstructure:"description"`
}

type ProgramNotificationTemplate struct {
	ID                    string
	NotificationTrigger   string            `json:"notificationTrigger"` // "ENROLLMENT" or "SCHEDULED_DAYS_DUE_DATE"
	RelativeScheduledDays int               `json:"relativeScheduledDays"`
	MessageTemplates      map[string]string `json:"messageTemplates"`
	RecipientAttributes   []string          `json:"recipientAttributes"`
}

type DHISInstance struct {
	ID         string `mapstructure:"id"`
	BaseURL    string `mapstructure:"baseURL"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	Pat        string `mapstructure:"pat"` // Personal Access Token
	AuthMethod string `mapstructure:"authMethod"`
}

type SMSOneConfig struct {
	SMSOneBaseURL     string `mapstructure:"smsone_baseurl" env:"SMSONE_BASEURL" env-description:"SMSOne Base URL"`
	SMSOneApiID       string `mapstructure:"smsone_api_id" env:"SMSONE_APIID" env-description:"SMSOne API ID"`
	SMSOneAPIPassword string `mapstructure:"smsone_api_password" env:"SMSONE_APIPASSWORD" env-description:"SMSOne API Password"`
	SMSOneSenderID    string `mapstructure:"smsone_sender_id" env:"SMSONE_SENDERID" env-description:"SMSOne Sender ID"`
	SMSOneSmsType     string `mapstructure:"smsone_sms_type" env:"SMSONE_SMSTYPE" env-description:"SMSOne SMS Type"`
	SMSOneEncoding    string `mapstructure:"smsone_encoding" env:"SMSONE_ENCODING" env-description:"SMSOne Encoding"`
}

type Config struct {
	Database struct {
		URI string `mapstructure:"uri" env:"DATABASE_URI" env-description:"Database URI"`
	} `yaml:"database"`
	Server struct {
		Port                int    `mapstructure:"port" env:"PORT" env-description:"Server Port"`
		MigrationsDirectory string `mapstructure:"migrations_directory" env:"MIGRATIONS_DIRECTORY" env-description:"Directory for database migrations"`
		InTestMode          bool   `mapstructure:"in_test_mode" env:"IN_TEST_MODE" env-description:"Run in test mode, disables certain features"`
	} `yaml:"server"`
	DHISInstances map[string]DHISInstance `mapstructure:"dhis_instances" env:"DHIS_INSTANCES" env-description:"DHIS2 Instances Configuration"`
	SMSOne        SMSOneConfig            `yaml:"smsone"`
	Telegram      struct {
		DefaultBot   TelegramBot            `mapstructure:"default_bot" env:"DEFAULT_BOT" env-description:"Default Telegram Bot Configuration"`
		TelegramBots map[string]TelegramBot `mapstructure:"telegram_bots" env:"TELEGRAM_BOTS" env-description:"Telegram Bots Configuration"`
	} `yaml:"telegram"`
	Templates struct {
		LanguageAttribute       string   `mapstructure:"language_attribute" env:"LANGUAGE_ATTRIBUTE" env-description:"Language Attribute for Templates"`
		AllowMessagingAttribute string   `mapstructure:"allow_messaging_attribute" env:"ALLOW_MESSAGING_ATTRIBUTE" env-description:"Attribute to allow messaging"`
		ConsentAttribute        string   `mapstructure:"consent_attribute" env:"CONSENT_ATTRIBUTE" env-description:"Consent Attribute for Messaging Next of Kin"`
		ConsentIgnoreAttributes []string `mapstructure:"consent_ignore_attributes" env:"CONSENT_IGNORE_ATTRIBUTES" env-description:"Attributes to ignore for consent checks"`

		ProgramNotificationTemplates []ProgramNotificationTemplate `mapstructure:"program_notification_templates" env:"TEMPLATES" env-description:"Program Notification Templates"`
	} `yaml:"templates"`
}

var AppConfig Config

func init() {
	var configFilePath string
	currentOS := runtime.GOOS
	switch currentOS {
	case "windows":
		configFilePath = "C:\\ProgramData\\smsgw\\smsgw.yml"
	case "darwin", "linux":
		configFilePath = "/etc/smsgw/smsgw.yml"
	default:
		fmt.Println("Unsupported operating system")
		return
	}

	configFile := flag.String("config-file", configFilePath,
		"The path to the configuration file of the application")

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	viper.SetConfigName("smsgw")
	viper.SetConfigType("yaml")

	if len(*configFile) > 0 {
		viper.SetConfigFile(*configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// log.Fatalf("Configuration File: %v Not Found", *configFile, err)
			panic(fmt.Errorf("Fatal Error %w \n", err))

		} else {
			log.Fatalf("Error Reading Config: %v", err)

		}
	}

	AppConfig.SMSOne.SMSOneApiID = "API123749713838"
	AppConfig.SMSOne.SMSOneAPIPassword = "Hi@2489XUg"
	AppConfig.SMSOne.SMSOneBaseURL = "http://apidocs.speedamobile.com/"
	AppConfig.SMSOne.SMSOneSenderID = "ANTENATAL"
	AppConfig.SMSOne.SMSOneSmsType = "P"
	AppConfig.SMSOne.SMSOneEncoding = "T"
	AppConfig.Server.InTestMode = false
	AppConfig.Server.Port = 8383
	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		err = viper.ReadInConfig()
		if err != nil {
			log.Fatalf("unable to reread configuration into global conf: %v", err)
		}
		_ = viper.Unmarshal(&AppConfig)
	})
	viper.WatchConfig()
}

// FindTemplateByID searches for a template by its ID in the given configuration
func FindTemplateByID(cfg *Config, id string) *ProgramNotificationTemplate {
	for _, t := range cfg.Templates.ProgramNotificationTemplates {
		if strings.EqualFold(t.ID, id) {
			return &t
		}
	}
	return nil
}

// SubstituteTemplate replaces A{attribute} and V{variable} placeholders
func SubstituteTemplate(template string, payload map[string]interface{}) string {
	// A{key} for attributes
	reAttr := regexp.MustCompile(`A\{([^\}]+)\}`)
	// V{key} for variables
	reVar := regexp.MustCompile(`V\{([^\}]+)\}`)

	// Replace attributes
	out := reAttr.ReplaceAllStringFunc(template, func(match string) string {
		key := reAttr.FindStringSubmatch(match)[1]
		if val, ok := payload[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return ""
	})

	// Replace variables (case-insensitive)
	out = reVar.ReplaceAllStringFunc(out, func(match string) string {
		key := reVar.FindStringSubmatch(match)[1]
		// Try exact key
		if val, ok := payload[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		// Try lower-case match
		for k, v := range payload {
			if strings.EqualFold(k, key) {
				return fmt.Sprintf("%v", v)
			}
		}
		return ""
	})

	return out
}
