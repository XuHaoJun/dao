package dao

import (
  "labix.org/v2/mgo/bson"
  "github.com/gorilla/websocket"
)

type Char struct {
  *BattleBioBase
  id bson.ObjectId
  account *Account
  ws *websocket.Conn
  db *DaoDB
}

type CharDumpDB struct {
    Id bson.ObjectId `bson:"_id"`
    AccountId bson.ObjectId `bson:"accountId"`
    Name string `bson:"name"`
    LastScene *SceneInfo `bson:"lastScene"`
}

func NewChar(acc *Account, ws *websocket.Conn) *Char {
  return &Char{
    BattleBioBase: NewBattleBioBase(),
    id: bson.NewObjectId(),
    account: acc,
    ws: ws,
    db: acc.world.db.clone(),
  }
}

func (c *Char) Save() {
  c.job <- func() {
    chars := c.db.chars
    if _, err := chars.UpsertId(c.id, c.DumpDB); err != nil {
      panic(err)
    }
  }
}

func (c *Char) DumpDB() *CharDumpDB {
  return &CharDumpDB{
    Id: c.id,
    AccountId: c.account.id,
    Name: c.name,
    LastScene: &SceneInfo{c.scene.name, c.pos.x, c.pos.y},
  }
}

func (c *Char) ReadClient(msg []byte) {
}
