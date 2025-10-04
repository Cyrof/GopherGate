package envx

import (
	"os"
)

func IsDev() bool {
	return os.Getenv("GOPHERGATE_ENV") == "dev"
}
