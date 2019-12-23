package config

var DBConfig map[string]map[string]string

func init() {
	DBConfig = map[string]map[string]string{
		"default": {
			"dialect":      Conf.Section("DB").Key("Dialect").String(),
			"host":         Conf.Section("DB").Key("HOST").String(),
			"port":         Conf.Section("DB").Key("PORT").String(),
			"database":     Conf.Section("DB").Key("DATABASE").String(),
			"username":     Conf.Section("DB").Key("USERNAME").String(),
			"password":     Conf.Section("DB").Key("PASSWORD").String(),
			"charset":      Conf.Section("DB").Key("CHARSET").String(),
			"maxIdleConns": Conf.Section("DB").Key("MAX_IDLE_CONN").String(),
			"maxOpenConns": Conf.Section("DB").Key("MAX_OPEN_CONN").String(),
		},
	}
}
