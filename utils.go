package ddpserver

import (
	"math/rand"
)

var dict = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateID(n int) string {
	a := make([]rune, n)
	l := len(dict)

	for i := range a {
		a[i] = dict[rand.Intn(l)]
	}

	return string(a)
}
