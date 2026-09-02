package distsig

func equalPoints(a, b []Point) bool {
	if len(a) != len(b) {
		return false
	}
	for i, p := range a {
		if !p.Equal(b[i]) {
			return false
		}
	}
	return true
}
