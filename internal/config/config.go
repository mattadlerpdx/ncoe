package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerAddress string
	DatabaseURL   string
	Environment   string
	TemplateDir   string // Absolute path to templates directory
	StaticDir     string // Absolute path to static directory
	Branding      Branding
}

type Branding struct {
	AgencyName     string `yaml:"agency_name"`
	ShortName      string `yaml:"short_name"`
	Tagline        string `yaml:"tagline"`
	Logo           string `yaml:"logo"`
	Favicon        string `yaml:"favicon"`
	PrimaryColor   string `yaml:"primary_color"`
	SecondaryColor string `yaml:"secondary_color"`
	AccentColor    string `yaml:"accent_color"`
	ContactEmail   string `yaml:"contact_email"`
	ContactPhone   string `yaml:"contact_phone"`
	Website        string `yaml:"website"`
	Address        string `yaml:"address"`
}

func Load() *Config {
	cfg := &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8081"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		TemplateDir:   getEnv("TEMPLATE_DIR", "templates"),
		StaticDir:     getEnv("STATIC_DIR", "static"),
	}

	// Load branding from YAML
	brandingFile := getEnv("BRANDING_CONFIG", "config/branding.yaml")
	if data, err := os.ReadFile(brandingFile); err == nil {
		yaml.Unmarshal(data, &cfg.Branding)
	} else {
		// Default branding
		cfg.Branding = Branding{
			AgencyName:     "Nevada Commission on Ethics",
			ShortName:      "NCOE",
			Tagline:        "Integrity and Trust",
			PrimaryColor:   "#003366",
			SecondaryColor: "#C4A000",
			ContactEmail:   "ncoe@ethics.nv.gov",
			ContactPhone:   "(775) 687-5469",
			Website:        "https://ethics.nv.gov",
		}
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
