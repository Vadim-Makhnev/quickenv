# quickenv

A minimal and reliable `.env` loader for Go.

No dependencies. No magic. Just loads environment variables from `.env` files — fast and safely.

## Features

- Loads `.env` from current directory or parent folders (configurable depth)
- Supports `export KEY=value`
- Handles `"double"` and `'single'` quoted values
- Removes surrounding quotes: `"value"` → `value`
- Skips empty lines and comments (`#`)
- Validates keys: must start with letter or `_`, rest: letters, digits, `_`
- Debug mode: log loaded and skipped lines
- Helper: `GetEnv(key, default)` and `GetEnvOrPanic(key)`

## Installation
```bash
go get github.com/Vadim-Makhnev/quickenv
```

## Usage
# Add your .env file in the root
```env
SECRET_KEY=qwerty123
DB_PASSWORD=password
```
# Load variables in your Go app
```go
package main

import (
    "log"
    "github.com/Vadim-Makhnev/quickenv"
)

func main() {
    count, err := quickenv.Load(&quickenv.LoadOptions{
        Overwrite: true,
        Debug:     true,
    })
    if err != nil {
        log.Fatal("Failed to load .env file")
    }
    log.Printf("Loaded %d environment variables", count)

    // Safely access values with fallbacks
    configPath := quickenv.GetEnv("CONFIG_PATH", "config/local.env")
    dbPort := quickenv.GetEnv("DB_PORT", "8000")

    log.Println("Config:", configPath)
    log.Println("DB Port:", dbPort)
}
```
# Or use MustLoad for fail-fast initialization
```go
func main() {
    count := quickenv.MustLoad() // panics on error
    log.Printf("Loaded %d variables", count)

    apiKey := quickenv.GetEnvOrPanic("API_KEY") // useful for required secrets
    log.Println("API Key:", apiKey)
}
```

