package teams

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func preferencePath() (string, error) {
	d, e := os.UserConfigDir()
	return filepath.Join(d, "hupu", "preferred-team"), e
}
func Preferred() (string, error) {
	p, e := preferencePath()
	if e != nil {
		return "", e
	}
	b, e := os.ReadFile(p)
	if os.IsNotExist(e) {
		return "sixers", nil
	}
	if e != nil {
		return "", e
	}
	key, ok := ResolveTeam(strings.TrimSpace(string(b)))
	if !ok || key == "nba" {
		return "sixers", nil
	}
	return key, nil
}
func Save(key string) error {
	key, ok := ResolveTeam(key)
	if !ok || key == "nba" {
		return fmt.Errorf("请选择一支 NBA 球队")
	}
	p, e := preferencePath()
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(p), ".team-*")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.WriteString(key); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(f.Name(), p)
}
func Name(key string) string {
	for _, t := range Teams {
		if t.Key == key {
			return t.Name
		}
	}
	return key
}
