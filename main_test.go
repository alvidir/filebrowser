package filebrowser

import (
	"os"
	"testing"

	"log"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Println("no dotenv file has been found")
	}

	code := m.Run()
	os.Exit(code)
}
