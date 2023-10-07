package util

func VersionInc(cv int) int {
	nv := cv + 1
	if nv > 10000 {
		// reset version each 10K updates
		return 0
	}
	return nv
}
