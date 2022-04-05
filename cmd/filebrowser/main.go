package main

import (
	"os"

	"github.com/alvidir/filebrowser"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	ENV_MONGO_DSN      = "MONGO_DSN"
	ENV_MONGO_DATABASE = "MONGO_INITDB_DATABASE"
)

var (
	logger = log.StandardLogger()
)

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Warn("no dotenv file has been found: ", err)
	}

	uri, exists := os.LookupEnv(ENV_MONGO_DSN)
	if !exists {
		logger.Error("mongo dsn must be set")
		return
	}

	database, exists := os.LookupEnv(ENV_MONGO_DATABASE)
	if !exists {
		logger.Error("mongo database name must be set")
		return
	}

	if _, err := filebrowser.NewMongoDBConn(uri, database); err != nil {
		logger.Errorf("establishing connection with %s: %s", uri, err)
		return
	} else {
		logger.Info("connection with mongo cluster established")
	}
}
