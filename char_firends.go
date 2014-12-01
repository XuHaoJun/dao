package dao

type CharFriends struct {
	owner            Charer
	friendNames      []string
	closeFriendNames []string
}

type CharFriendsDumpDB struct {
	FriendNames      []string `bson:"friendNames"`
	CloseFriendNames []string `bson:"closeFriendNames"`
}

func NewCharFriends(owner Charer) *CharFriends {
	return &CharFriends{
		owner:            owner,
		friendNames:      []string{},
		closeFriendNames: []string{},
	}
}

func (f *CharFriends) DumpDB() *CharFriendsDumpDB {
	return &CharFriendsDumpDB{
		FriendNames:      f.friendNames,
		CloseFriendNames: f.closeFriendNames,
	}
}

func (fDump *CharFriendsDumpDB) Load(owner Charer) *CharFriends {
	f := NewCharFriends(owner)
	f.friendNames = fDump.FriendNames
	f.closeFriendNames = fDump.CloseFriendNames
	return f
}

func (f *CharFriends) OnlineFriends() map[string]Charer {
	onlineFriends := make(map[string]Charer)
	onlineChars := f.owner.World().OnlineChars()
	for name, char := range onlineChars {
		if f.HasFriend(name) {
			onlineFriends[name] = char
		}
	}
	return onlineFriends
}

func (f *CharFriends) OfflineFriends() []string {
	offlineFriends := make([]string, len(f.friendNames))
	copy(offlineFriends, f.friendNames)
	onlineChars := f.owner.World().OnlineChars()
	for name, _ := range onlineChars {
		for i, fname := range f.friendNames {
			if name == fname {
				offlineFriends = append(offlineFriends[:i],
					offlineFriends[i+1:]...)
				continue
			}
		}
	}
	return offlineFriends
}

func (f *CharFriends) HasFriend(name string) bool {
	for _, fname := range f.friendNames {
		if fname == name {
			return true
		}
	}
	return false
}
