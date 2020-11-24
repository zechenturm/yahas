package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var chars = [...]string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f",
}

func formatHexDigit(b byte) string {
	lowNibble := b & 0xF
	highNibble := b >> 4
	return chars[highNibble] + chars[lowNibble]
}

func formatHex(bytes []byte) string {
	s := ""
	for _, b := range bytes {
		s += formatHexDigit(b)
	}
	return s
}

func createUser(name, password string) error {
	pwHash, err := hashPW(password)

	userFile, err := os.OpenFile("users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(userFile, "%s\t%s\n", name, formatHex(pwHash))
	return err
}

func toNibble(letter byte) byte {
	if letter > '9' {
		return letter - 'a' + 10
	}
	return letter - '0'
}

func fromString(hashString string) []byte {
	hash := make([]byte, len(hashString)/2)
	for i := range hash {
		si := 2 * i
		hash[i] = (toNibble(hashString[si]) << 4) | toNibble(hashString[si+1])
	}
	return hash
}

func fromLine(line string) (string, []byte) {
	fields := strings.Split(line, "\t")
	return fields[0], fromString(fields[1])
}

func readUsers() (map[string][]byte, error) {
	bytes, err := ioutil.ReadFile("config/users.txt")
	if err != nil {
		return nil, err
	}
	text := string(bytes)
	lines := strings.Split(text, "\n")

	users := make(map[string][]byte)
	for _, line := range lines[:len(lines)-1] {
		name, pw := fromLine(line)
		users[name] = pw
	}
	return users, nil
}

func hashPW(password string) ([]byte, error) {
	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, err
	}
	return pwHash, nil
}

func checkUser(users map[string][]byte, name, password string) bool {
	if pw, ok := users[name]; ok {
		return bcrypt.CompareHashAndPassword(pw, []byte(password)) == nil
	}
	return false
}
