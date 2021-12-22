package store

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
)

type Links map[string]Info

type Info struct {
	Long string
	UUID string
}

var LinksMap Links

func NewLinksMap() error {
	LinksMap = make(map[string]Info)
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

func (l *Links) GetByUserID(id string) []link {
	var res []link
	for k, v := range LinksMap {
		if v.UUID == id {
			res = append(res, link{Long: v.Long, Short: config.Config().BaseURL + "/" + k}) //todo config.Config().BaseURL + "/"
		}
	}

	return res
}

func (l *Links) Write(uuid, long string) (string, error) {
	s := app.ShortLink([]byte(long))
	(*l)[s] = Info{Long: long, UUID: uuid}

	err := writeFile(uuid, s, long)
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

		(*l)[link.Short] = Info{link.Long, link.UUID}
	}
	return nil
}

func writeFile(uuid, short, long string) error {
	m := link{
		UUID:  uuid,
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
	UUID  string `json:"uuid,omitempty"`
	Short string `json:"short_url"`
	Long  string `json:"original_url"`
}
