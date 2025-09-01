package main

// Helper functions to safely dereference pointers
func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func derefInt(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

func derefBool(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}
