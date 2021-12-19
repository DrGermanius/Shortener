package store

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
)

type Links map[string]string

var LinksMap Links

func NewLinksMap() error {
	LinksMap = make(map[string]string)
	err := LinksMap.readFile()
	if err != nil {
		return err
	}
	return nil
}

func Clear() error {
	f := config.Config().FilePath
	err := os.Remove(f)
	if err != nil {
		return err
	}
	return nil
}

func (l *Links) Write(long string) (string, error) {
	s := app.ShortLink([]byte(long))
	(*l)[s] = long

	err := writeFile(s, long)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (l *Links) readFile() error {
	p := config.Config().FilePath

	f, err := os.OpenFile(p, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)

	for s.Scan() {
		var link link
		err = json.Unmarshal(s.Bytes(), &link)
		if err != nil {
			return err
		}

		(*l)[link.Short] = link.Long
	}
	return nil
}

func writeFile(short, long string) error {
	m := link{
		Short: short,
		Long:  long,
	}

	p := config.Config().FilePath

	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	w := bufio.NewWriter(f)

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

type link struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}
