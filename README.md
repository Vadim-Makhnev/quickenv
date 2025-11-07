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
