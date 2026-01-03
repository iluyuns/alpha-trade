package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

type columnMeta struct {
	Name          string
	UDTName       string
	IsNullable    bool
	IsIdentity    bool
	ColumnDefault sql.NullString
	Comment       string
}

type tableMeta struct {
	Schema           string
	Table            string
	TypeName         string
	LowerTypeName    string
	FileBase         string
	PKColumns        []string
	PKParams         []param
	AutoSetColumns   []string
	Columns          []column
	InsertColumns    []column
	UpdateColumns    []column
	Imports          []string
	GeneratedAtUTC   string
	GeneratorName    string
	GeneratorVersion string
}

type column struct {
	ColName string
	Field   string
	GoType  string
	Comment string
}

type param struct {
	Column string
	Name   string
	GoType string
	Field  string
}

func main() {
	var (
		url        = flag.String("url", "", "postgres url, e.g. postgres://user:pass@host:5432/db?sslmode=disable")
		schema     = flag.String("schema", "public", "schema name")
		table      = flag.String("table", "", "table name (without schema)")
		outDir     = flag.String("dir", "./internal/model", "output dir")
		pkg        = flag.String("package", "model", "go package name")
		withCustom = flag.Bool("with-custom", true, "generate *_model.go wrapper (if not exists)")
	)
	flag.Parse()

	if *url == "" || *table == "" {
		fmt.Fprintln(os.Stderr, "required: --url and --table")
		os.Exit(2)
	}

	db, err := sql.Open("postgres", *url)
	if err != nil {
		die(err)
	}
	defer db.Close()

	meta, err := introspect(db, *schema, *table)
	if err != nil {
		die(err)
	}

	meta.GeneratorName = "pgmodelgen"
	meta.GeneratorVersion = "0.1.0"
	meta.GeneratedAtUTC = time.Now().UTC().Format(time.RFC3339)

	genPath := filepath.Join(*outDir, meta.FileBase+"_model_gen.go")
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		die(err)
	}
	if err := renderToFile(genTpl, map[string]any{
		"Package": *pkg,
		"Meta":    meta,
	}, genPath); err != nil {
		die(err)
	}

	if *withCustom {
		customPath := filepath.Join(*outDir, meta.FileBase+"_model.go")
		if _, err := os.Stat(customPath); err == nil {
			// don't overwrite
		} else if os.IsNotExist(err) {
			if err := renderToFile(customTpl, map[string]any{
				"Package": *pkg,
				"Meta":    meta,
			}, customPath); err != nil {
				die(err)
			}
		} else {
			die(err)
		}
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func introspect(db *sql.DB, schema, table string) (tableMeta, error) {
	cols, err := readColumns(db, schema, table)
	if err != nil {
		return tableMeta{}, err
	}
	comments, err := readColumnComments(db, schema, table)
	if err != nil {
		return tableMeta{}, err
	}
	for i := range cols {
		if c, ok := comments[cols[i].Name]; ok {
			cols[i].Comment = c
		}
	}

	pkCols, err := readPrimaryKeyColumns(db, schema, table)
	if err != nil {
		return tableMeta{}, err
	}
	if len(pkCols) == 0 {
		return tableMeta{}, fmt.Errorf("table %s.%s: missing primary key (pgmodelgen requires PK; composite PK is supported)", schema, table)
	}

	typeName := toCamel(table)
	lowerTypeName := lowerFirst(typeName)

	// Decide auto-set columns (identity or nextval()).
	autoSet := map[string]bool{}
	for _, c := range cols {
		if c.IsIdentity {
			autoSet[c.Name] = true
			continue
		}
		if c.ColumnDefault.Valid && strings.HasPrefix(strings.ToLower(strings.TrimSpace(c.ColumnDefault.String)), "nextval(") {
			autoSet[c.Name] = true
		}
	}
	autoSetCols := make([]string, 0, len(autoSet))
	for k := range autoSet {
		autoSetCols = append(autoSetCols, k)
	}
	sort.Strings(autoSetCols)

	colModels := make([]column, 0, len(cols))
	insertCols := make([]column, 0, len(cols))
	updateCols := make([]column, 0, len(cols))
	pkSet := make(map[string]bool, len(pkCols))
	for _, p := range pkCols {
		pkSet[p] = true
	}

	for _, c := range cols {
		goType := pgTypeToGoType(c.UDTName)
		field := toCamel(c.Name)
		colModels = append(colModels, column{
			ColName: c.Name,
			Field:   field,
			GoType:  goType,
			Comment: c.Comment,
		})
		if !autoSet[c.Name] {
			insertCols = append(insertCols, column{
				ColName: c.Name,
				Field:   field,
				GoType:  goType,
				Comment: c.Comment,
			})
		}
		// For updates, don't update PK columns or auto-set columns.
		if !autoSet[c.Name] && !pkSet[c.Name] {
			updateCols = append(updateCols, column{
				ColName: c.Name,
				Field:   field,
				GoType:  goType,
				Comment: c.Comment,
			})
		}
	}

	// Primary key params (typed based on the column).
	colTypeByName := map[string]string{}
	for _, c := range colModels {
		colTypeByName[c.ColName] = c.GoType
	}
	pkParams := make([]param, 0, len(pkCols))
	for _, pk := range pkCols {
		pkParams = append(pkParams, param{
			Column: pk,
			Name:   toLowerCamel(pk),
			GoType: colTypeByName[pk],
			Field:  toCamel(pk),
		})
	}

	importSet := map[string]bool{
		`"context"`:                         true,
		`"database/sql"`:                    true,
		`"fmt"`:                             true,
		`"strings"`:                         true,
		`"github.com/Masterminds/squirrel"`: true,
		`"github.com/zeromicro/go-zero/core/stores/builder"`: true,
		`"github.com/zeromicro/go-zero/core/stores/sqlx"`:    true,
		`"github.com/zeromicro/go-zero/core/stringx"`:        true,
	}
	for _, c := range colModels {
		if c.GoType == "time.Time" {
			importSet[`"time"`] = true
			break
		}
	}
	imports := make([]string, 0, len(importSet))
	for imp := range importSet {
		imports = append(imports, imp)
	}
	sort.Strings(imports)

	return tableMeta{
		Schema:         schema,
		Table:          table,
		TypeName:       typeName,
		LowerTypeName:  lowerTypeName,
		FileBase:       table,
		PKColumns:      pkCols,
		PKParams:       pkParams,
		AutoSetColumns: autoSetCols,
		Columns:        colModels,
		InsertColumns:  insertCols,
		UpdateColumns:  updateCols,
		Imports:        imports,
	}, nil
}

func readColumns(db *sql.DB, schema, table string) ([]columnMeta, error) {
	const q = `
select
  c.column_name,
  c.udt_name,
  c.is_nullable = 'YES' as is_nullable,
  c.is_identity = 'YES' as is_identity,
  c.column_default
from information_schema.columns c
where c.table_schema = $1
  and c.table_name = $2
order by c.ordinal_position`
	rows, err := db.Query(q, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []columnMeta
	for rows.Next() {
		var m columnMeta
		if err := rows.Scan(&m.Name, &m.UDTName, &m.IsNullable, &m.IsIdentity, &m.ColumnDefault); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func readPrimaryKeyColumns(db *sql.DB, schema, table string) ([]string, error) {
	const q = `
select kcu.column_name
from information_schema.table_constraints tc
join information_schema.key_column_usage kcu
  on tc.constraint_name = kcu.constraint_name
  and tc.table_schema = kcu.table_schema
where tc.table_schema = $1
  and tc.table_name = $2
  and tc.constraint_type = 'PRIMARY KEY'
order by kcu.ordinal_position`
	rows, err := db.Query(q, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func readColumnComments(db *sql.DB, schema, table string) (map[string]string, error) {
	const q = `
select
  a.attname as column_name,
  coalesce(d.description, '') as description
from pg_catalog.pg_attribute a
join pg_catalog.pg_class c on a.attrelid = c.oid
join pg_catalog.pg_namespace n on c.relnamespace = n.oid
left join pg_catalog.pg_description d on d.objoid = c.oid and d.objsubid = a.attnum
where n.nspname = $1
  and c.relname = $2
  and a.attnum > 0
  and not a.attisdropped`
	rows, err := db.Query(q, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var name, desc string
		if err := rows.Scan(&name, &desc); err != nil {
			return nil, err
		}
		out[name] = desc
	}
	return out, rows.Err()
}

func pgTypeToGoType(udt string) string {
	switch strings.ToLower(udt) {
	case "int2", "int4", "int8", "integer", "bigint", "smallint":
		return "int64"
	case "bool":
		return "bool"
	case "varchar", "text", "bpchar", "uuid":
		return "string"
	case "json", "jsonb":
		return "string"
	case "float4", "float8":
		return "float64"
	case "numeric", "decimal":
		// keep simple; can be upgraded to decimal.Decimal with a config if needed.
		return "float64"
	case "timestamp", "timestamptz", "date":
		// most of this repo stores millis in bigint; keep time types explicit if they appear.
		return "time.Time"
	default:
		return "string"
	}
}

func toCamel(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	for i := range parts {
		p := strings.ToLower(parts[i])
		if p == "id" {
			parts[i] = "Id"
			continue
		}
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func toLowerCamel(s string) string {
	cc := toCamel(s)
	return lowerFirst(cc)
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func renderToFile(tpl string, data any, outPath string) error {
	t, err := template.New("tpl").Funcs(template.FuncMap{
		"Join":    strings.Join,
		"Add":     func(a, b int) int { return a + b },
		"ToCamel": toCamel,
	}).Parse(tpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// keep raw for easier debugging
		formatted = buf.Bytes()
	}
	return os.WriteFile(outPath, formatted, 0o644)
}

const genTpl = `
// Code generated by {{.Meta.GeneratorName}}. DO NOT EDIT.
// generated_at_utc: {{.Meta.GeneratedAtUTC}}
// version: {{.Meta.GeneratorVersion}}

package {{.Package}}

import (
{{- range .Meta.Imports }}
	{{ . }}
{{- end }}
)

var (
	{{.Meta.LowerTypeName}}FieldNames          = builder.RawFieldNames(&{{.Meta.TypeName}}{}, true)
	{{.Meta.LowerTypeName}}Rows                = strings.Join({{.Meta.LowerTypeName}}FieldNames, ",")
	{{.Meta.LowerTypeName}}RowsExpectAutoSet   = strings.Join(stringx.Remove({{.Meta.LowerTypeName}}FieldNames{{- range .Meta.AutoSetColumns}}, "{{.}}"{{- end}}), ",")
)

type (
	{{.Meta.LowerTypeName}}Model interface {
		Insert(ctx context.Context, data *{{.Meta.TypeName}}) (sql.Result, error)
		InsertReturn(ctx context.Context, session sqlx.Session, data *{{.Meta.TypeName}}) (*{{.Meta.TypeName}}, error)
		BatchInsertReturn(ctx context.Context, session sqlx.Session, dataList []*{{.Meta.TypeName}}) ([]*{{.Meta.TypeName}}, error)
		FindOne(ctx context.Context{{range .Meta.PKParams}}, {{.Name}} {{.GoType}}{{end}}) (*{{.Meta.TypeName}}, error)
		Update(ctx context.Context, data *{{.Meta.TypeName}}) error
		Delete(ctx context.Context{{range .Meta.PKParams}}, {{.Name}} {{.GoType}}{{end}}) error
	}

	default{{.Meta.TypeName}}Model struct {
		conn  sqlx.SqlConn
		table string
	}

	{{.Meta.TypeName}} struct {
	{{- range .Meta.Columns }}
		{{.Field}} {{.GoType}} ` + "`" + `db:"{{.ColName}}"` + "`" + `{{if .Comment}} // {{.Comment}}{{end}}
	{{- end }}
	}
)

func new{{.Meta.TypeName}}Model(conn sqlx.SqlConn) *default{{.Meta.TypeName}}Model {
	return &default{{.Meta.TypeName}}Model{
		conn:  conn,
		table: "\"{{.Meta.Schema}}\".\"{{.Meta.Table}}\"",
	}
}

func (m *default{{.Meta.TypeName}}Model) Delete(ctx context.Context{{range .Meta.PKParams}}, {{.Name}} {{.GoType}}{{end}}) error {
	query := fmt.Sprintf("delete from %s where {{range $i, $pk := .Meta.PKColumns}}{{if $i}} and {{end}}{{$pk}} = ${{Add $i 1}}{{end}}", m.table)
	_, err := m.conn.ExecCtx(ctx, query{{- range .Meta.PKParams}}, {{.Name}}{{- end}})
	return err
}

func (m *default{{.Meta.TypeName}}Model) FindOne(ctx context.Context{{range .Meta.PKParams}}, {{.Name}} {{.GoType}}{{end}}) (*{{.Meta.TypeName}}, error) {
	query := fmt.Sprintf("select %s from %s where {{range $i, $pk := .Meta.PKColumns}}{{if $i}} and {{end}}{{$pk}} = ${{Add $i 1}}{{end}} limit 1", {{.Meta.LowerTypeName}}Rows, m.table)
	var resp {{.Meta.TypeName}}
	err := m.conn.QueryRowCtx(ctx, &resp, query{{- range .Meta.PKParams}}, {{.Name}}{{- end}})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *default{{.Meta.TypeName}}Model) Insert(ctx context.Context, data *{{.Meta.TypeName}}) (sql.Result, error) {
	builder := m.insertBuilder().Columns({{.Meta.LowerTypeName}}RowsExpectAutoSet).Values({{range $i, $c := .Meta.InsertColumns}}{{if $i}}, {{end}}data.{{$c.Field}}{{end}})
	querySql, values, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	return m.conn.ExecCtx(ctx, querySql, values...)
}

func (m *default{{.Meta.TypeName}}Model) BatchInsertReturn(ctx context.Context, session sqlx.Session, dataList []*{{.Meta.TypeName}}) ([]*{{.Meta.TypeName}}, error) {
	builder := m.insertBuilder().Columns({{.Meta.LowerTypeName}}RowsExpectAutoSet)
	for _, data := range dataList {
		builder = builder.Values({{range $i, $c := .Meta.InsertColumns}}{{if $i}}, {{end}}data.{{$c.Field}}{{end}})
	}
	return m.insertListWithReturn(ctx, session, builder)
}

func (m *default{{.Meta.TypeName}}Model) InsertReturn(ctx context.Context, session sqlx.Session, data *{{.Meta.TypeName}}) (*{{.Meta.TypeName}}, error) {
	builder := m.insertBuilder().Columns({{.Meta.LowerTypeName}}RowsExpectAutoSet).Values({{range $i, $c := .Meta.InsertColumns}}{{if $i}}, {{end}}data.{{$c.Field}}{{end}})
	return m.insertWithReturn(ctx, session, builder)
}

func (m *default{{.Meta.TypeName}}Model) Update(ctx context.Context, newData *{{.Meta.TypeName}}) error {
	builder := m.updateBuilder()
	{{- range .Meta.UpdateColumns}}
	builder = builder.Set("{{.ColName}}", newData.{{.Field}})
	{{- end }}
	builder = builder.Where(squirrel.Eq{
	{{- range .Meta.PKParams}}
		"{{.Column}}": newData.{{.Field}},
	{{- end }}
	})
	return m.execCtxWithSession(ctx, nil, builder)
}

func (m *default{{.Meta.TypeName}}Model) tableName() string {
	return m.table
}

func (m *default{{.Meta.TypeName}}Model) selectBuilder() squirrel.SelectBuilder {
	return squirrel.Select().PlaceholderFormat(squirrel.Dollar).From(m.table)
}

func (m *default{{.Meta.TypeName}}Model) insertBuilder() squirrel.InsertBuilder {
	return squirrel.Insert(m.table).PlaceholderFormat(squirrel.Dollar)
}

func (m *default{{.Meta.TypeName}}Model) replaceBuilder() squirrel.InsertBuilder {
	return squirrel.Replace(m.table).PlaceholderFormat(squirrel.Dollar)
}

func (m *default{{.Meta.TypeName}}Model) updateBuilder() squirrel.UpdateBuilder {
	return squirrel.Update(m.table).PlaceholderFormat(squirrel.Dollar)
}

func (m *default{{.Meta.TypeName}}Model) deleteBuilder() squirrel.DeleteBuilder {
	return squirrel.Delete(m.table).PlaceholderFormat(squirrel.Dollar)
}

func (m *default{{.Meta.TypeName}}Model) execCtxWithSession(ctx context.Context, session sqlx.Session, sqlizer squirrel.Sqlizer) error {
	sqlStr, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}
	if session != nil {
		_, err = session.Exec(sqlStr, args...)
	} else {
		_, err = m.conn.ExecCtx(ctx, sqlStr, args...)
	}
	return err
}

func (m *default{{.Meta.TypeName}}Model) insertListWithReturn(ctx context.Context, session sqlx.Session, sqlizer squirrel.InsertBuilder) ([]*{{.Meta.TypeName}}, error) {
	querySql, values, err := sqlizer.Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, err
	}
	var resp []*{{.Meta.TypeName}}
	if session != nil {
		err = session.QueryRowsCtx(ctx, &resp, querySql, values...)
	} else {
		err = m.conn.QueryRowsCtx(ctx, &resp, querySql, values...)
	}
	return resp, err
}

func (m *default{{.Meta.TypeName}}Model) insertWithReturn(ctx context.Context, session sqlx.Session, sqlizer squirrel.InsertBuilder) (*{{.Meta.TypeName}}, error) {
	querySql, values, err := sqlizer.Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, err
	}
	var resp {{.Meta.TypeName}}
	if session != nil {
		err = session.QueryRowCtx(ctx, &resp, querySql, values...)
	} else {
		err = m.conn.QueryRowCtx(ctx, &resp, querySql, values...)
	}
	return &resp, err
}

// findCount 根据squirrel.SelectBuilder生成的sql查询当前表条数
func (m *default{{.Meta.TypeName}}Model) findCount(ctx context.Context, builder squirrel.SelectBuilder) (int64, error) {
	builder = builder.Columns("COUNT(" + m.tableName() + ".{{index .Meta.PKColumns 0}})")
	query, values, err := builder.ToSql()
	if err != nil {
		return 0, err
	}
	var resp int64
	err = m.conn.QueryRowCtx(ctx, &resp, query, values...)
	if err != nil {
		return 0, err
	}
	return resp, nil
}

// findList 根据squirrel.SelectBuilder生成的sql查询当前表所有字段返回对象
func (m *default{{.Meta.TypeName}}Model) findList(ctx context.Context, builder squirrel.SelectBuilder) ([]*{{.Meta.TypeName}}, error) {
	builder = builder.Columns(m.tableName() + ".*")
	querySql, values, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	var resp []*{{.Meta.TypeName}}
	err = m.conn.QueryRowsCtx(ctx, &resp, querySql, values...)
	return resp, err
}

// findListWithAny 根据squirrel.SelectBuilder生成的sql查询指定的
func (m *default{{.Meta.TypeName}}Model) findListWithAny(ctx context.Context, sqlizer squirrel.Sqlizer, v any) error {
	querySql, values, err := sqlizer.ToSql()
	if err != nil {
		return err
	}
	return m.conn.QueryRowsCtx(ctx, v, querySql, values...)
}

// findWithAny 根据squirrel.SelectBuilder生成的sql查询指定的
func (m *default{{.Meta.TypeName}}Model) findWithAny(ctx context.Context, sqlizer squirrel.Sqlizer, v any) error {
	querySql, values, err := sqlizer.ToSql()
	if err != nil {
		return err
	}
	return m.conn.QueryRowCtx(ctx, v, querySql, values...)
}

// buildPageQuery 构建分页条件
func (m *default{{.Meta.TypeName}}Model) buildPageQuery(builder squirrel.SelectBuilder, offset int64, limit int64, orderBy string, isAsc bool) squirrel.SelectBuilder {
	builder = builder.Offset(uint64(offset))
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	orderType := "DESC"
	if isAsc {
		orderType = "ASC"
	}
	if !stringx.Contains({{.Meta.LowerTypeName}}FieldNames, orderBy) {
		orderBy = "{{index .Meta.PKColumns 0}}"
	}
	return builder.OrderBy(fmt.Sprintf("%s %s", orderBy, orderType))
}

// execResultCtxWithSession 根据sqlizer生产sql执行并返回结果
func (m *default{{.Meta.TypeName}}Model) execResultCtxWithSession(ctx context.Context, session sqlx.Session, sqlizer squirrel.Sqlizer) (sql.Result, error) {
	sqlStr, args, err := sqlizer.ToSql()
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session.Exec(sqlStr, args...)
	}
	return m.conn.ExecCtx(ctx, sqlStr, args...)
}

// updateWithReturn 根据squirrel.UpdateBuilder条件构建更新语句并返回更新后的对象
func (m *default{{.Meta.TypeName}}Model) updateWithReturn(ctx context.Context, session sqlx.Session, sqlizer squirrel.UpdateBuilder) ([]*{{.Meta.TypeName}}, error) {
	querySql, values, err := sqlizer.Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, err
	}
	var resp []*{{.Meta.TypeName}}
	if session != nil {
		err = session.QueryRowsCtx(ctx, &resp, querySql, values...)
	} else {
		err = m.conn.QueryRowsCtx(ctx, &resp, querySql, values...)
	}
	return resp, err
}

// deleteWithReturn 根据squirrel.DeleteBuilder条件构建删除语句并返回被删除的对象
func (m *default{{.Meta.TypeName}}Model) deleteWithReturn(ctx context.Context, session sqlx.Session, sqlizer squirrel.DeleteBuilder) ([]*{{.Meta.TypeName}}, error) {
	querySql, values, err := sqlizer.Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, err
	}
	var resp []*{{.Meta.TypeName}}
	if session != nil {
		err = session.QueryRowsCtx(ctx, &resp, querySql, values...)
	} else {
		err = m.conn.QueryRowsCtx(ctx, &resp, querySql, values...)
	}
	return resp, err
}
`

const customTpl = `
package {{.Package}}

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ {{.Meta.TypeName}}Model = (*custom{{.Meta.TypeName}}Model)(nil)

type (
	// {{.Meta.TypeName}}Model is an interface to be customized, add more methods here,
	// and implement the added methods in custom{{.Meta.TypeName}}Model.
	{{.Meta.TypeName}}Model interface {
		{{.Meta.LowerTypeName}}Model
		withSession(session sqlx.Session) {{.Meta.TypeName}}Model
	}

	custom{{.Meta.TypeName}}Model struct {
		*default{{.Meta.TypeName}}Model
	}
)

// New{{.Meta.TypeName}}Model returns a model for the database table.
func New{{.Meta.TypeName}}Model(conn sqlx.SqlConn) {{.Meta.TypeName}}Model {
	return &custom{{.Meta.TypeName}}Model{
		default{{.Meta.TypeName}}Model: new{{.Meta.TypeName}}Model(conn),
	}
}

func (m *custom{{.Meta.TypeName}}Model) withSession(session sqlx.Session) {{.Meta.TypeName}}Model {
	return New{{.Meta.TypeName}}Model(sqlx.NewSqlConnFromSession(session))
}
`
