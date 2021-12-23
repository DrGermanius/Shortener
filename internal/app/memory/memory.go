package memory

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/models"
)

type MemoryStore map[string]models.LinkInfo

func NewLinkMemoryStore() (*MemoryStore, error) {
	LinksMap := make(MemoryStore)

	err := LinksMap.readFile()
	if err != nil {
		return nil, err
	}
	return &LinksMap, nil
}

func (l *MemoryStore) Ping() bool {
	return true //todo
}

func (l *MemoryStore) Get(s string) (string, bool) {
	long, exist := (*l)[s]
	return long.Long, exist
}

func (l *MemoryStore) GetByUserID(id string) []models.LinkJSON {
	var res []models.LinkJSON
	for k, v := range *l {
		if v.UUID == id {
			res = append(res, models.LinkJSON{Long: v.Long, Short: config.Config().BaseURL + "/" + k}) //todo config.Config().BaseURL + "/"
		}
	}

	return res
}

func (l *MemoryStore) Write(uuid, long string) (string, error) {
	s := app.ShortLink([]byte(long))
	(*l)[s] = models.LinkInfo{Long: long, UUID: uuid}

	err := writeFile(uuid, s, long)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (l *MemoryStore) readFile() error {
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
