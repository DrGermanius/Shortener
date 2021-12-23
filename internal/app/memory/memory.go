package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/models"
)

type LinkMemoryStore map[string]models.LinkInfo

func NewLinkMemoryStore() (*LinkMemoryStore, error) {
	LinksMap := make(LinkMemoryStore)

	err := LinksMap.readFile()
	if err != nil {
		return nil, err
	}
	return &LinksMap, nil
}

func (l *LinkMemoryStore) Ping(ctx context.Context) bool {
	_ = ctx
	return true //todo
}

func (l *LinkMemoryStore) Get(s string) (string, bool) {
	long, exist := (*l)[s]
	return long.Long, exist
}

func (l *LinkMemoryStore) GetByUserID(id string) []models.LinkJSON {
	var res []models.LinkJSON
	for k, v := range *l {
		if v.UUID == id {
			res = append(res, models.LinkJSON{Long: v.Long, Short: config.Config().BaseURL + "/" + k}) //todo config.Config().BaseURL + "/"
		}
	}

	return res
}

func (l *LinkMemoryStore) Write(uuid, long string) (string, error) {
	s := app.ShortLink([]byte(long))
	(*l)[s] = models.LinkInfo{Long: long, UUID: uuid}

	err := writeFile(uuid, s, long)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (l *LinkMemoryStore) readFile() error {
	p := config.Config().FilePath

	f, err := os.OpenFile(p, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)

	for s.Scan() {
		var link models.LinkJSON
		err = json.Unmarshal(s.Bytes(), &link)
		if err != nil {
			return err
		}

		(*l)[link.Short] = models.LinkInfo{Long: link.Long, UUID: link.UUID}
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

func writeFile(uuid, short, long string) error {
	m := models.LinkJSON{
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
