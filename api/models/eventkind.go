package models

type EventKind int

const (
	NewVulnerability EventKind = 1 << iota // 1 << 0 which is 1
	NewComment                             // 1 << 1 which is 2
)

// EventKinds returns a slice of EventKind values based on the input bitmask.
func EventKinds(bitmask int) []EventKind {
	var kinds []EventKind

	// Check each event kind
	if bitmask&int(NewVulnerability) != 0 {
		kinds = append(kinds, NewVulnerability)
	}
	if bitmask&int(NewComment) != 0 {
		kinds = append(kinds, NewComment)
	}

	return kinds
}
