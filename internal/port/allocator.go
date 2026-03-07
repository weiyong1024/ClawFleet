package port

import (
	"fmt"
	"net"
)

// FindAvailable returns the first available TCP port >= start that is not in used.
func FindAvailable(start int, used map[int]bool) (int, error) {
	for p := start; p < start+1000; p++ {
		if used[p] {
			continue
		}
		if isAvailable(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range [%d, %d)", start, start+1000)
}

func isAvailable(p int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
