package client

import "encoding/json"

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
	SelfUrl       string `json:"_selfUrl,omitempty"`
	MembersUrl    string `json:"_membersUrl,omitempty"`
	PoliciesUrl   string `json:"_policiesUrl,omitempty"`
	AdminsUrl     string `json:"_adminsUrl,omitempty"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	MemberCount   int    `json:"memberCount,omitempty"`
	Version       int    `json:"version,omitempty"`
	IsDefaultTeam bool   `json:"isDefaultTeam,omitempty"`
	Description   string `json:"description,omitempty"`
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

type TeamsOnCallResponse struct {
	TeamsOnCall []TeamOnCall `json:"teamsOnCall"`
}

type TeamOnCall struct {
	Team      Team         `json:"team"`
	OnCallNow []OnCallInfo `json:"onCallNow"`
}

type OnCallInfo struct {
	EscalationPolicy Policy       `json:"escalationPolicy"`
	Users            []OnCallUser `json:"users"`
}

type Policy struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type OnCallUser struct {
	OnCallUser OnCallUserInfo `json:"onCallUser"`
}

type OnCallUserInfo struct {
	Username string `json:"username"`
}

// errorResponse keeps the vendor's error message, or the raw body when it has no "message" field.
type errorResponse struct {
	Msg string
	raw string
}

func (e *errorResponse) UnmarshalJSON(data []byte) error {
	var body struct {
		Message string `json:"message"`
	}
	e.raw = string(data)
	if err := json.Unmarshal(data, &body); err == nil {
		e.Msg = body.Message
	}
	return nil
}

func (e *errorResponse) Message() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.raw
}
