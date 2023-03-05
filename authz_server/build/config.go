package build

import "github.com/spf13/viper"

type config struct {
	DBDriver        string `mapstructure:"DB_DRIVER"`
	DBHost          string `mapstructure:"DB_HOST"`
	DBPort          string `mapstructure:"DB_PORT"`
	DBUsername      string `mapstructure:"DB_USERNAME"`
	DBDatabase      string `mapstructure:"DB_DATABASE"`
	DBPassword      string `mapstructure:"DB_PASSWORD"`
	DBSslMode       string `mapstructure:"DB_SSLMODE"`
	TimeZone        string `mapstructure:"TIME_ZONE"`
	MinioBucket     string `mapstructure:"MINIO_BUCKET"`
	MinioKey        string `mapstructure:"MINIO_KEY"`
	MinioSecret     string `mapstructure:"MINIO_SECRET"`
	MinioUrl        string `mapstructure:"MINIO_URL"`
	MinioSslMode    bool   `mapstructure:"MINIO_SSL"`
	MinioRegion     string `mapstructure:"MINIO_REGION"`
	JwtAccessSecret string `mapstructure:"JWT_ACCESS_SECRET"`
	RedisHost       string `mapstructure:"REDIS_HOST"`
	RedisPort       string `mapstructure:"REDIS_PORT"`
	RedisPassword   string `mapstructure:"REDIS_PASSWORD"`
	RedisDatabase   int `mapstructure:"REDIS_DATABASE"`

}

type ConfigInstance struct {
	config
}

var Config ConfigInstance

func LoadConfig(path string) {
	config := config{}
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	Config = ConfigInstance{
		config,
	}
	return
}
