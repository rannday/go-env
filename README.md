# go-env
Go environment configuration utility.

## Usage
1. Define a struct with `env` tags.
2. Call `goenv.Load(&cfg, goenv.Options{DotEnvPath: ".env"})`.
3. Check the returned error.

```go
package main

import (
	"fmt"
	"time"

	"github.com/rannday/go-env"
)

type Config struct {
	Addr      string        `env:"APP_ADDR" default:":8080"`
	Timeout   time.Duration `env:"APP_TIMEOUT" default:"5s"`
	Features  []string      `env:"APP_FEATURES"`
	Published time.Time     `env:"APP_PUBLISHED_AT"`
}

func main() {
	var cfg Config
	if err := goenv.Load(&cfg, goenv.Options{DotEnvPath: ".env"}); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", cfg)
}
```

## Behavior
- Reads values from the process environment first.
- Falls back to an optional `.env` file if a variable is missing.
- Uses the struct tag `default` only when neither source provides a value.
- Leaves the process environment unchanged.
- Supports `default`, `required`, and `allowempty` tags.

## Supported Types
- Strings
- Booleans
- Signed integers
- Unsigned integers
- Floats
- `time.Duration`
- `time.Time`
- `[]string`, `[]int`, and other comma-separated slices of supported scalar types
- `[]byte`
- Types that implement `encoding.TextUnmarshaler`

## Tags
- `env:"NAME"` binds a field to an environment variable.
- `default:"value"` provides a fallback when the environment and `.env` file do not set a value.
- `required:"true"` causes `Load` to fail when no value is resolved.
- `allow_empty:"true"` allows an explicit empty string for required values.
- `allowempty:"true"` is still accepted for compatibility.

## Precedence
1. Process environment
2. Optional `.env` file
3. Struct tag `default`

## Notes
- `.env` parsing is intentionally small and strict enough to catch malformed lines early.
- String slices are comma-separated and whitespace around items is trimmed.
- Empty slice values resolve to an empty slice rather than a single empty item.
- `required:"true"` means a value must resolve from the process environment, `.env`, or `default`.
- Use `allow_empty:"true"` with `required:"true"` when an explicit empty string should be accepted.
