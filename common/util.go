package common

import "fmt"

func PrettySize(b uint64) string {
	gb := float64(b) / float64(1024*1024*1024)
	if gb < 0.00009 {
		return fmt.Sprintf("%.4fMB", gb*1024)
	}
	return fmt.Sprintf("%.4fGB", gb)
}
