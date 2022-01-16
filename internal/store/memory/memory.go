package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/DrGermanius/Shortener/internal/app/util"

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

func (l *LinkMemoryStore) BatchWrite(ctx context.Context, uid string, originals []models.BatchOriginal) ([]string, error) {
	_ = ctx
	shorts := make([]string, 0, len(originals))
	for _, v := range originals {
		s, err := l.Write(ctx, uid, v.OriginalURL)
		if err != nil {
			return nil, err
		}

		shorts = append(shorts, s)
	}
	return shorts, nil
}

func (l *LinkMemoryStore) Ping(ctx context.Context) bool {
	_ = ctx
	return true
}

func (l *LinkMemoryStore) BatchDelete(ctx context.Context, uid string, links []string) error {
	_ = ctx

	for _, link := range links {
		if (*l)[link].UUID == uid {
			updatedLink := (*l)[link]
			updatedLink.IsDeleted = true

			(*l)[link] = updatedLink
		}
	}
	return nil
}

func (l *LinkMemoryStore) Get(ctx context.Context, s string) (string, error) {
	_ = ctx
	long, exist := (*l)[s]
	if !exist {
		return "", app.ErrLinkNotFound
	}

	if long.IsDeleted {
		return "", app.ErrDeletedLink
	}
	return long.Long, nil
}

func (l *LinkMemoryStore) GetByUserID(ctx context.Context, id string) ([]models.LinkJSON, error) {
	_ = ctx
	var res []models.LinkJSON
	for k, v := range *l {
		if v.UUID == id {
			res = append(res, models.LinkJSON{Long: v.Long, Short: util.FullLink(k)})
		}
	}

	if len(res) == 0 {
		return nil, app.ErrUserHasNoRecords
	}

	return res, nil
}

func (l *LinkMemoryStore) Write(ctx context.Context, uuid, long string) (string, error) {
	_ = ctx
	s := app.ShortLink([]byte(long))
	(*l)[s] = models.LinkInfo{Long: long, UUID: uuid, IsDeleted: false}

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

		(*l)[link.Short] = models.LinkInfo{Long: link.Long, UUID: link.UUID, IsDeleted: link.IsDeleted}
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
		UUID:      uuid,
		Short:     short,
		Long:      long,
		IsDeleted: false,
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
