package database

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/models"
)

const (
	linkFields      = "user_id, long_link, short_link"
	insertLinkQuery = "INSERT INTO links  (" + linkFields + ") VALUES ( $1, $2, $3 )"
)

type DB struct {
	conn *pgxpool.Pool
}

func NewDatabaseStore(connString string) (*DB, error) {
	conn, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	err = createDatabaseAndTable(conn)
	if err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func (d *DB) Get(ctx context.Context, short string) (string, error) {
	var long string
	var isDel bool

	row := d.conn.QueryRow(ctx, "SELECT long_link, is_deleted FROM links where short_link = $1", short)

	err := row.Scan(&long, &isDel)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", app.ErrLinkNotFound
	}
	if err != nil {
		return "", err
	}

	if isDel {
		return "", app.ErrDeletedLink
	}

	return long, nil
}

func (d *DB) GetByUserID(ctx context.Context, id string) ([]models.LinkJSON, error) {
	var links []models.LinkJSON

	rows, err := d.conn.Query(ctx, "SELECT "+linkFields+" FROM links where user_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var l models.LinkJSON
		err = rows.Scan(&l.UUID, &l.Long, &l.Short)
		l.Short = app.FullLink(l.Short)

		if err != nil {
			return nil, err
		}

		links = append(links, l)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(links) == 0 {
		return nil, app.ErrUserHasNoRecords
	}

	return links, nil
}

func (d *DB) Write(ctx context.Context, uuid, long string) (string, error) {
	short := app.ShortLink([]byte(long))

	_, err := d.conn.Exec(ctx, insertLinkQuery, uuid, long, short)

	if err != nil {
		pgErr := new(pgconn.PgError)
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return short, app.ErrLinkAlreadyExists
		}
		return "", err
	}

	return short, nil
}

func (d *DB) Delete(ctx context.Context, uid string, link string) error {
	//_, err := d.conn.Exec(ctx, "UPDATE links SET is_deleted = true WHERE user_id = $1 AND short_link = any($2)", uid, link)

	_, err := d.conn.Exec(ctx, "UPDATE links SET is_deleted = true WHERE user_id = $1 AND short_link = $2", uid, link)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) BatchWrite(ctx context.Context, uid string, originals []models.BatchOriginal) ([]string, error) {
	conn, err := d.conn.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	shorts := make([]string, 0, len(originals))
	for _, v := range originals {
		shorts = append(shorts, app.ShortLink([]byte(v.OriginalURL)))
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Prepare(ctx, "batch-insert", insertLinkQuery)
	if err != nil {
		return nil, err
	}

	for i, v := range originals {
		rows, err := tx.Query(ctx, "batch-insert", uid, v.OriginalURL, shorts[i])
		if err != nil {
			return nil, err
		}
		rows.Close()
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return shorts, nil
}

func (d *DB) Ping(ctx context.Context) bool {
	return d.conn.Ping(ctx) == nil
}

func createDatabaseAndTable(c *pgxpool.Pool) error {
	rows, err := c.Query(context.Background(), "SELECT 1 FROM pg_catalog.pg_database WHERE datname = 'links'")
	if err != nil {
		return err
	}

	if !rows.Next() {
		_, err = c.Exec(context.Background(), "CREATE DATABASE links")
		if err != nil {
			return err
		}

		_, err = c.Exec(context.Background(), "CREATE TABLE links ("+
			"id    		SERIAL 			PRIMARY KEY,"+
			"user_id    VARCHAR ( 50 )  NOT NULL,"+
			"long_link  VARCHAR  		NOT NULL,"+
			"short_link VARCHAR  		NOT NULL,"+
			"is_deleted bool DEFAULT false  	NOT NULL,"+
			"UNIQUE(long_link)"+
			");")
		if err != nil {
			return err
		}
	}

	return nil
}
