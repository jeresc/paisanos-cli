package utils

import (
	"os/user"
)

func Max(n, m int) int {
	if m > n {
		return m
	}
	return n
}

func GetCurrentUser() (*user.User, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	return usr, nil
}
