package envx

import (
	"github.com/joho/godotenv"
)

func LoadDotenvIfPresent() {
	_ = godotenv.Load()
}
