package main

func seed(s *State) {
	// Users — diversity matters: 5 users, 1 unverified, 1 with no team assignments.
	users := []*User{
		{
			Username:  "alice@example.com",
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     "alice@example.com",
			CreatedAt: "2024-01-01T00:00:00Z",
			Verified:  true,
			SelfUrl:   "http://localhost:8765/api-public/v1/user/alice@example.com",
		},
		{
			// bob is in two teams — tests overlapping membership
			Username:  "bob@example.com",
			FirstName: "Bob",
			LastName:  "Jones",
			Email:     "bob@example.com",
			CreatedAt: "2024-01-02T00:00:00Z",
			Verified:  true,
			SelfUrl:   "http://localhost:8765/api-public/v1/user/bob@example.com",
		},
		{
			// carol is unverified — tests the Verified=false path in sync
			Username:  "carol@example.com",
			FirstName: "Carol",
			LastName:  "Davis",
			Email:     "carol@example.com",
			CreatedAt: "2024-01-03T00:00:00Z",
			Verified:  false,
			SelfUrl:   "http://localhost:8765/api-public/v1/user/carol@example.com",
		},
		{
			// dave has no team assignments — tests the empty-grants path
			Username:  "dave@example.com",
			FirstName: "Dave",
			LastName:  "Brown",
			Email:     "dave@example.com",
			CreatedAt: "2024-01-04T00:00:00Z",
			Verified:  true,
			SelfUrl:   "http://localhost:8765/api-public/v1/user/dave@example.com",
		},
		{
			Username:  "eve@example.com",
			FirstName: "Eve",
			LastName:  "Wilson",
			Email:     "eve@example.com",
			CreatedAt: "2024-01-05T00:00:00Z",
			Verified:  true,
			SelfUrl:   "http://localhost:8765/api-public/v1/user/eve@example.com",
		},
	}
	for _, u := range users {
		cp := *u
		s.users[u.Username] = &cp
		s.userList = append(s.userList, &cp)
	}

	// Teams — 3 teams so multi-team grant emission is tested.
	teams := []*Team{
		{
			Slug:          "team-alpha",
			Name:          "Team Alpha",
			Description:   "Primary on-call team",
			MemberCount:   2,
			Version:       1,
			IsDefaultTeam: true,
			SelfUrl:       "http://localhost:8765/api-public/v1/team/team-alpha",
			MembersUrl:    "http://localhost:8765/api-public/v1/team/team-alpha/members",
			AdminsUrl:     "http://localhost:8765/api-public/v1/team/team-alpha/admins",
		},
		{
			// team-beta shares bob with team-alpha — tests overlapping membership
			Slug:        "team-beta",
			Name:        "Team Beta",
			Description: "Secondary on-call team",
			MemberCount: 2,
			Version:     1,
			SelfUrl:     "http://localhost:8765/api-public/v1/team/team-beta",
			MembersUrl:  "http://localhost:8765/api-public/v1/team/team-beta/members",
			AdminsUrl:   "http://localhost:8765/api-public/v1/team/team-beta/admins",
		},
		{
			Slug:        "team-ops",
			Name:        "Team Ops",
			Description: "Operations team",
			MemberCount: 1,
			Version:     1,
			SelfUrl:     "http://localhost:8765/api-public/v1/team/team-ops",
			MembersUrl:  "http://localhost:8765/api-public/v1/team/team-ops/members",
			AdminsUrl:   "http://localhost:8765/api-public/v1/team/team-ops/admins",
		},
	}
	for _, t := range teams {
		cp := *t
		s.teams[t.Slug] = &cp
		s.teamList = append(s.teamList, &cp)
	}

	// Team memberships — alice+bob in team-alpha, bob+carol in team-beta (bob in two teams),
	// eve in team-ops; dave has no memberships.
	s.members["team-alpha"] = []string{"alice@example.com", "bob@example.com"}
	s.members["team-beta"] = []string{"bob@example.com", "carol@example.com"}
	s.members["team-ops"] = []string{"eve@example.com"}

	// Team admins — admin entitlement is immutable so this slice is never mutated after seed.
	s.admins["team-alpha"] = []string{"alice@example.com"}
	s.admins["team-beta"] = []string{"bob@example.com"}
	s.admins["team-ops"] = []string{"eve@example.com"}

	// On-call schedules — alice on-call for team-alpha, bob for team-beta;
	// team-ops has an empty OnCallNow to exercise the no-schedule path.
	s.teamsOnCall = []TeamOnCall{
		{
			Team: *s.teams["team-alpha"],
			OnCallNow: []OnCallInfo{
				{
					EscalationPolicy: Policy{Name: "Primary Escalation", Slug: "primary-escalation"},
					Users: []OnCallUser{
						{OnCallUser: OnCallUserInfo{Username: "alice@example.com"}},
					},
				},
			},
		},
		{
			Team: *s.teams["team-beta"],
			OnCallNow: []OnCallInfo{
				{
					EscalationPolicy: Policy{Name: "Secondary Escalation", Slug: "secondary-escalation"},
					Users: []OnCallUser{
						{OnCallUser: OnCallUserInfo{Username: "bob@example.com"}},
					},
				},
			},
		},
		{
			// team-ops has no current on-call — scheduleBuilder.List must handle empty OnCallNow
			Team:      *s.teams["team-ops"],
			OnCallNow: []OnCallInfo{},
		},
	}
}
