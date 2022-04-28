package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	General GeneralConfig `toml:"general"`
	Theme   ThemeConfig   `toml:"theme"`
	Keys    KeysConfig    `toml:"keys"`
}

func New() *Config {
	return &Config{
		General: newGeneralConfig(),
		Theme:   newThemeConfig(),
		Keys:    newKeysConfig(),
	}
}

func (c *Config) Load(path string) {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		panic(err)
	}

	if _, err = os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}

		err = toml.NewEncoder(f).Encode(c)
		if err != nil {
			panic(err)
		}
	} else {
		_, err = toml.DecodeFile(path, &c)
		if err != nil {
			panic(err)
		}
	}
}

func Channels() string {
	return `
	841493678274904105
	816706286817247282
	764952301663420416
	823571359845842974
	709804784089301002
	699275726704083064
	892862426498887721
	819778230197682186
	819650801856544851
	815539904573341736
	885856947172147220
	885229686664347648
	817129188837687389

	928367089985679400
	816320427085135892
	717035372248432660
	621847426201944074
	672864319645548571
	771745542471942204
	758034249067659274

	613813861610684421
	
	621529603718119424
	811693600403357706

	865330852962500608


	#FBC

	941323572041875468
	943419513397973012
	966371997649084567
	964926385149866104

	#EC
	906264258877218849
	917304724145991680
	921993047061966848
	901573738586325042

	901573738586325042

	`
}

func DefaultPath() string {
	path, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	path += "/discordo/config.toml"
	return path
}
