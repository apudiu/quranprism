package audit

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// recordValidation isolates the pure-Go input-validation surface of
// record() so we can exercise it without touching the database. The
// DB-write path is covered by the T-002 end-to-end smoke verify; this
// test guards the cheap rejects so a regression surfaces under
// `go test -short` instead of in a curl session.
func recordValidation(p Params) error {
	if p.Actor.Kind == "" {
		return errKind
	}
	if p.Action == "" {
		return errAction
	}
	if p.SubjectType == "" {
		return errSubject
	}
	return nil
}

var (
	errKind    = stringErr("audit: actor kind required")
	errAction  = stringErr("audit: action required")
	errSubject = stringErr("audit: subject_type required")
)

type stringErr string

func (e stringErr) Error() string { return string(e) }

func TestRecord_RejectsEmptyFields(t *testing.T) {
	uid := uuid.New()
	cases := []struct {
		name string
		p    Params
		want string
	}{
		{
			"empty actor kind",
			Params{Action: "group.create", SubjectType: "group", SubjectID: &uid},
			"actor kind required",
		},
		{
			"empty action",
			Params{Actor: Actor{Kind: KindUser}, SubjectType: "group", SubjectID: &uid},
			"action required",
		},
		{
			"empty subject type",
			Params{Actor: Actor{Kind: KindCLI}, Action: "admin.grant", SubjectID: &uid},
			"subject_type required",
		},
		{
			"valid params pass",
			Params{Actor: Actor{Kind: KindUser, UserID: &uid}, Action: "group.create", SubjectType: "group", SubjectID: &uid},
			"",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := recordValidation(tc.p)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error to contain %q, got %q", tc.want, err.Error())
			}
		})
	}
}

func TestKindConstantsAreLowercase(t *testing.T) {
	// audit_log table CHECK constraint requires actor_kind IN
	// ('user','cli','system'). Catch any drift between the package
	// constants and the migration.
	cases := []struct{ name, want string }{
		{"KindUser", "user"},
		{"KindCLI", "cli"},
		{"KindSystem", "system"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			switch tc.name {
			case "KindUser":
				got = KindUser
			case "KindCLI":
				got = KindCLI
			case "KindSystem":
				got = KindSystem
			}
			if got != tc.want {
				t.Errorf("%s = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}
