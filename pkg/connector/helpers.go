package connector

import (
	"strings"

	"github.com/conductorone/baton-victorops/pkg/connector/client"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/types/known/structpb"
)

func titleCase(s string) string {
	return cases.Title(language.English).String(strings.ToLower(s))
}

func getProfileStringArray(profile *structpb.Struct, k string) ([]string, bool) {
	var values []string
	if profile == nil {
		return nil, false
	}

	v, ok := profile.Fields[k]
	if !ok {
		return nil, false
	}

	s, ok := v.Kind.(*structpb.Value_ListValue)
	if !ok {
		return nil, false
	}

	for _, v := range s.ListValue.Values {
		if strVal := v.GetStringValue(); strVal != "" {
			values = append(values, strVal)
		}
	}

	return values, true
}

func scheduleMembersToInterfaceSlice(users []client.OnCallUser) []interface{} {
	var i []interface{}
	for _, u := range users {
		i = append(i, u.OnCallUser.Username)
	}
	return i
}
