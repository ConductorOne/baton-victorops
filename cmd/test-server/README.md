# Test server

Mock VictorOps API for baton-victorops, used when no real-tenant credentials are available.
Replicates the upstream API's header auth, endpoints, and error envelopes so the connector
exercises every sync, provisioning, and grant/revoke code path in CI.

## Auth

| Real API | Test server |
|---|---|
| `X-VO-Api-Id` header with the API ID | Same header; hardcoded value: `test-api-id` |
| `X-VO-Api-Key` header with the API key | Same header; hardcoded value: `test-api-key` |

Requests with missing or wrong headers get HTTP 401.

## Endpoints

| Path | Method | Doc URL |
|---|---|---|
| `/api-public/v1/user` | GET | https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_user |
| `/api-public/v1/team` | GET | https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team |
| `/api-public/v1/team/{slug}/members` | GET | https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team_team_members |
| `/api-public/v1/team/{slug}/members` | POST | https://portal.victorops.com/public/api-doc.html#!/API_Public/post_api_public_v1_team_team_members |
| `/api-public/v1/team/{slug}/members/{username}` | DELETE | https://portal.victorops.com/public/api-doc.html#!/API_Public/delete_api_public_v1_team_team_members_user |
| `/api-public/v1/team/{slug}/admins` | GET | https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_team_team_admins |
| `/api-public/v1/oncall/current` | GET | https://portal.victorops.com/public/api-doc.html#!/API_Public/get_api_public_v1_oncall_current |

> **Known connector bug:** `ListTeamAdmins` calls the `/members` endpoint instead of `/admins`
> (wrong constant). The test server correctly serves `/admins` per the docs — so admin grants
> will come back empty from the connector until the bug is fixed.

## Seed data

See `seeds.go` for the canonical list. Summary:

| Resource | Count | Notes |
|---|---|---|
| Users | 5 | 1 unverified (carol), 1 with no team assignments (dave) |
| Teams | 3 | team-alpha, team-beta, team-ops |
| Team members | 5 memberships | bob is in team-alpha and team-beta (overlap) |
| Team admins | 3 | one admin per team (immutable entitlement) |
| On-call schedules | 2 | team-alpha primary (alice), team-beta secondary (bob) |
| On-call — empty | 1 | team-ops has no current on-call (exercises empty OnCallNow path) |

## Running locally

```bash
# Start the test server (from the project root)
go run ./cmd/test-server/

# In a separate terminal, run the connector against it
./baton-victorops \
  --victorops-api-id=test-api-id \
  --victorops-api-key=test-api-key \
  --base-url=http://localhost:8765

# Inspect the sync output
baton resources --file=sync.c1z
baton grants   --file=sync.c1z
```

## Curl examples

```bash
# List users
curl -s http://localhost:8765/api-public/v1/user \
  -H 'X-VO-Api-Id: test-api-id' \
  -H 'X-VO-Api-Key: test-api-key' | jq .

# List teams
curl -s http://localhost:8765/api-public/v1/team \
  -H 'X-VO-Api-Id: test-api-id' \
  -H 'X-VO-Api-Key: test-api-key' | jq .

# List team members
curl -s http://localhost:8765/api-public/v1/team/team-alpha/members \
  -H 'X-VO-Api-Id: test-api-id' \
  -H 'X-VO-Api-Key: test-api-key' | jq .

# List team admins
curl -s http://localhost:8765/api-public/v1/team/team-alpha/admins \
  -H 'X-VO-Api-Id: test-api-id' \
  -H 'X-VO-Api-Key: test-api-key' | jq .

# Add a user to a team (provisioning)
curl -s -X POST http://localhost:8765/api-public/v1/team/team-ops/members \
  -H 'X-VO-Api-Id: test-api-id' \
  -H 'X-VO-Api-Key: test-api-key' \
  -H 'Content-Type: application/json' \
  -d '{"username":"dave@example.com"}' | jq .

# Remove a user from a team (revoke)
curl -s -X DELETE http://localhost:8765/api-public/v1/team/team-ops/members/dave@example.com \
  -H 'X-VO-Api-Id: test-api-id' \
  -H 'X-VO-Api-Key: test-api-key' \
  -H 'Content-Type: application/json' \
  -d '{}' | jq .

# On-call current
curl -s http://localhost:8765/api-public/v1/oncall/current \
  -H 'X-VO-Api-Id: test-api-id' \
  -H 'X-VO-Api-Key: test-api-key' | jq .

# Missing auth headers → 401
curl -s http://localhost:8765/api-public/v1/user | jq .
```
