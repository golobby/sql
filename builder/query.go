package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type query struct {
	table     string
	projected *selectClause
	filters   []*whereClause
	orderBy   *orderbyClause
	groupBy   *groupByClause
	joins     []*joinClause
}

type whereClause struct {
	parent *query
	conds  []string
}

type whereHelpers struct {
	Like    func(column string, pattern string) string
	In      func(column string, values ...string) string
	Between func(column string, lower string, higher string) string
	Not     func(cond ...string) string
}

var WhereHelpers = &whereHelpers{
	Like:    like,
	In:      in,
	Between: between,
	Not:     not,
}

func like(column string, pattern string) string {
	return fmt.Sprintf("%s LIKE %s", column, pattern)
}

func in(column string, values ...string) string {
	return fmt.Sprintf("%s IN (%s)", column, strings.Join(values, ", "))
}

func between(column string, lower string, higher string) string {
	return fmt.Sprintf("%s BETWEEN %s AND %s", column, lower, higher)
}

func not(cond ...string) string {
	return fmt.Sprintf("NOT %s", strings.Join(cond, " "))
}

func (w *whereClause) And(parts ...string) *whereClause {
	w.conds = append(w.conds, "AND "+strings.Join(parts, " "))
	return w
}
func (w *whereClause) Query() *query {
	return w.parent
}
func (w *whereClause) Or(parts ...string) *whereClause {
	w.conds = append(w.conds, "OR "+strings.Join(parts, " "))
	return w
}
func (w *whereClause) Not(parts ...string) *whereClause {
	w.conds = append(w.conds, "NOT "+strings.Join(parts, " "))
	return w
}

func (w *whereClause) String() string {
	return strings.Join(w.conds, " ")
}
func (q *query) Table(t string) *query {
	q.table = t
	return q
}
func (q *query) Select(columns ...string) *selectClause {
	q.projected = &selectClause{
		parent:   q,
		distinct: false,
		columns:  columns,
	}
	return q.projected
}
func (q *query) InnerJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "INNER"}
	q.joins = append(q.joins, j)
	return j
}

func (q *query) RightJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "RIGHT"}
	q.joins = append(q.joins, j)
	return j

}
func (q *query) LeftJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "LEFT"}
	q.joins = append(q.joins, j)
	return j

}
func (q *query) FullOuterJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "FULL OUTER"}
	q.joins = append(q.joins, j)
	return j

}
func (q *query) GroupBy(columns ...string) *groupByClause {
	q.groupBy = &groupByClause{
		parent:  q,
		columns: columns,
	}
	return q.groupBy
}

func (q *query) Where(parts ...string) *whereClause {
	w := &whereClause{conds: []string{strings.Join(parts, " ")}, parent: q}
	q.filters = append(q.filters, w)
	return w
}

func (q *query) OrderBy(columns ...string) *orderbyClause {
	q.orderBy = &orderbyClause{
		parent:  q,
		columns: columns,
		desc:    false,
	}

	return q.orderBy
}

func (q *query) SQL() (string, error) {
	sections := []string{}
	// handle select
	if q.projected == nil {
		q.projected = &selectClause{parent: q, distinct: false, columns: []string{"*"}}
	}

	sections = append(sections, "SELECT", q.projected.String())

	if q.table == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	// Handle from TABLE-NAME
	sections = append(sections, "FROM "+q.table)

	// handle where
	if q.filters != nil {
		sections = append(sections, "WHERE")
		for _, f := range q.filters {
			sections = append(sections, f.String())
		}
	}

	if q.orderBy != nil {
		sections = append(sections, "ORDER BY", q.orderBy.String())
	}

	if q.groupBy != nil {
		sections = append(sections, "GROUP BY", q.groupBy.String())
	}

	if q.joins != nil {
		for _, join := range q.joins {
			sections = append(sections, join.String())
		}
	}

	return strings.Join(sections, " "), nil
}

func (q *query) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}

func (q *query) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}

func (q *query) Bind(db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	return _bind(context.Background(), db, v, s, args)
}

func (q *query) BindContext(ctx context.Context, db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	return _bind(ctx, db, v, s, args)
}

func NewQuery() *query {
	return &query{}
}
