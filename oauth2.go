package dao

type FacebookProfile struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Link        string `json:"link"`
	Locale      string `json:"locale"`
	Verified    bool   `json:"verified"`
	Id          string `json:"id"`
	Gender      string `json:"male"`
	Name        string `json:"name"`
	Timezone    int    `json:"timezone"`
	UpdatedTime string `json:"updated_time"`
}
