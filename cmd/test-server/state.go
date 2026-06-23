package main

import (
	"slices"
	"sync"
)

// User mirrors the shape from
// https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_user
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

// Team mirrors the shape from
// https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team
type Team struct {
	SelfUrl       string `json:"_selfUrl,omitempty"`
	MembersUrl    string `json:"_membersUrl,omitempty"`
	PoliciesUrl   string `json:"_policiesUrl,omitempty"`
	AdminsUrl     string `json:"_adminsUrl,omitempty"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	MemberCount   int    `json:"memberCount"`
	Version       int    `json:"version"`
	IsDefaultTeam bool   `json:"isDefaultTeam"`
	Description   string `json:"description,omitempty"`
}

// TeamMember mirrors the shape from
// https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team_team_members
type TeamMember struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Version   int    `json:"version"`
	Verified  bool   `json:"verified"`
}

// TeamMemberAdmin mirrors the shape from
// https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team_team_admins
type TeamMemberAdmin struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	SelfUrl   string `json:"_selfUrl"`
}

// Policy, OnCallUserInfo, OnCallUser, OnCallInfo, TeamOnCall mirror the shape from
// https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_oncall_current
type Policy struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type OnCallUserInfo struct {
	Username string `json:"username"`
}

type OnCallUser struct {
	OnCallUser OnCallUserInfo `json:"onCallUser"`
}

type OnCallInfo struct {
	EscalationPolicy Policy       `json:"escalationPolicy"`
	Users            []OnCallUser `json:"users"`
}

type TeamOnCall struct {
	Team      Team         `json:"team"`
	OnCallNow []OnCallInfo `json:"onCallNow"`
}

type State struct {
	mu          sync.Mutex
	users       map[string]*User    // username → User
	userList    []*User             // insertion-ordered for deterministic list responses
	teams       map[string]*Team    // slug → Team
	teamList    []*Team             // insertion-ordered
	members     map[string][]string // teamSlug → []username
	admins      map[string][]string // teamSlug → []username (immutable after seed)
	teamsOnCall []TeamOnCall        // seeded on-call state
}

func NewState() *State {
	s := &State{
		users:   make(map[string]*User),
		teams:   make(map[string]*Team),
		members: make(map[string][]string),
		admins:  make(map[string][]string),
	}
	seed(s)
	return s
}

func (s *State) GetUser(username string) (*User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[username]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

func (s *State) ListUsers() []*User {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*User, len(s.userList))
	for i, u := range s.userList {
		cp := *u
		out[i] = &cp
	}
	return out
}

func (s *State) ListTeams() []*Team {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Team, len(s.teamList))
	for i, t := range s.teamList {
		cp := *t
		out[i] = &cp
	}
	return out
}

func (s *State) GetTeamMembers(slug string) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.teams[slug]; !ok {
		return nil, false
	}
	return slices.Clone(s.members[slug]), true
}

func (s *State) GetTeamAdmins(slug string) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.teams[slug]; !ok {
		return nil, false
	}
	return slices.Clone(s.admins[slug]), true
}

// AddTeamMember appends username to the team member list. Idempotent — adding an already-present
// member is a no-op and returns alreadyMember=true. Returns teamExists=false if the slug is
// unknown, userExists=false if the username is unknown.
func (s *State) AddTeamMember(teamSlug, username string) (alreadyMember, teamExists, userExists bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.teams[teamSlug]; !ok {
		return false, false, false
	}
	if _, ok := s.users[username]; !ok {
		return false, true, false
	}
	if slices.Contains(s.members[teamSlug], username) {
		return true, true, true
	}
	s.members[teamSlug] = append(s.members[teamSlug], username)
	return false, true, true
}

// RemoveTeamMember removes username from the team member list. Returns teamExists=false if the
// slug is unknown. Removing a non-member is a no-op (wasMember=false) but not an error.
func (s *State) RemoveTeamMember(teamSlug, username string) (teamExists, wasMember bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.teams[teamSlug]; !ok {
		return false, false
	}
	if !slices.Contains(s.members[teamSlug], username) {
		return true, false
	}
	s.members[teamSlug] = slices.DeleteFunc(s.members[teamSlug], func(u string) bool { return u == username })
	return true, true
}

func (s *State) GetTeamsOnCall() []TeamOnCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.teamsOnCall)
}
