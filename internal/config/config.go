// Пакет для работы с файлом конфига.
package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Структура конфига
type Config struct {
	StoragePath   string   `yaml:"storage_path"`
	StorageUser   string   `yaml:"storage_user"`
	StoragePasswd string   `yaml:"storage_passwd"`
	ContentLength int      `yaml:"content_length"`
	CensorList    []string `yaml:"censor_list"`
	HTTPServer    `yaml:"http_server"`
}
type HTTPServer struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// MustLoad - инициализирует данные из конфиг файла. Путь к файлу берет из
// переменной окружения COMMENTS_CONFIG_PATH, пароль для доступа к БД - из переменной
// окружения COMMENTS_DB_PASSWD. Если не удается, то завершает приложение с ошибкой.
func MustLoad() *Config {
	configPath := os.Getenv("COMMENTS_CONFIG_PATH")
	if configPath == "" {
		log.Fatal("COMMENTS_CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("cannot read config file: %s, %s", configPath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("cannot decode config file: %s, %s", configPath, err)
	}

	cfg.StoragePasswd = os.Getenv("MONGO_DB_PASSWD")
	if cfg.StoragePasswd == "" {
		log.Printf("MONGO_DB_PASSWD is not set\n")
	}

	return &cfg
}
