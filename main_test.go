package filebrowser

import (
	"log"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := godotenv.Load(); err != nil {
		log.Println("no dotenv file has been found.", err)
	}
}
