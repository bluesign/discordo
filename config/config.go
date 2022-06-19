package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/gdamore/tcell/v2"
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
	
	969934070849077301

	720983219205570570
	970402696731431005
	970574640378417202
	984358040553795617
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

type UIConfig struct {
	PinnedTabMarker string `ini:"pinned-tab-marker"`
	style           StyleSet
}

type BindingConfig struct {
	Global *KeyBindings
}

type MainConfig struct {
	Ui       UIConfig
	Bindings BindingConfig
}

// Input: TimestampFormat
// Output: timestamp-format
func mapName(raw string) string {
	newstr := make([]rune, 0, len(raw))
	for i, chr := range raw {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if i > 0 {
				newstr = append(newstr, '-')
			}
		}
		newstr = append(newstr, unicode.ToLower(chr))
	}
	return string(newstr)
}

func installTemplate(root, sharedir, name string) error {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
	}
	data, err := ioutil.ReadFile(path.Join(sharedir, name))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(root, name), data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (config *MainConfig) LoadConfig() error {

	if err := config.Ui.loadStyleSet(); err != nil {
		return err
	}
	return nil
}

func LoadConfig() (*MainConfig, error) {

	config := &MainConfig{
		Ui: UIConfig{},
		Bindings: BindingConfig{
			Global: NewKeyBindings(),
		},
	}
	config.LoadConfig()
	return config, nil
}

func parseLayout(layout string) [][]string {
	rows := strings.Split(layout, ",")
	l := make([][]string, len(rows))
	for i, r := range rows {
		l[i] = strings.Split(r, "|")
	}
	return l
}

func (ui *UIConfig) loadStyleSet() error {
	ui.style = NewStyleSet()
	err := ui.style.LoadStyleSet("./default_styleset")
	if err != nil {
		return fmt.Errorf("Unable to load default styleset: %s", err)
	}

	return nil
}

func (uiConfig UIConfig) GetStyle(so StyleObject) tcell.Style {
	return uiConfig.style.Get(so)
}

func (uiConfig UIConfig) GetStyleSelected(so StyleObject) tcell.Style {
	return uiConfig.style.Selected(so)
}

func (uiConfig UIConfig) GetComposedStyle(base StyleObject,
	styles []StyleObject) tcell.Style {
	return uiConfig.style.Compose(base, styles)
}

func (uiConfig UIConfig) GetComposedStyleSelected(base StyleObject, styles []StyleObject) tcell.Style {
	return uiConfig.style.ComposeSelected(base, styles)
}
