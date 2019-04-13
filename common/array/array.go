package array

// InArrayString return true if  find it in  array or return false
func InArrayString(array []string, substr string) bool {
	for _, s := range array {
		if s == substr {
			return true
		}
	}
	return false
}

// InArrayInt return true if  find it in  array or return false
func InArrayInt(array []int, substr int) bool {
	for _, s := range array {
		if s == substr {
			return true
		}
	}
	return false
}

// InArrayInt return true if  find it in  array or return false
func InArrayFloat(array []float64, substr float64) bool {
	for _, s := range array {
		if s == substr {
			return true
		}
	}
	return false
}
