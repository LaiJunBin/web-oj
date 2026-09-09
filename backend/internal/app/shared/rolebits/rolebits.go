package rolebits

// Named types prevent fx from confusing these with other int parameters.
// Each mask is computed once at composition root via roleconverter.ToBit.
type AdminRoleMask int
type ContestantRoleMask int
type APIReaderRoleMask int

// RoleMasks holds all role bitmasks and is injected as a single fx value.
type RoleMasks struct {
	Admin      AdminRoleMask
	Contestant ContestantRoleMask
	APIReader  APIReaderRoleMask
}

// HasRole returns true if rolesBit includes the given role mask.
// Accepts any named type with underlying type int (all *RoleMask types qualify).
func HasRole[M ~int](rolesBit int, m M) bool {
	return rolesBit&int(m) != 0
}
