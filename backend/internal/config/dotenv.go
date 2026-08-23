package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// dotEnvCandidates are searched in order; the first existing file wins. The
// canonical location is the repository root, which is also where Docker
// Compose looks, so `docker compose` and `go run ./cmd/api` read one file.
// Running from backend/ finds it via the parent path.
var dotEnvCandidates = []string{".env", filepath.Join("..", ".env")}

// loadDotEnv loads a .env file when one exists. Real environment variables
// always win: godotenv.Load never overwrites what is already set.
//
// A missing file is not an error — production sets real environment variables
// and ships no .env at all.
func loadDotEnv() {
	if os.Getenv("APP_ENV") == EnvProduction {
		return
	}

	for _, path := range dotEnvCandidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		// Errors are deliberately ignored: a malformed .env must not stop the
		// service when real environment variables may already be sufficient.
		_ = godotenv.Load(path)
		return
	}
}
