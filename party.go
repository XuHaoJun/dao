package dao

import (
	"errors"
	"github.com/nu7hatch/gouuid"
)

type Party struct {
	name      string
	uuid      string
	maxMember int
	leader    Bioer
	members   []Bioer
}

func NewParty() *Party {
	base, _ := uuid.NewV4()
	return &Party{
		uuid:    base.String(),
		members: make([]Bioer, 0),
	}
}

type MemberInfo struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type PartyClient struct {
	UUID        string        `json:"uuid"`
	Name        string        `json:"name"`
	MemberInfos []*MemberInfo `json:"memberInfos"`
}

type PartyClientBasic struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func (p *Party) MemberInfos() []*MemberInfo {
	memberInfos := make([]*MemberInfo, len(p.members))
	for i, b := range p.members {
		memberInfos[i] = &MemberInfo{
			b.Name(),
			b.Level(),
		}
	}
	return memberInfos
}

func (p *Party) PartyClient() *PartyClient {
	return &PartyClient{
		UUID:        p.uuid,
		Name:        p.name,
		MemberInfos: p.MemberInfos(),
	}
}

func (p *Party) PartyClientBasic() *PartyClientBasic {
	return &PartyClientBasic{
		UUID: p.uuid,
		Name: p.name,
	}
}

func (p *Party) CharMembers() []Charer {
	chars := make([]Charer, 0)
	for _, bio := range p.members {
		char, isChar := bio.(Charer)
		if isChar {
			chars = append(chars, char)
		}
	}
	return chars
}

func (p *Party) Leader() Bioer {
	return p.leader
}

func (p *Party) IsIn(b Bioer) bool {
	for _, pBio := range p.members {
		if pBio == b {
			return true
		}
	}
	return false
}

func (p *Party) Add(b Bioer) error {
	if !p.IsIn(b) && (p.maxMember <= 0 || len(p.members) < p.maxMember) {
		p.members = append(p.members, b)
		return nil
	}
	return errors.New("wrong")
}

func (p *Party) HasMember() bool {
	return len(p.members) > 0
}

func (p *Party) Remove(b Bioer) {
	isIn := false
	foundI := 0
	for i, pBio := range p.members {
		if pBio == b {
			isIn = true
			foundI = i
		}
	}
	if isIn {
		a := p.members
		a[foundI], a[len(a)-1], a = a[len(a)-1], nil, a[:len(a)-1]
		p.members = a
		b.SetParty(nil)
		if p.leader == b {
			p.leader = nil
		}
	}
}
