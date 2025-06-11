package env

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type EnvVal struct {
	stringVal string
}

type Config struct {
	InitSqlite bool
	SqlitePath string
}

func LoadConfig(path string) (*Config, error) {

	if err := godotenv.Load(path); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg := &Config{
		InitSqlite: getEnv("INIT_SQLITE", "true").Bool(),
		SqlitePath: getEnv("SQLITE_PATH", "data/app.db").String(),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) EnvVal {
	if stringVal, exists := os.LookupEnv(key); exists {
		return EnvVal{stringVal: stringVal}
	}
	return EnvVal{stringVal: defaultValue}
}

func (ev EnvVal) String() string {
	return ev.stringVal
}

func (ev EnvVal) Bool() bool {
	val := strings.ToLower(ev.stringVal)
	return val == "true" || val == "1" || val == "yes" || val == "on"
}

func (ev EnvVal) Int() (int, error) {
	return strconv.Atoi(ev.stringVal)
}

func (ev EnvVal) IntDefault(defaultValue int) int {
	if val, err := strconv.Atoi(ev.stringVal); err == nil {
		return val
	}
	return defaultValue
}

func (ev EnvVal) Float64() (float64, error) {
	return strconv.ParseFloat(ev.stringVal, 64)
}

func (ev EnvVal) Float64Default(defaultValue float64) float64 {
	if val, err := strconv.ParseFloat(ev.stringVal, 64); err == nil {
		return val
	}
	return defaultValue
}

func (ev EnvVal) StringSlice(sep string) []string {
	if ev.stringVal == "" {
		return []string{}
	}
	return strings.Split(ev.stringVal, sep)
}
