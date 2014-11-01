package dao

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type EtcItemConfigs struct {
	MaxStackCount int `json:"maxStackCount"`
}

type UseSelfItemConfigs struct {
	MaxStackCount int `json:"maxStackCount"`
}

type ItemConfigs struct {
	EtcItemConfigs     *EtcItemConfigs     `json:"etcItem"`
	UseSelfItemConfigs *UseSelfItemConfigs `json:"useSelfItem"`
}

type CharFirstScene struct {
	Name string  `json:"name"`
	X    float32 `json:"x"`
	Y    float32 `json:"y"`
}

type CharConfigs struct {
	InitDzeny    int `json:"initDzeny"`
	MaxCharItems int `json:"maxCharItems"`
	FirstScene   *CharFirstScene
}

type AccountConfigs struct {
	MaxChars int `json:"maxChars"`
}

type MongoDBConfigs struct {
	URL    string `json:"url"`
	DBName string `json:"dbName"`
}

type WorldConfigs struct {
	Name string `json:"name"`
}

type ServerConfigs struct {
	HttpPort int `json:"httpPort"`
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
		"./conf/char.json":    dc.CharConfigs,
		"./conf/account.json": dc.AccountConfigs,
		"./conf/world.json":   dc.WorldConfigs,
		"./conf/mongodb.json": dc.MongoDBConfigs,
		"./conf/server.json":  dc.ServerConfigs,
		"./conf/item.json":    dc.ItemConfigs,
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
			json.Unmarshal(file, config)
			wg.Done()
		}(fname, cg)
	}
	wg.Wait()
}
