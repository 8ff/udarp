package bitManipulation

// Function that converts string of 1/0s to []int
func StringToIntSlice(s string) []int {
	a := make([]int, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '0' {
			a[i] = 0
		} else {
			a[i] = 1
		}
	}
	return a
}

// Function that converts []int to string of 1/0s
func IntSliceToString(a []int) string {
	s := ""
	for i := 0; i < len(a); i++ {
		if a[i] == 0 {
			s += "0"
		} else {
			s += "1"
		}
	}
	return s
}
