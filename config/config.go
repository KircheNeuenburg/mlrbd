package config

import (
	"github.com/go-ini/ini"
	"github.com/kyoh86/xdg"
	"path"
)

type GeneralConfig struct {
	KeepRooms    bool `ini:"keep-rooms"`
	SyncInterval int  `ini:"sync-interval"`
}

type DbConfig struct {
	Connection string `ini:"connection"`
}

type LdapConfig struct {
	Server                 string `ini:"server"`
	Port                   int    `ini:"port"`
	BindDn                 string `ini:"bind-dn"`
	BindPassword           string `ini:"bind-password"`
	GroupFilter            string `ini:"group-filter"`
	GroupBaseDn            string `ini:"group-base-dn"`
	UserBaseDn             string `ini:"user-base-dn"`
	UserFilter             string `ini:"user-filter"`
	GroupMemberAssociation string `ini:"group-association"`
	GroupMemberAttribute   string `ini:"group-member-attribute"`
	GroupUniqueIdentifier  string `ini:"group-unique-identifier"`
	GroupName              string `ini:"group-name"`
	UserLoginAttribute     string `ini:"user-login-attribute"`
}

type MatrixConfig struct {
	E2eEncryption bool   `ini:"e2e-encryption"`
	Homeserver    string `ini:"homeserver"`
	Mxid          string `ini:"mxid"`
	AccessToken   string `ini:"access-token"`
	KickMessage   string `ini:"kick-message"`
}

type Config struct {
	General GeneralConfig
	Db      DbConfig
	Ldap    LdapConfig
	Matrix  MatrixConfig
}

func (config *Config) LoadConfig(file *ini.File) error {
	if general, err := file.GetSection("general"); err == nil {
		if err := general.MapTo(&config.General); err != nil {
			return err
		}
	}
	if database, err := file.GetSection("database"); err == nil {
		if err := database.MapTo(&config.Db); err != nil {
			return err
		}
	}
	if ldap, err := file.GetSection("ldap"); err == nil {
		if err := ldap.MapTo(&config.Ldap); err != nil {
			return err
		}
	}
	if matrix, err := file.GetSection("matrix"); err == nil {
		if err := matrix.MapTo(&config.Matrix); err != nil {
			return err
		}
	}
	return nil
}

func LoadConfigFromFile(root *string, sharedir string) (*Config, error) {
	if root == nil {
		_root := path.Join(xdg.ConfigHome(), "mlrbd")
		root = &_root
	}
	filename := path.Join(*root, "mlrbd.conf")
	file, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}
	config := &Config{General: GeneralConfig{SyncInterval: 5}}
	if err = config.LoadConfig(file); err != nil {
		return nil, err
	}
	return config, nil
}
