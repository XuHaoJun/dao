package dao

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
)

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

type ServerConfigs struct {
	HttpPort int `yaml:"httpPort"`
}

type DaoConfigs struct {
	CharConfigs    *CharConfigs
	AccountConfigs *AccountConfigs
	WorldConfigs   *WorldConfigs
	MongoDBConfigs *MongoDBConfigs
	ServerConfigs  *ServerConfigs
	ItemConfigs    *ItemConfigs
	pathMapping    map[string]interface{}
}

func NewDefaultDaoConfigs() *DaoConfigs {
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
			HttpPort: 3000,
		},
		ItemConfigs: &ItemConfigs{
			EtcItemConfigs: &EtcItemConfigs{
				MaxStackCount: 100,
			},
			UseSelfItemConfigs: &UseSelfItemConfigs{
				MaxStackCount: 100,
			},
		},
	}
	pathMapping := map[string]interface{}{
		"./conf/char.yaml":    dc.CharConfigs,
		"./conf/account.yaml": dc.AccountConfigs,
		"./conf/world.yaml":   dc.WorldConfigs,
		"./conf/mongodb.yaml": dc.MongoDBConfigs,
		"./conf/server.yaml":  dc.ServerConfigs,
		"./conf/item.yaml":    dc.ItemConfigs,
	}
	dc.pathMapping = pathMapping
	return dc
}

func NewDaoConfigsByConfigFiles() *DaoConfigs {
	dc := NewDefaultDaoConfigs()
	dc.ReloadConfigFiles()
	return dc
}

func (dc *DaoConfigs) ReloadConfigFiles() {
	var wg sync.WaitGroup
	for fname, cg := range dc.pathMapping {
		wg.Add(1)
		go func(fileName string, config interface{}) {
			file, e := ioutil.ReadFile(fileName)
			if e != nil {
				wg.Done()
				return
			}
			err := yaml.Unmarshal(file, config)
			if err != nil {
				log.Println("parse config file error: ", fileName, err)
			}
			wg.Done()
		}(fname, cg)
	}
	wg.Wait()
}
