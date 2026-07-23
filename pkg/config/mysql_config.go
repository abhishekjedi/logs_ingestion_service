package config

import "github.com/knadh/koanf/v2"

type MysqlConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewMysqlConfig(k *koanf.Koanf) MysqlConfig {
	return MysqlConfig{
		Host:     k.String("mysql.host"),
		Port:     k.String("mysql.port"),
		User:     k.String("mysql.user"),
		Password: k.String("mysql.password"),
		DBName:   k.String("mysql.dbname"),
	}
}
