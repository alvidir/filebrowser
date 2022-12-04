package filebrowser

import (
	"os"
	"testing"

	"log"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Println("loading dotenv file")
	}

	code := m.Run()
	os.Exit(code)
}
