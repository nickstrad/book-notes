package store

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
)

type Store struct{ DB *sql.DB }

func Open(url string) (*Store, error) {
	d, e := sql.Open("postgres", url)
	if e != nil {
		return nil, e
	}
	d.SetMaxOpenConns(16)
	return &Store{d}, d.Ping()
}
func (s *Store) Rows(query string, args ...any) ([]map[string]any, error) {
	r, e := s.DB.Query(query, args...)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	cols, e := r.Columns()
	if e != nil {
		return nil, e
	}
	out := []map[string]any{}
	for r.Next() {
		vals := make([]any, len(cols))
		ptr := make([]any, len(cols))
		for i := range vals {
			ptr[i] = &vals[i]
		}
		if e = r.Scan(ptr...); e != nil {
			return nil, e
		}
		m := map[string]any{}
		for i, k := range cols {
			v := vals[i]
			if b, ok := v.([]byte); ok {
				if k == "payload" || k == "result" || k == "body" {
					var j any
					if json.Unmarshal(b, &j) == nil {
						v = j
					} else {
						v = string(b)
					}
				} else {
					v = string(b)
				}
			}
			m[k] = v
		}
		out = append(out, m)
	}
	return out, r.Err()
}
func (s *Store) One(query string, args ...any) (map[string]any, error) {
	r, e := s.Rows(query, args...)
	if e != nil {
		return nil, e
	}
	if len(r) == 0 {
		return nil, sql.ErrNoRows
	}
	return r[0], nil
}
