package goenv

import (
	"os"
	"strings"
	"testing"
	"time"
)

type testConfig struct {
	Port  int    `env:"ENVX_TEST_PORT" default:"8080"`
	Host  string `env:"ENVX_TEST_HOST" required:"true"`
	Debug bool   `env:"ENVX_TEST_DEBUG" default:"false"`
}

func TestLoad_Success(t *testing.T) {
	t.Setenv("ENVX_TEST_HOST", "localhost")
	t.Setenv("ENVX_TEST_PORT", "9090")

	var cfg testConfig
	if err := Load(&cfg, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 9090 {
		t.Fatalf("expected Port=9090, got %d", cfg.Port)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("expected Host=localhost, got %s", cfg.Host)
	}

	if cfg.Debug {
		t.Fatalf("expected Debug=false, got %v", cfg.Debug)
	}
}

func TestLoad_DefaultValue(t *testing.T) {
	t.Setenv("ENVX_TEST_HOST", "localhost")

	var cfg testConfig
	if err := Load(&cfg, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Fatalf("expected default Port=8080, got %d", cfg.Port)
	}
}

func TestLoad_RequiredMissing(t *testing.T) {
	type cfg struct {
		Host string `env:"ENVX_TEST_REQUIRED_MISSING" required:"true"`
	}

	var c cfg
	if err := Load(&c, Options{}); err == nil {
		t.Fatal("expected error for missing required field")
	}
}

func TestLoad_RequiredRejectsExplicitEmpty(t *testing.T) {
	type cfg struct {
		Value string `env:"ENVX_TEST_REQUIRED_EMPTY" required:"true"`
	}

	t.Setenv("ENVX_TEST_REQUIRED_EMPTY", "")

	var c cfg
	if err := Load(&c, Options{}); err == nil {
		t.Fatal("expected error for empty required value")
	}
}

func TestLoad_RequiredAllowsExplicitEmptyWhenAllowEmpty(t *testing.T) {
	type cfg struct {
		Value string `env:"ENVX_TEST_REQUIRED_ALLOW_EMPTY" required:"true" allow_empty:"true"`
	}

	t.Setenv("ENVX_TEST_REQUIRED_ALLOW_EMPTY", "")

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Value != "" {
		t.Fatalf("expected empty string, got %q", c.Value)
	}
}

func TestLoad_RequiredAllowsExplicitEmptyWhenLegacyAllowEmptyTag(t *testing.T) {
	type cfg struct {
		Value string `env:"ENVX_TEST_REQUIRED_ALLOW_EMPTY_LEGACY" required:"true" allowempty:"true"`
	}

	t.Setenv("ENVX_TEST_REQUIRED_ALLOW_EMPTY_LEGACY", "")

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Value != "" {
		t.Fatalf("expected empty string, got %q", c.Value)
	}
}

func TestLoad_InvalidInt(t *testing.T) {
	t.Setenv("ENVX_TEST_HOST", "localhost")
	t.Setenv("ENVX_TEST_PORT", "notanumber")

	var cfg testConfig
	if err := Load(&cfg, Options{}); err == nil {
		t.Fatal("expected error for invalid int")
	}
}

func TestLoad_NotPointer(t *testing.T) {
	var cfg testConfig
	if err := Load(cfg, Options{}); err == nil {
		t.Fatal("expected error when passing non-pointer")
	}
}

func TestLoad_NilPointer(t *testing.T) {
	var cfg *testConfig
	if err := Load(cfg, Options{}); err == nil {
		t.Fatal("expected error when passing nil pointer")
	}
}

func TestLoad_Int64(t *testing.T) {
	type cfg struct {
		Timeout int64 `env:"ENVX_TEST_TIMEOUT_I64" default:"60"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Timeout != 60 {
		t.Fatalf("expected 60, got %d", c.Timeout)
	}
}

func TestLoad_Duration(t *testing.T) {
	type cfg struct {
		Timeout time.Duration `env:"ENVX_TEST_TIMEOUT_DURATION" default:"5s"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Timeout != 5*time.Second {
		t.Fatalf("expected 5s, got %v", c.Timeout)
	}
}

func TestLoad_StringSlice(t *testing.T) {
	type cfg struct {
		Hosts []string `env:"ENVX_TEST_HOSTS" default:"a.com,b.com"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(c.Hosts))
	}

	if c.Hosts[0] != "a.com" || c.Hosts[1] != "b.com" {
		t.Fatalf("unexpected slice values: %v", c.Hosts)
	}
}

func TestLoad_IntSlice(t *testing.T) {
	type cfg struct {
		Ports []int `env:"ENVX_TEST_PORTS" default:"80,443"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Ports) != 2 || c.Ports[0] != 80 || c.Ports[1] != 443 {
		t.Fatalf("unexpected ports: %v", c.Ports)
	}
}

func TestLoad_EmptySliceValueYieldsEmptySlice(t *testing.T) {
	type cfg struct {
		List []string `env:"ENVX_TEST_EMPTY_SLICE" default:""`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.List == nil {
		t.Fatal("expected non-nil empty slice")
	}

	if len(c.List) != 0 {
		t.Fatalf("expected empty slice, got %v", c.List)
	}
}

func TestLoad_OptionalEmptyInt(t *testing.T) {
	type cfg struct {
		Port int `env:"ENVX_TEST_OPTIONAL_INT"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Port != 0 {
		t.Fatalf("expected zero value, got %d", c.Port)
	}
}

func TestLoad_TimeValue(t *testing.T) {
	type cfg struct {
		CreatedAt time.Time `env:"ENVX_TEST_TIME" default:"2026-05-15T12:34:56Z"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.CreatedAt.Format(time.RFC3339) != "2026-05-15T12:34:56Z" {
		t.Fatalf("unexpected time value: %s", c.CreatedAt.Format(time.RFC3339))
	}
}

func TestLoad_FloatValue(t *testing.T) {
	type cfg struct {
		Ratio float64 `env:"ENVX_TEST_FLOAT" default:"1.25"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Ratio != 1.25 {
		t.Fatalf("expected 1.25, got %v", c.Ratio)
	}
}

func TestLoad_UintValue(t *testing.T) {
	type cfg struct {
		Count uint `env:"ENVX_TEST_UINT" default:"42"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Count != 42 {
		t.Fatalf("expected 42, got %d", c.Count)
	}
}

func TestLoad_ByteSlice(t *testing.T) {
	type cfg struct {
		Token []byte `env:"ENVX_TEST_BYTES" default:"abc"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(c.Token) != "abc" {
		t.Fatalf("expected abc, got %q", string(c.Token))
	}
}

type customString struct {
	value string
}

func (c *customString) UnmarshalText(text []byte) error {
	c.value = strings.ToUpper(string(text))
	return nil
}

func TestLoad_TextUnmarshaler(t *testing.T) {
	type cfg struct {
		Value customString `env:"ENVX_TEST_TEXT_UNMARSHALER" default:"hello"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Value.value != "HELLO" {
		t.Fatalf("unexpected unmarshaled value: %q", c.Value.value)
	}
}

func TestLoad_UnexportedIgnored(t *testing.T) {
	type cfg struct {
		host string `env:"ENVX_TEST_UNEXPORTED_HOST"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_Pointers(t *testing.T) {
	t.Setenv("ENVX_TEST_PTR_PORT", "8080")

	type cfg struct {
		Port *int `env:"ENVX_TEST_PTR_PORT"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Port == nil {
		t.Fatal("expected Port pointer to be initialized")
	}

	if *c.Port != 8080 {
		t.Fatalf("expected 8080, got %d", *c.Port)
	}
}

func TestLoad_DotEnvFallback(t *testing.T) {
	content := "ENVX_TEST_DOTENV_KEY=\"secret-key-123\"\nENVX_TEST_DOTENV_NAME='My App'"
	tmpFile, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed writing temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed closing temp file: %v", err)
	}

	type cfg struct {
		Key  string `env:"ENVX_TEST_DOTENV_KEY"`
		Name string `env:"ENVX_TEST_DOTENV_NAME"`
	}

	var c cfg
	if err := Load(&c, Options{DotEnvPath: tmpFile.Name()}); err != nil {
		t.Fatalf("failed to load dotenv: %v", err)
	}

	if c.Key != "secret-key-123" {
		t.Errorf("expected secret-key-123, got %s", c.Key)
	}
	if c.Name != "My App" {
		t.Errorf("expected My App, got %s", c.Name)
	}
}

func TestLoad_EnvOverridesDotEnv(t *testing.T) {
	content := "ENVX_TEST_ORDER_KEY=from_file"
	tmpFile, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed writing temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed closing temp file: %v", err)
	}

	t.Setenv("ENVX_TEST_ORDER_KEY", "from_env")

	type cfg struct {
		Key string `env:"ENVX_TEST_ORDER_KEY"`
	}

	var c cfg
	if err := Load(&c, Options{DotEnvPath: tmpFile.Name()}); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if c.Key != "from_env" {
		t.Fatalf("expected env to win, got %q", c.Key)
	}
}

func TestLoad_ExplicitEmptyEnvOverridesDefault(t *testing.T) {
	type cfg struct {
		Value string `env:"ENVX_TEST_EMPTY_OVERRIDES_DEFAULT" default:"fallback"`
	}

	t.Setenv("ENVX_TEST_EMPTY_OVERRIDES_DEFAULT", "")

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if c.Value != "" {
		t.Fatalf("expected explicit empty env value, got %q", c.Value)
	}
}

func TestLoad_DotEnvDoesNotMutateProcessEnv(t *testing.T) {
	content := "ENVX_TEST_NO_MUTATE=from_file"
	tmpFile, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed writing temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed closing temp file: %v", err)
	}

	type cfg struct {
		Value string `env:"ENVX_TEST_NO_MUTATE"`
	}

	var c cfg
	if err := Load(&c, Options{DotEnvPath: tmpFile.Name()}); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if c.Value != "from_file" {
		t.Fatalf("expected value from dotenv fallback, got %q", c.Value)
	}

	if _, ok := os.LookupEnv("ENVX_TEST_NO_MUTATE"); ok {
		t.Fatal("expected dotenv load not to mutate process environment")
	}
}

func TestLoad_SliceTrimming(t *testing.T) {
	t.Setenv("ENVX_TEST_LIST", " one, two , three ")

	type cfg struct {
		List []string `env:"ENVX_TEST_LIST"`
	}

	var c cfg
	if err := Load(&c, Options{}); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	expected := []string{"one", "two", "three"}
	for i, v := range c.List {
		if v != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], v)
		}
	}
}
