package settings

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Conf = new(AppConfig)

type AppConfig struct {
	Mode            string `mapstructure:"mode"`
	Port            int    `mapstructure:"port"`
	Name            string `mapstructure:"name"`
	Version         string `mapstructure:"version"`
	StartTime       string `mapstructure:"start_time"`
	MachineID       int    `mapstructure:"machine_id"`
	SmmsToken       string `mapstructure:"smms_token"`
	*LogConfig      `mapstructure:"log"`
	*MySQLConfig    `mapstructure:"mysql"`
	*RedisConfig    `mapstructure:"redis"`
	*RabbitMQConfig `mapstructure:"rabbitmq"`
	*OSSConfig      `mapstructure:"oss"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DB           string `mapstructure:"dbname"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Password     string `mapstructure:"password"`
	Port         int    `mapstructure:"port"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	DefaultTTL   int    `mapstructure:"default_ttl"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type RabbitMQQueueConfig struct {
	Name       string `mapstructure:"name"`
	Exchange   string `mapstructure:"exchange"`
	RoutingKey string `mapstructure:"routing_key"`
}

type RabbitMQConfig struct {
	URL    string                `mapstructure:"url"`
	Queues []RabbitMQQueueConfig `mapstructure:"queues"`
}

type OSSConfig struct {
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyID"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
}

func Init() error {
	// 读取配置文件
	viper.SetConfigFile("./conf/config.yaml")
	// 读取环境变量
	viper.WatchConfig()
	// 监听配置文件变化
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件被修改了...")
		viper.Unmarshal(&Conf)
	})
	// 查找并读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("ReadInConfig failed, err: %v", err))
	}
	// 把读取到的配置信息反序列化到Conf变量中
	if err := viper.Unmarshal(&Conf); err != nil {
		panic(fmt.Errorf("unmarshal to Conf failed, err:%v", err))
	}
	return err
}
