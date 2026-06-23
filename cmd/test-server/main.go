package main

import (
	"log"
	"net/http"
)

const (
	testAPIId  = "test-api-id"
	testAPIKey = "test-api-key"
	listenAddr = ":8765"
)

type Server struct {
	state *State
}

func main() {
	state := NewState()
	server := &Server{state: state}
	mux := http.NewServeMux()

	// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_user
	mux.HandleFunc("GET /api-public/v1/user", server.requireAuth(server.handleListUsers))

	// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team
	mux.HandleFunc("GET /api-public/v1/team", server.requireAuth(server.handleListTeams))

	// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team_team_members
	mux.HandleFunc("GET /api-public/v1/team/{slug}/members", server.requireAuth(server.handleListTeamMembers))

	// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/post_api_public_v1_team_team_members
	mux.HandleFunc("POST /api-public/v1/team/{slug}/members", server.requireAuth(server.handleAddTeamMember))

	// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/delete_api_public_v1_team_team_members_user
	mux.HandleFunc("DELETE /api-public/v1/team/{slug}/members/{username}", server.requireAuth(server.handleRemoveTeamMember))

	// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team_team_admins
	mux.HandleFunc("GET /api-public/v1/team/{slug}/admins", server.requireAuth(server.handleListTeamAdmins))

	// Doc URL: https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_oncall_current
	mux.HandleFunc("GET /api-public/v1/oncall/current", server.requireAuth(server.handleOnCallCurrent))

	log.Printf("baton-victorops test server listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
