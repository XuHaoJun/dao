package dao

import (
	"github.com/nu7hatch/gouuid"
)

type Party struct {
	name     string
	uuid     string
	maxBioer int
	leader   Bioer
	bioers   []Bioer
}

func NewParty() *Party {
	base, _ := uuid.NewV4()
	return &Party{
		uuid:   base.String(),
		bioers: make([]Bioer, 0),
	}
}

type PartyClientBasic struct {
	UUID  string   `json:"uuid"`
	Names []string `json:"names"`
}

func (p *Party) Names() []string {
	names := make([]string, len(p.bioers))
	for i, b := range p.bioers {
		names[i] = b.Name()
	}
	return names
}

func (p *Party) PartyClientBasic() *PartyClientBasic {
	return &PartyClientBasic{
		UUID:  p.uuid,
		Names: p.Names(),
	}
}

func (p *Party) Leader() Bioer {
	return p.leader
}

func (p *Party) IsIn(b Bioer) bool {
	for _, pBio := range p.bioers {
		if pBio == b {
			return true
		}
	}
	return false
}

func (p *Party) Add(b Bioer) {
	if !p.IsIn(b) && (p.maxBioer <= 0 || len(p.bioers) < p.maxBioer) {
		p.bioers = append(p.bioers, b)
	}
}

func (p *Party) Remove(b Bioer) {
	isIn := false
	foundI := 0
	for i, pBio := range p.bioers {
		if pBio == b {
			isIn = true
			foundI = i
		}
	}
	if isIn {
		a := p.bioers
		a[foundI], a[len(a)-1], a = a[len(a)-1], nil, a[:len(a)-1]
	}
}
