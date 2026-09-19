package model

import "testing"

func TestUserRoleNormalize(t *testing.T) {
	cases := map[UserRole]UserRole{
		"user":            RolePathologist,
		"viewer":          RoleDatascientist,
		RoleAdmin:         RoleAdmin,
		RolePathologist:   RolePathologist,
		RoleDatascientist: RoleDatascientist,
		RoleUnassigned:    RoleUnassigned,
		"":                "",
		"something-else":  "something-else",
	}
	for in, want := range cases {
		if got := in.Normalize(); got != want {
			t.Errorf("UserRole(%q).Normalize() = %q, want %q", in, got, want)
		}
	}
}

func TestUserRoleIsAssignable(t *testing.T) {
	cases := map[UserRole]bool{
		RoleAdmin:         true,
		RolePathologist:   true,
		RoleDatascientist: true,
		RoleUnassigned:    false,
		"user":            false,
		"viewer":          false,
		"":                false,
	}
	for role, want := range cases {
		if got := role.IsAssignable(); got != want {
			t.Errorf("UserRole(%q).IsAssignable() = %v, want %v", role, got, want)
		}
	}
}
