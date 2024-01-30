package config

import (
	"log"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conf struct {
	InitialAccount initialAccountConf
	AppServer      serverConf
	DB             dbConfig
	AWS            awsConfig
	Emailer        mailgunConfig
	PDFBuilder     pdfBuilderConfig
}

type initialAccountConf struct {
	AdminEmail              string
	AdminPassword           string
	AdminTenantID           primitive.ObjectID
	AdminTenantName         string
	AdminTenantOpenAIAPIKey string
	AdminTenantOpenAIOrgKey string
}

type serverConf struct {
	Port         string
	IP           string
	HMACSecret   []byte
	HasDebugging bool
	DomainName   string
}

type dbConfig struct {
	URI  string
	Name string
}

type awsConfig struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string
	Region     string
	BucketName string
}

type mailgunConfig struct {
	APIKey      string
	Domain      string
	APIBase     string
	SenderEmail string
}

type pdfBuilderConfig struct {
	AssociateInvoiceTemplatePath string
	DataDirectoryPath            string
}

func New() *Conf {
	var c Conf
	c.InitialAccount.AdminEmail = getEnv("DATABOUTIQUE_BACKEND_INITIAL_ADMIN_EMAIL", true)
	c.InitialAccount.AdminPassword = getEnv("DATABOUTIQUE_BACKEND_INITIAL_ADMIN_PASSWORD", true)
	c.InitialAccount.AdminTenantID = getObjectIDEnv("DATABOUTIQUE_BACKEND_INITIAL_ADMIN_STORE_ID", true)
	c.InitialAccount.AdminTenantName = getEnv("DATABOUTIQUE_BACKEND_INITIAL_ADMIN_STORE_NAME", true)
	c.InitialAccount.AdminTenantOpenAIAPIKey = getEnv("DATABOUTIQUE_BACKEND_INITIAL_ADMIN_STORE_OPENAI_KEY", true)
	c.InitialAccount.AdminTenantOpenAIOrgKey = getEnv("DATABOUTIQUE_BACKEND_INITIAL_ADMIN_STORE_OPENAI_ORGANIZATION_KEY", true)

	c.AppServer.Port = getEnv("DATABOUTIQUE_BACKEND_PORT", true)
	c.AppServer.IP = getEnv("DATABOUTIQUE_BACKEND_IP", false)
	c.AppServer.HMACSecret = []byte(getEnv("DATABOUTIQUE_BACKEND_HMAC_SECRET", true))
	c.AppServer.HasDebugging = getEnvBool("DATABOUTIQUE_BACKEND_HAS_DEBUGGING", true, true)
	c.AppServer.DomainName = getEnv("DATABOUTIQUE_BACKEND_DOMAIN_NAME", true)

	c.DB.URI = getEnv("DATABOUTIQUE_BACKEND_DB_URI", true)
	c.DB.Name = getEnv("DATABOUTIQUE_BACKEND_DB_NAME", true)

	c.AWS.AccessKey = getEnv("DATABOUTIQUE_BACKEND_AWS_ACCESS_KEY", true)
	c.AWS.SecretKey = getEnv("DATABOUTIQUE_BACKEND_AWS_SECRET_KEY", true)
	c.AWS.Endpoint = getEnv("DATABOUTIQUE_BACKEND_AWS_ENDPOINT", true)
	c.AWS.Region = getEnv("DATABOUTIQUE_BACKEND_AWS_REGION", true)
	c.AWS.BucketName = getEnv("DATABOUTIQUE_BACKEND_AWS_BUCKET_NAME", true)

	c.Emailer.APIKey = getEnv("DATABOUTIQUE_BACKEND_MAILGUN_API_KEY", true)
	c.Emailer.Domain = getEnv("DATABOUTIQUE_BACKEND_MAILGUN_DOMAIN", true)
	c.Emailer.APIBase = getEnv("DATABOUTIQUE_BACKEND_MAILGUN_API_BASE", true)
	c.Emailer.SenderEmail = getEnv("DATABOUTIQUE_BACKEND_MAILGUN_SENDER_EMAIL", true)

	c.PDFBuilder.DataDirectoryPath = getEnv("DATABOUTIQUE_BACKEND_PDF_BUILDER_DATA_DIRECTORY_PATH", true)
	c.PDFBuilder.AssociateInvoiceTemplatePath = getEnv("DATABOUTIQUE_BACKEND_PDF_BUILDER_ASSOCIATE_INVOICE_PATH", true)

	return &c
}

func getEnv(key string, required bool) string {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	return value
}

func getEnvBool(key string, required bool, defaultValue bool) bool {
	valueStr := getEnv(key, required)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Fatalf("Invalid boolean value for environment variable %s", key)
	}
	return value
}

func getObjectIDEnv(key string, required bool) primitive.ObjectID {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	objectID, err := primitive.ObjectIDFromHex(value)
	if err != nil {
		log.Fatalf("Invalid mongodb primitive object id value for environment variable %s", key)
	}
	return objectID
}
