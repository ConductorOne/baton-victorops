package connector

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-victorops/pkg/connector/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const testTeamSlug = "team-alpha"

type recordedRequest struct {
	method string
	path   string
	body   string
}

type fakeVictorOps struct {
	mu       sync.Mutex
	requests []recordedRequest
	handlers map[string]http.HandlerFunc
}

func newFakeVictorOps(t *testing.T, handlers map[string]http.HandlerFunc) (*fakeVictorOps, *client.VictorOpsClient) {
	t.Helper()

	f := &fakeVictorOps{handlers: handlers}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.requests = append(f.requests, recordedRequest{method: r.Method, path: r.URL.Path, body: string(body)})
		f.mu.Unlock()

		h, ok := f.handlers[r.Method+" "+r.URL.Path]
		if !ok {
			writeTestJSON(w, http.StatusNotFound, `{"message":"not found"}`)
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)

	c, err := client.NewVictorOpsClient(context.Background(), "api-id", "api-key", srv.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return f, c
}

func (f *fakeVictorOps) count(method, path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	n := 0
	for _, r := range f.requests {
		if r.method == method && r.path == path {
			n++
		}
	}
	return n
}

func (f *fakeVictorOps) all() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]recordedRequest(nil), f.requests...)
}

func writeTestJSON(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}

func respond(code int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeTestJSON(w, code, body)
	}
}

func testTeamResource(t *testing.T) *v2.Resource {
	t.Helper()

	r, err := teamResource(&client.Team{Name: "Team Alpha", Slug: testTeamSlug})
	if err != nil {
		t.Fatalf("failed to build team resource: %v", err)
	}
	return r
}

func testUserResource(t *testing.T, username string) *v2.Resource {
	t.Helper()

	r, err := userResource(context.Background(), &client.User{Username: username, Email: username})
	if err != nil {
		t.Fatalf("failed to build user resource: %v", err)
	}
	return r
}

func testEntitlement(team *v2.Resource, slug string) *v2.Entitlement {
	return &v2.Entitlement{
		Id:       team.Id.ResourceType + ":" + team.Id.Resource + ":" + slug,
		Resource: team,
		Slug:     slug,
	}
}

const (
	membersPath = "/api-public/v1/team/" + testTeamSlug + "/members"
	adminsPath  = "/api-public/v1/team/" + testTeamSlug + "/admins"
)

func TestTeamGrants(t *testing.T) {
	f, c := newFakeVictorOps(t, map[string]http.HandlerFunc{
		"GET " + membersPath: respond(http.StatusOK, `{"members":[{"username":"alice"},{"username":"bob"}]}`),
		"GET " + adminsPath:  respond(http.StatusOK, `{"teamAdmins":[{"username":"alice"},{"username":"carol"}]}`),
	})
	team := testTeamResource(t)

	grants, _, err := newTeamBuilder(c, "").Grants(context.Background(), team, rs.SyncOpAttrs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n := f.count(http.MethodGet, membersPath); n != 1 {
		t.Errorf("expected 1 call to members endpoint, got %d", n)
	}
	if n := f.count(http.MethodGet, adminsPath); n != 1 {
		t.Errorf("expected 1 call to admins endpoint, got %d", n)
	}

	got := map[string][]string{}
	for _, g := range grants {
		if g.Principal.Id.ResourceType != userResourceType.Id {
			t.Errorf("unexpected principal type %s", g.Principal.Id.ResourceType)
		}
		slug := g.Entitlement.Id[strings.LastIndex(g.Entitlement.Id, ":")+1:]
		got[slug] = append(got[slug], g.Principal.Id.Resource)

		grantAnnos := annotations.Annotations(g.Annotations)
		isImmutable := grantAnnos.Contains(&v2.GrantImmutable{})
		if slug == teamAdminEntitlement && !isImmutable {
			t.Errorf("admin grant for %s is not immutable", g.Principal.Id.Resource)
		}
		if slug == teamMemberEntitlement && isImmutable {
			t.Errorf("member grant for %s should not be immutable", g.Principal.Id.Resource)
		}
	}

	if strings.Join(got[teamMemberEntitlement], ",") != "alice,bob" {
		t.Errorf("unexpected member grants: %v", got[teamMemberEntitlement])
	}
	if strings.Join(got[teamAdminEntitlement], ",") != "alice,carol" {
		t.Errorf("unexpected admin grants: %v", got[teamAdminEntitlement])
	}
}

func TestTeamGrantsAdminsError(t *testing.T) {
	for _, code := range []int{http.StatusForbidden, http.StatusInternalServerError} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			_, c := newFakeVictorOps(t, map[string]http.HandlerFunc{
				"GET " + membersPath: respond(http.StatusOK, `{"members":[{"username":"alice"}]}`),
				"GET " + adminsPath:  respond(code, `{"message":"failure"}`),
			})

			grants, _, err := newTeamBuilder(c, "").Grants(context.Background(), testTeamResource(t), rs.SyncOpAttrs{})
			if err == nil {
				t.Fatal("expected an error")
			}
			if grants != nil {
				t.Errorf("expected no grants, got %d", len(grants))
			}
		})
	}
}

func TestTeamEntitlementDescriptions(t *testing.T) {
	ents, _, err := newTeamBuilder(nil, "").Entitlements(context.Background(), testTeamResource(t), rs.SyncOpAttrs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]string{
		teamMemberEntitlement: "Member of Team Alpha team",
		teamAdminEntitlement:  "Admin of Team Alpha team",
	}
	for _, e := range ents {
		if e.Description != want[e.Slug] {
			t.Errorf("entitlement %s: got description %q, want %q", e.Slug, e.Description, want[e.Slug])
		}
	}
}

func TestTeamGrant(t *testing.T) {
	f, c := newFakeVictorOps(t, map[string]http.HandlerFunc{
		"POST " + membersPath: respond(http.StatusOK, `{}`),
	})
	team := testTeamResource(t)
	user := testUserResource(t, "alice")

	grants, annos, err := newTeamBuilder(c, "").Grant(context.Background(), user, testEntitlement(team, teamMemberEntitlement))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(annos) != 0 {
		t.Errorf("expected no annotations, got %v", annos)
	}
	if len(grants) != 1 {
		t.Fatalf("expected 1 grant, got %d", len(grants))
	}

	g := grants[0]
	if g.Entitlement.Resource.Id.ResourceType != teamResourceType.Id || g.Entitlement.Resource.Id.Resource != testTeamSlug {
		t.Errorf("unexpected entitlement resource %v", g.Entitlement.Resource.Id)
	}
	if want := "team:" + testTeamSlug + ":" + teamMemberEntitlement; g.Entitlement.Id != want {
		t.Errorf("unexpected entitlement id %q, want %q", g.Entitlement.Id, want)
	}
	if g.Principal.Id.ResourceType != userResourceType.Id || g.Principal.Id.Resource != "alice" {
		t.Errorf("unexpected principal %v", g.Principal.Id)
	}

	reqs := f.all()
	if len(reqs) != 1 || strings.TrimSpace(reqs[0].body) != `{"username":"alice"}` {
		t.Errorf("unexpected requests: %+v", reqs)
	}
}

func TestTeamGrantAlreadyExists(t *testing.T) {
	_, c := newFakeVictorOps(t, map[string]http.HandlerFunc{
		"POST " + membersPath: respond(http.StatusConflict, `{"message":"already a member"}`),
	})
	team := testTeamResource(t)

	grants, annos, err := newTeamBuilder(c, "").Grant(context.Background(), testUserResource(t, "alice"), testEntitlement(team, teamMemberEntitlement))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(grants) != 1 {
		t.Errorf("expected 1 grant, got %d", len(grants))
	}
	if !annos.Contains(&v2.GrantAlreadyExists{}) {
		t.Errorf("expected GrantAlreadyExists annotation, got %v", annos)
	}
}

func TestTeamGrantRevokeAdminRejected(t *testing.T) {
	f, c := newFakeVictorOps(t, nil)
	team := testTeamResource(t)
	user := testUserResource(t, "alice")
	adminEnt := testEntitlement(team, teamAdminEntitlement)
	b := newTeamBuilder(c, "replacement")

	if _, _, err := b.Grant(context.Background(), user, adminEnt); err == nil {
		t.Error("expected Grant on admin entitlement to fail")
	}
	if _, err := b.Revoke(context.Background(), &v2.Grant{Entitlement: adminEnt, Principal: user}); err == nil {
		t.Error("expected Revoke on admin entitlement to fail")
	}
	if n := len(f.all()); n != 0 {
		t.Errorf("expected no HTTP calls, got %d", n)
	}
}

func TestTeamRevoke(t *testing.T) {
	removePath := membersPath + "/alice"
	f, c := newFakeVictorOps(t, map[string]http.HandlerFunc{
		"DELETE " + removePath: respond(http.StatusOK, `{}`),
	})
	team := testTeamResource(t)
	g := &v2.Grant{Entitlement: testEntitlement(team, teamMemberEntitlement), Principal: testUserResource(t, "alice")}

	if _, err := newTeamBuilder(c, "oncall-lead").Revoke(context.Background(), g); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reqs := f.all()
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	if reqs[0].method != http.MethodDelete || reqs[0].path != removePath {
		t.Errorf("unexpected request %s %s", reqs[0].method, reqs[0].path)
	}
	if strings.TrimSpace(reqs[0].body) != `{"replacement":"oncall-lead"}` {
		t.Errorf("unexpected body %q", reqs[0].body)
	}
}

func TestTeamRevokeKeepsVendorErrorBody(t *testing.T) {
	_, c := newFakeVictorOps(t, map[string]http.HandlerFunc{
		"DELETE " + membersPath + "/alice": respond(http.StatusBadRequest, `{"error":"unexpected payload"}`),
	})
	team := testTeamResource(t)
	g := &v2.Grant{Entitlement: testEntitlement(team, teamMemberEntitlement), Principal: testUserResource(t, "alice")}

	_, err := newTeamBuilder(c, "oncall-lead").Revoke(context.Background(), g)
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, client.ErrInvalidReplacement) {
		t.Errorf("400 should not be reported as an invalid replacement: %v", err)
	}
	if !strings.Contains(err.Error(), `unexpected payload`) {
		t.Errorf("expected error to include the raw body, got %v", err)
	}
}

func TestTeamRevokePreconditions(t *testing.T) {
	tests := []struct {
		name        string
		replacement string
	}{
		{name: "replacement not configured", replacement: ""},
		{name: "replacement is the removed user", replacement: "Alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFakeVictorOps(t, nil)
			team := testTeamResource(t)
			g := &v2.Grant{Entitlement: testEntitlement(team, teamMemberEntitlement), Principal: testUserResource(t, "alice")}

			_, err := newTeamBuilder(c, tt.replacement).Revoke(context.Background(), g)
			if status.Code(err) != codes.FailedPrecondition {
				t.Errorf("expected FailedPrecondition, got %v", err)
			}
			if n := len(f.all()); n != 0 {
				t.Errorf("expected no HTTP calls, got %d", n)
			}
		})
	}
}

func TestTeamRevokeInvalidReplacement(t *testing.T) {
	_, c := newFakeVictorOps(t, map[string]http.HandlerFunc{
		"DELETE " + membersPath + "/alice": respond(http.StatusUnprocessableEntity, `{"message":"replacement not found"}`),
	})
	team := testTeamResource(t)
	g := &v2.Grant{Entitlement: testEntitlement(team, teamMemberEntitlement), Principal: testUserResource(t, "alice")}

	_, err := newTeamBuilder(c, "ghost").Revoke(context.Background(), g)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, client.ErrInvalidReplacement) {
		t.Errorf("expected ErrInvalidReplacement, got %v", err)
	}
	if !strings.Contains(err.Error(), "removal-replacement-user ghost exists in VictorOps") {
		t.Errorf("expected error to mention the replacement user and team, got %v", err)
	}
	if !strings.Contains(err.Error(), "replacement not found") {
		t.Errorf("expected error to include the vendor message, got %v", err)
	}
}
