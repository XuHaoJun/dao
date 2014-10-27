package dao

type Shop struct {
	name        string
	itemBaseIds []int
	owner       Bioer
	world       *World
	// may Be log something or show shop style
}

func (s *Shop) NewItemBySellIndex(i int) Itemer {
	if s.itemBaseIds == nil || i > len(s.itemBaseIds) {
		return nil
	}
	w := s.world
	item, err := w.NewItemByBaseId(s.itemBaseIds[i])
	if err != nil {
		return nil
	}
	return item
}

type ShopClient struct {
	Name  string        `json:"name"`
	Items []interface{} `json:"items"`
}

func (s *Shop) ShopClient() *ShopClient {
	return &ShopClient{
		Name:  s.name,
		Items: s.ShopItemsClient(),
	}
}

func (s *Shop) ShopItemsClient() []interface{} {
	var w *World = nil
	if s.world != nil {
		w = s.world
	} else if s.owner.World() != nil {
		w = s.owner.World()
	}
	if w == nil {
		return nil
	}
	shopItemsClient := make([]interface{}, len(s.itemBaseIds))
	for i, id := range s.itemBaseIds {
		item, err := w.NewItemByBaseId(id)
		if err != nil {
			shopItemsClient[i] = nil
		} else {
			shopItemsClient[i] = item.Client()
		}
	}
	return shopItemsClient
}
