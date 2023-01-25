package cmd

import (
	"crypto/ecdsa"
	"os"
	"time"

	fb "github.com/alvidir/filebrowser"
	cert "github.com/alvidir/filebrowser/certificate"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	ENV_SERVICE_PORT            = "SERVICE_PORT"
	ENV_SERVICE_ADDR            = "SERVICE_ADDR"
	ENV_SERVICE_NETW            = "SERVICE_NETW"
	ENV_UID_HEADER              = "UID_HEADER"
	ENV_MONGO_DSN               = "MONGO_DSN"
	ENV_MONGO_DATABASE          = "MONGO_DATABASE"
	ENV_REDIS_DSN               = "REDIS_DSN"
	ENV_TOKEN_TIMEOUT           = "TOKEN_TIMEOUT"
	ENV_JWT_SECRET              = "JWT_SECRET"
	ENV_TOKEN_ISSUER            = "TOKEN_ISSUER"
	ENV_EVENT_ISSUER            = "EVENT_ISSUER"
	ENV_RABBITMQ_USERS_EXCHANGE = "RABBITMQ_USERS_EXCHANGE"
	ENV_RABBITMQ_USERS_QUEUE    = "RABBITMQ_USERS_QUEUE"
	ENV_RABBITMQ_FILES_EXCHANGE = "RABBITMQ_FILES_EXCHANGE"
	ENV_RABBITMQ_FILES_QUEUE    = "RABBITMQ_FILES_QUEUE"
	ENV_RABBITMQ_DSN            = "RABBITMQ_DSN"
)

var (
	ServicePort = "8000"
	ServiceAddr = "127.0.0.1"
	ServiceNetw = "tcp"
	UidHeader   = "X-Uid"
)

func GetMongoConnection(logger *zap.Logger) *mongo.Database {
	mongoUri, exists := os.LookupEnv(ENV_MONGO_DSN)
	if !exists {
		logger.Fatal("mongo dsn must be set")
	}

	database, exists := os.LookupEnv(ENV_MONGO_DATABASE)
	if !exists {
		logger.Fatal("mongo database name must be set")
	}

	mongoConn, err := fb.NewMongoDBConn(mongoUri, database)
	if err != nil {
		logger.Fatal("failed establishing connection",
			zap.String("uri", mongoUri),
			zap.Error(err))
	} else {
		logger.Info("connection with mongodb cluster established")
	}

	return mongoConn
}

func GetPrivateKey(logger *zap.Logger) *ecdsa.PrivateKey {
	secret, exists := os.LookupEnv(ENV_JWT_SECRET)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_JWT_SECRET))
	}

	privateKey, err := cert.ParsePKCS8PrivateKey(secret)
	if err != nil {
		logger.Fatal("parsing PKCS8 private key",
			zap.Error(err))
	}

	return privateKey
}

func GetTokenTTL(logger *zap.Logger) *time.Duration {
	value, exists := os.LookupEnv(ENV_TOKEN_TIMEOUT)
	if !exists {
		return nil
	}

	ttl, err := time.ParseDuration(value)
	if err != nil {
		logger.Fatal("invalid token ttl",
			zap.String("value", value),
			zap.Error(err))
	}

	return &ttl
}

func GetTokenIssuer(logger *zap.Logger) string {
	value, exists := os.LookupEnv(ENV_TOKEN_ISSUER)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_TOKEN_ISSUER))
	}

	return value
}

func GetEventIssuer(logger *zap.Logger) string {
	value, exists := os.LookupEnv(ENV_EVENT_ISSUER)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_EVENT_ISSUER))
	}

	return value
}

func GetAmqpConnection(logger *zap.Logger) *amqp.Connection {
	addr, exists := os.LookupEnv(ENV_RABBITMQ_DSN)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_RABBITMQ_DSN))
	}

	conn, err := amqp.Dial(addr)
	if err != nil {
		logger.Fatal("establishing connection",
			zap.String("addr", addr),
			zap.Error(err))
	}

	return conn
}

func GetAmqpChannel(conn *amqp.Connection, logger *zap.Logger) *amqp.Channel {
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("openning channel",
			zap.Error(err))
	}

	return ch
}

func GetFileExchange(logger *zap.Logger) string {
	value, exists := os.LookupEnv(ENV_RABBITMQ_FILES_EXCHANGE)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_RABBITMQ_FILES_EXCHANGE))
	}

	return value
}
