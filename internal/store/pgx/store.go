package pgx

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/models"
	"go.elastic.co/apm/module/apmpgx/v2"
)

type PgxStore struct {
	pool *pgxpool.Pool
}

func (d PgxStore) Pool() *pgxpool.Pool {
	return d.pool
}

func (d *PgxStore) Close() {
	d.pool.Close()
}

func Init() *PgxStore {
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable connect_timeout=5", config.Conf.DbUsername, config.Conf.DbDatabase, config.Conf.DbPassword, config.Conf.DbHost, config.Conf.DbPort)

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal(err)
	}

	apmpgx.Instrument(cfg.ConnConfig)
	pool, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &PgxStore{pool: pool}
}

func parseColumnsForScan(sub interface{}, addColumns ...interface{}) []interface{} {
	s := reflect.ValueOf(sub).Elem()
	numCols := s.NumField() - len(sub.(models.HasRelationFields).RelationFields())

	columns := []interface{}{}
	for i := 0; i < numCols; i++ {
		field := s.Field(i)
		columns = append(columns, field.Addr().Interface())
	}
	columns = append(columns, addColumns...)
	return columns
}

type pgxWithTx func(tx pgx.Tx) (rollback bool, err error)
type pgxQuery func(conn *pgxpool.Conn) (err error)

func (d *PgxStore) runQuery(ctx context.Context, f pgxQuery) (err error) {
	p := d.Pool()
	if config.Conf.AppIsReadonly != nil {
		if *config.Conf.AppIsReadonly {

			_, err := p.Exec(context.Background(), "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY;")
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err := p.Exec(context.Background(), "SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE;")
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	err = p.AcquireFunc(ctx, f)
	if err != nil {
		return err
	}
	return
}

func (d *PgxStore) runInTx(ctx context.Context, f pgxWithTx) (err error) {
	var conn *pgxpool.Conn

	defer func() {
		if conn != nil {
			conn.Release()
		}
	}()
	conn, err = d.Pool().Acquire(ctx)
	if err != nil {
		return err
	}

	rollback := true
	var tx pgx.Tx
	defer func() {
		if rollback && tx != nil {
			rErr := tx.Rollback(ctx)
			if rErr != nil && err != pgx.ErrTxClosed {
				log.Println("Rolling back: " + rErr.Error())
			}
		}
	}()
	tx, err = conn.Begin(ctx)
	if err != nil {
		return err
	}
	rollback, err = f(tx)
	if err != nil {
		return err
	}
	if !rollback {
		err = tx.Commit(ctx)
		if err != nil {
			return err
		}
	}
	return
}
