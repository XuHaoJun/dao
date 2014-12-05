package dao

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type ScriptLoadConfigs struct {
	Imports []string `yaml:"imports,omitempty"`
	Scripts []string `yaml:"scripts,omitempty"`
}

type EtcItemConfigs struct {
	MaxStackCount int `yaml:"maxStackCount"`
}

type UseSelfItemConfigs struct {
	MaxStackCount int `yaml:"maxStackCount"`
}

type ItemConfigs struct {
	EtcItemConfigs     *EtcItemConfigs     `yaml:"etcItem"`
	UseSelfItemConfigs *UseSelfItemConfigs `yaml:"useSelfItem"`
}

type CharFirstScene struct {
	Name string  `yaml:"name"`
	X    float32 `yaml:"x"`
	Y    float32 `yaml:"y"`
}

type CharConfigs struct {
	InitDzeny    int             `yaml:"initDzeny"`
	InitItems    [][]int         `yaml:"initItems"`
	MaxCharItems int             `yaml:"maxCharItems"`
	FirstScene   *CharFirstScene `yaml:"firstScene"`
}

type AccountConfigs struct {
	MaxChars int `yaml:"maxChars"`
}

type MongoDBConfigs struct {
	URL    string `yaml:"url"`
	DBName string `yaml:"dbName"`
}

type WorldConfigs struct {
	Name string `yaml:"name"`
}

type SceneBaseConfig struct {
	AutoClearItemDuration time.Duration `yaml:"autoClearItemDuration"`
}

type SceneConfigs struct {
	Default *SceneBaseConfig            `yaml:"default,omitempty"`
	Custom  map[string]*SceneBaseConfig `yaml:"custom,omitempty"`
}

type Oauth2Config struct {
	ClientId     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
	RedirectURL  string `yaml:"redirectURL"`
	Scope        string `yaml:"scope"`
}

type Oauth2Configs struct {
	Google   *Oauth2Config `yaml:"google"`
	Facebook *Oauth2Config `yaml:"facebook"`
}

func (conf *SceneConfigs) SetScenes(scenes map[string]*Scene) {
	for name, scene := range scenes {
		if conf.Default != nil {
			scene.autoClearItemDuration = conf.Default.AutoClearItemDuration
		}
		if conf.Custom != nil {
			customConfig, ok := conf.Custom[name]
			if ok {
				scene.autoClearItemDuration = customConfig.AutoClearItemDuration
			}
		}
	}
}

type ServerConfigs struct {
	HttpPort      int    `yaml:"httpPort"`
	WebsocketPort int    `yaml:"websocketPort"`
	EnableOauth2  bool   `yaml:"enableOauth2"`
	SessionKey    string `yaml:"sessionKey"`
}

type DaoConfigs struct {
	CharConfigs     *CharConfigs
	AccountConfigs  *AccountConfigs
	WorldConfigs    *WorldConfigs
	MongoDBConfigs  *MongoDBConfigs
	ServerConfigs   *ServerConfigs
	ItemConfigs     *ItemConfigs
	SceneConfigs    *SceneConfigs
	Oauth2Configs   *Oauth2Configs
	ConfigDirPrefix string
	pathMapping     map[string]interface{}
}

func NewDaoConfigs(dirPrefix string) *DaoConfigs {
	dc := &DaoConfigs{
		CharConfigs: &CharConfigs{
			InitDzeny:    0,
			MaxCharItems: 30,
			FirstScene: &CharFirstScene{
				Name: "daoCity",
				X:    0,
				Y:    0,
			},
		},
		AccountConfigs: &AccountConfigs{
			MaxChars: 5,
		},
		WorldConfigs: &WorldConfigs{
			Name: "develop",
		},
		MongoDBConfigs: &MongoDBConfigs{
			URL:    "127.0.0.1",
			DBName: "dao",
		},
		ServerConfigs: &ServerConfigs{
			HttpPort:      3000,
			WebsocketPort: 3001,
			SessionKey:    "DaoSecret",
		},
		ItemConfigs: &ItemConfigs{
			EtcItemConfigs: &EtcItemConfigs{
				MaxStackCount: 100,
			},
			UseSelfItemConfigs: &UseSelfItemConfigs{
				MaxStackCount: 100,
			},
		},
		SceneConfigs: &SceneConfigs{
			Default: &SceneBaseConfig{
				AutoClearItemDuration: 5 * time.Minute,
			},
			Custom: make(map[string]*SceneBaseConfig),
		},
		Oauth2Configs:   &Oauth2Configs{},
		ConfigDirPrefix: "./",
	}
	if dirPrefix != "" {
		dc.ConfigDirPrefix = dirPrefix
	}
	pathMapping := map[string]interface{}{
		dc.ConfigDirPrefix + "conf/char.yaml":    dc.CharConfigs,
		dc.ConfigDirPrefix + "conf/account.yaml": dc.AccountConfigs,
		dc.ConfigDirPrefix + "conf/world.yaml":   dc.WorldConfigs,
		dc.ConfigDirPrefix + "conf/mongodb.yaml": dc.MongoDBConfigs,
		dc.ConfigDirPrefix + "conf/server.yaml":  dc.ServerConfigs,
		dc.ConfigDirPrefix + "conf/item.yaml":    dc.ItemConfigs,
		dc.ConfigDirPrefix + "conf/scene.yaml":   dc.SceneConfigs,
		dc.ConfigDirPrefix + "conf/oauth2.yaml":  dc.Oauth2Configs,
	}
	dc.pathMapping = pathMapping
	return dc
}

func (dc *DaoConfigs) LoadConfigFiles() {
	dc.ReloadConfigFiles()
}

func (dc *DaoConfigs) ReloadConfigFiles() {
	var wg sync.WaitGroup
	wg.Add(len(dc.pathMapping))
	for fname, cg := range dc.pathMapping {
		go func(fileName string, config interface{}) {
			file, e := ioutil.ReadFile(fileName)
			if e != nil {
				wg.Done()
				return
			}
			err := yaml.Unmarshal(file, config)
			if err != nil {
				log.Println("parse config file error: ", fileName, err)
				wg.Done()
				return
			}
			if fileName == "./conf/scene.yaml" {
				conf := config.(*SceneConfigs)
				if conf.Default != nil {
					conf.Default.AutoClearItemDuration *= time.Second
				}
				if conf.Custom != nil {
					for _, c := range conf.Custom {
						c.AutoClearItemDuration *= time.Second
					}
				}
			}
			wg.Done()
		}(fname, cg)
	}
	wg.Wait()
}
