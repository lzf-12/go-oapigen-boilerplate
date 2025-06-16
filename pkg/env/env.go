package env

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type EnvVariable struct {
	stringVal string
}

type Config struct {
	InitSqlite bool
	SqlitePath string
	Env        string
	DebugMode  bool
}

type Environment int

const (
	Dev Environment = iota
	Staging
	Production
)

func (e Environment) String() string {
	switch e {
	case Dev:
		return "dev"
	case Staging:
		return "staging"
	case Production:
		return "production"
	default:
		return "unknown"
	}
}

func LoadConfig(path string) (*Config, error) {

	if err := godotenv.Load(path); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg := &Config{
		InitSqlite: getEnv("INIT_SQLITE", "true").Bool(),
		SqlitePath: getEnv("SQLITE_PATH", "data/app.db").String(),
		Env:        getEnv("ENV", "").String(),
		DebugMode:  getEnv("DEBUG_MODE", "").Bool(),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) EnvVariable {
	if stringVal, exists := os.LookupEnv(key); exists {
		return EnvVariable{stringVal: stringVal}
	}
	return EnvVariable{stringVal: defaultValue}
}

func (ev EnvVariable) String() string {
	return ev.stringVal
}

func (ev EnvVariable) Bool() bool {
	val := strings.ToLower(ev.stringVal)
	return val == "true" || val == "1" || val == "yes" || val == "on"
}

func (ev EnvVariable) Int() (int, error) {
	return strconv.Atoi(ev.stringVal)
}

func (ev EnvVariable) IntDefault(defaultValue int) int {
	if val, err := strconv.Atoi(ev.stringVal); err == nil {
		return val
	}
	return defaultValue
}

func (ev EnvVariable) Float64() (float64, error) {
	return strconv.ParseFloat(ev.stringVal, 64)
}

func (ev EnvVariable) Float64Default(defaultValue float64) float64 {
	if val, err := strconv.ParseFloat(ev.stringVal, 64); err == nil {
		return val
	}
	return defaultValue
}

func (ev EnvVariable) StringSlice(sep string) []string {
	if ev.stringVal == "" {
		return []string{}
	}
	return strings.Split(ev.stringVal, sep)
}
