package client

type User struct {
	FirstName           string `json:"firstName"`
	LastName            string `json:"lastName"`
	Username            string `json:"username"`
	Email               string `json:"email"`
	CreatedAt           string `json:"createdAt"`
	PasswordLastUpdated string `json:"passwordLastUpdated"`
	Verified            bool   `json:"verified"`
	SelfUrl             string `json:"_selfUrl"`
}

type Team struct {
	SelfUrl       string `json:"_selfUrl"`
	MembersUrl    string `json:"_membersUrl"`
	PoliciesUrl   string `json:"_policiesUrl"`
	AdminsUrl     string `json:"_adminsUrl"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	MemberCount   int    `json:"memberCount"`
	Version       int    `json:"version"`
	IsDefaultTeam bool   `json:"isDefaultTeam"`
	Description   string `json:"description"`
}

type TeamMemberAdmin struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	SelfUrl   string `json:"_selfUrl"`
}

type TeamMember struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Version   int    `json:"version"`
	Verified  bool   `json:"verified"`
}
