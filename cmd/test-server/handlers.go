package main

import (
	"encoding/json"
	"net/http"
)

// requireAuth validates the X-VO-Api-Id and X-VO-Api-Key headers that the VictorOps client
// sends on every request. The test server rejects any request that omits or mismatches either
// header — a permissive server hides auth wiring bugs.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiId := r.Header.Get("X-VO-Api-Id")
		apiKey := r.Header.Get("X-VO-Api-Key")
		if apiId == "" || apiKey == "" {
			writeError(w, http.StatusUnauthorized, "missing X-VO-Api-Id or X-VO-Api-Key header")
			return
		}
		if apiId != testAPIId {
			writeError(w, http.StatusUnauthorized, "invalid X-VO-Api-Id")
			return
		}
		if apiKey != testAPIKey {
			writeError(w, http.StatusUnauthorized, "invalid X-VO-Api-Key")
			return
		}
		next(w, r)
	}
}

// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_user
func (s *Server) handleListUsers(w http.ResponseWriter, _ *http.Request) {
	users := s.state.ListUsers()
	// VictorOps returns users as an array-of-arrays ([][]User). The connector iterates over
	// each inner array and flattens the result. We return all users in one inner array.
	writeJSON(w, map[string]any{"users": []any{users}})
}

// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team
func (s *Server) handleListTeams(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.state.ListTeams())
}

// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team_team_members
func (s *Server) handleListTeamMembers(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	usernames, ok := s.state.GetTeamMembers(slug)
	if !ok {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}
	members := make([]TeamMember, 0, len(usernames))
	for _, username := range usernames {
		u, ok := s.state.GetUser(username)
		if !ok {
			continue
		}
		members = append(members, TeamMember{
			Username:  u.Username,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Verified:  u.Verified,
		})
	}
	writeJSON(w, map[string]any{"members": members})
}

// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team_team_admins
//
// NOTE: the connector has a bug — ListTeamAdmins calls TeamMembersEndpoint instead of
// TeamAdminsEndpoint. The test server implements the correct /admins path per the docs so CI
// exposes the bug (admins will always come back empty from the connector side).
func (s *Server) handleListTeamAdmins(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	usernames, ok := s.state.GetTeamAdmins(slug)
	if !ok {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}
	admins := make([]TeamMemberAdmin, 0, len(usernames))
	for _, username := range usernames {
		u, ok := s.state.GetUser(username)
		if !ok {
			continue
		}
		admins = append(admins, TeamMemberAdmin{
			Username:  u.Username,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			SelfUrl:   u.SelfUrl,
		})
	}
	writeJSON(w, map[string]any{"teamAdmins": admins})
}

// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/post_api_public_v1_team_team_members
func (s *Server) handleAddTeamMember(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "malformed request body")
		return
	}
	if body.Username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}
	_, teamExists, userExists := s.state.AddTeamMember(slug, body.Username)
	if !teamExists {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}
	if !userExists {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	// Idempotent: adding an existing member returns 200, not a conflict.
	writeJSON(w, map[string]any{})
}

// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/delete_api_public_v1_team_team_members_user
func (s *Server) handleRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	username := r.PathValue("username")
	teamExists, _ := s.state.RemoveTeamMember(slug, username)
	if !teamExists {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}
	// Idempotent: removing a non-member is not an error.
	writeJSON(w, map[string]any{})
}

// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_oncall_current
func (s *Server) handleOnCallCurrent(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{"teamsOnCall": s.state.GetTeamsOnCall()})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": message})
}
