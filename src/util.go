package main

func stringInSlice(arr []string, val string) bool {
	for _, s := range arr {
		if s == val {
			return true
		}
	}

	return false
}
