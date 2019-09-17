package pgmodel

import (
	"context"
	"fmt"
	"reflect"

	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgstmt"
)

func Do(ctx context.Context, model interface{}, filter ...Filter) error {
	switch m := model.(type) {
	case Selector:
		return m.Scan(pgstmt.Select(func(b pgstmt.SelectStatement) {
			m.Select(b)
			for _, f := range filter {
				f.Apply(ctx, b)
			}
		}).QueryRowWith(ctx).Scan)
	case Inserter:
		stmt := pgstmt.Insert(func(b pgstmt.InsertStatement) {
			m.Insert(b)
		})

		if scanner, ok := m.(Scanner); ok {
			return scanner.Scan(stmt.QueryRowWith(ctx).Scan)
		}
		_, err := stmt.ExecWith(ctx)
		return err
	case Updater:
		stmt := pgstmt.Update(func(b pgstmt.UpdateStatement) {
			m.Update(b)
			for _, f := range filter {
				f.Apply(ctx, condUpdateWrapper{b})
			}
		})

		if scanner, ok := m.(Scanner); ok {
			return scanner.Scan(stmt.QueryRowWith(ctx).Scan)
		}
		_, err := stmt.ExecWith(ctx)
		return err
	}

	// *[]*model => []*model => *model => model
	rf := reflect.ValueOf(model).Elem()
	typeSlice := rf.Type()
	typeElem := typeSlice.Elem().Elem()
	rs := reflect.MakeSlice(typeSlice, 0, 0)
	m := reflect.New(typeElem).Interface()

	if m, ok := m.(Selector); ok {
		err := pgstmt.Select(func(b pgstmt.SelectStatement) {
			m.Select(b)
			for _, f := range filter {
				f.Apply(ctx, b)
			}
		}).IterWith(ctx, func(scan pgsql.Scanner) error {
			rx := reflect.New(typeElem)
			err := rx.Interface().(Selector).Scan(scan)
			if err != nil {
				return err
			}
			rs = reflect.Append(rs, rx)
			return nil
		})
		if err != nil {
			return err
		}
		rf.Set(rs)
		return nil
	}

	return fmt.Errorf("not implement")
}
