package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("APP_ENV", "test")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "test-db")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("REDIS_HOST", "test-redis")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("LOG_LEVEL", "debug")

	defer func() {
		// Clean up
		os.Unsetenv("APP_ENV")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded values
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Environment", cfg.Environment, "test"},
		{"Port", cfg.Server.Port, "9090"},
		{"DB Host", cfg.Database.Host, "test-db"},
		{"DB Port", cfg.Database.Port, "5433"},
		{"DB User", cfg.Database.User, "testuser"},
		{"DB Password", cfg.Database.Password, "testpass"},
		{"DB Name", cfg.Database.Name, "testdb"},
		{"Redis Host", cfg.Redis.Host, "test-redis"},
		{"Redis Port", cfg.Redis.Port, "6380"},
		{"JWT Secret", cfg.Auth.JWTSecret, "test-secret"},
		{"Log Level", cfg.Logging.Level, "debug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestLoadWithDefaults(t *testing.T) {
	// Clear environment variables to test defaults
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify default values
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Default Environment", cfg.Environment, "development"},
		{"Default Port", cfg.Server.Port, "8080"},
		{"Default DB Host", cfg.Database.Host, "localhost"},
		{"Default DB Port", cfg.Database.Port, "5432"},
		{"Default Redis Host", cfg.Redis.Host, "localhost"},
		{"Default Redis Port", cfg.Redis.Port, "6379"},
		{"Default Log Level", cfg.Logging.Level, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"development environment", "development", true},
		{"production environment", "production", false},
		{"test environment", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}
			if got := cfg.IsDevelopment(); got != tt.want {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"production environment", "production", true},
		{"development environment", "development", false},
		{"test environment", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}
			if got := cfg.IsProduction(); got != tt.want {
				t.Errorf("IsProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetDatabaseDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "secret",
			Name:     "testdb",
			SSLMode:  "disable",
		},
	}

	want := "host=localhost port=5432 user=postgres password=secret dbname=testdb sslmode=disable"
	got := cfg.GetDatabaseDSN()

	if got != want {
		t.Errorf("GetDatabaseDSN() = %v, want %v", got, want)
	}
}

func TestConfig_GetRedisAddr(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
	}

	want := "localhost:6379"
	got := cfg.GetRedisAddr()

	if got != want {
		t.Errorf("GetRedisAddr() = %v, want %v", got, want)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Environment: "production",
				Server:      ServerConfig{Port: "8080"},
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "user",
					Password: "pass",
					Name:     "db",
				},
				Auth: AuthConfig{JWTSecret: "secret"},
			},
			wantErr: false,
		},
		{
			name: "missing JWT secret in production",
			cfg: &Config{
				Environment: "production",
				Server:      ServerConfig{Port: "8080"},
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "user",
					Password: "pass",
					Name:     "db",
				},
				Auth: AuthConfig{JWTSecret: ""},
			},
			wantErr: true,
		},
		{
			name: "missing database password in production",
			cfg: &Config{
				Environment: "production",
				Server:      ServerConfig{Port: "8080"},
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "user",
					Password: "",
					Name:     "db",
				},
				Auth: AuthConfig{JWTSecret: "secret"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		name         string
		key          string
		defaultValue string
		want         string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_VAR",
			defaultValue: "default",
			want:         "test_value",
		},
		{
			name:         "returns default when env not set",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEnv(tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         "parses valid duration",
			value:        "5m",
			defaultValue: 1 * time.Minute,
			want:         5 * time.Minute,
		},
		{
			name:         "returns default for invalid duration",
			value:        "invalid",
			defaultValue: 1 * time.Minute,
			want:         1 * time.Minute,
		},
		{
			name:         "returns default for empty string",
			value:        "",
			defaultValue: 2 * time.Minute,
			want:         2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDuration(tt.value, tt.defaultValue); got != tt.want {
				t.Errorf("parseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
