package models

import (
    "fotstat/global/config"
    "fotstat/models/injury"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"

    log "fotstat/global/log"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"

)

type Injury struct {

    Id                int64 `json:"id"`
    Player                int `json:"player"`
    Type                string `json:"type"`
    Startdate                string `json:"startdate"`
    Returndate                string `json:"returndate"`
    Memo                string `json:"memo"`
    Createddate                string `json:"createddate"`
    Updateddate                string `json:"updateddate"`

    Extra                    map[string]interface{} `json:"extra"`
}

type InjuryManager struct {
    Conn    *Connection
    Result  *sql.Result
    Index   string
    Isolation   bool
    SelectQuery  string
    JoinQuery string
    CountQuery   string
    GroupQuery string
    SelectLog bool
    Log bool
}

func (c *Injury) AddExtra(key string, value interface{}) {
	c.Extra[key] = value
}

func NewInjuryManager(conn *Connection) *InjuryManager {
    var item InjuryManager


    if conn == nil {
        item.Conn = NewConnection()
        item.Isolation = false
    } else {
        item.Conn = conn
        item.Isolation = conn.Isolation
    }

    item.Index = ""
    item.SelectLog = config.Log.Database
    item.Log = config.Log.Database

    return &item
}

func (p *InjuryManager) Close() {
    if p.Conn != nil {
        p.Conn.Close()
    }
}

func (p *InjuryManager) SetIndex(index string) {
    p.Index = index
}

func (p *InjuryManager) SetCountQuery(query string) {
    p.CountQuery = query
}

func (p *InjuryManager) SetSelectQuery(query string) {
    p.SelectQuery = query
}

func (p *InjuryManager) Exec(query string, params ...interface{}) (sql.Result, error) {
    if p.Log {
       if len(params) > 0 {
	       log.Debug().Str("query", query).Any("param", params).Msg("SQL")
       } else {
	       log.Debug().Str("query", query).Msg("SQL")
       }
    }

    return p.Conn.Exec(query, params...)
}

func (p *InjuryManager) Query(query string, params ...interface{}) (*sql.Rows, error) {
    if p.Isolation {
        query += " for update"
    }

    if p.SelectLog {
       if len(params) > 0 {
	       log.Debug().Str("query", query).Any("param", params).Msg("SQL")
       } else {
	       log.Debug().Str("query", query).Msg("SQL")
       }
    }

    return p.Conn.Query(query, params...)
}

func (p *InjuryManager) GetQuery() string {
    if p.SelectQuery != "" {
        return p.SelectQuery
    }

    var ret strings.Builder

    ret.WriteString("select i_id, i_player, i_type, i_startdate, i_returndate, i_memo, i_createddate, i_updateddate from injury_tb")

    if p.Index != "" {
        ret.WriteString(" use index(")
        ret.WriteString(p.Index)
        ret.WriteString(")")
    }

    if p.JoinQuery != "" {
        ret.WriteString(", ")
        ret.WriteString(p.JoinQuery)
    }

    ret.WriteString(" where 1=1 ")


    return ret.String()
}

func (p *InjuryManager) GetQuerySelect() string {
    if p.CountQuery != "" {
        return p.CountQuery
    }

    var ret strings.Builder

    ret.WriteString("select count(*) from injury_tb")

    if p.Index != "" {
        ret.WriteString(" use index(")
        ret.WriteString(p.Index)
        ret.WriteString(")")
    }

    if p.JoinQuery != "" {
        ret.WriteString(", ")
        ret.WriteString(p.JoinQuery)
    }

    ret.WriteString(" where 1=1 ")


    return ret.String()
}

func (p *InjuryManager) GetQueryGroup(name string) string {
    if p.SelectQuery != "" {
        return p.SelectQuery
    }

    var ret strings.Builder
    ret.WriteString("select i_")
    ret.WriteString(name)
    ret.WriteString(", count(*) from injury_tb ")

    if p.Index != "" {
        ret.WriteString(" use index(")
        ret.WriteString(p.Index)
        ret.WriteString(")")
    }

    ret.WriteString(" where 1=1 ")


    return ret.String()
}

func (p *InjuryManager) Truncate() error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "truncate injury_tb "
    _, err := p.Exec(query)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return nil
}

// nullableText maps an optional string to a value safe for a nullable column:
// blank values become SQL NULL instead of an empty string.
func nullableText(value string) interface{} {
    if value == "" {
        return nil
    }
    return value
}

func (p *InjuryManager) Insert(item *Injury) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    if item.Createddate == "" {
        t := time.Now().UTC().Add(time.Hour * 9)
        item.Createddate = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
    }


    if item.Createddate == "" {
       item.Createddate = "1000-01-01 00:00:00"
    }

    if item.Updateddate == "" {
       item.Updateddate = "1000-01-01 00:00:00"
    }

    // i_startdate/i_returndate are DATE columns; store NULL (not "") when blank.
    startdate := nullableDate(item.Startdate)
    returndate := nullableDate(item.Returndate)

    query := ""
    var res sql.Result
    var err error
    if item.Id > 0 {
        query = "insert into injury_tb (i_id, i_player, i_type, i_startdate, i_returndate, i_memo, i_createddate, i_updateddate) values (?, ?, ?, ?, ?, ?, ?, ?)"
        res, err = p.Exec(query, item.Id, item.Player, nullableText(item.Type), startdate, returndate, nullableText(item.Memo), item.Createddate, item.Updateddate)
    } else {
        query = "insert into injury_tb (i_player, i_type, i_startdate, i_returndate, i_memo, i_createddate, i_updateddate) values (?, ?, ?, ?, ?, ?, ?)"
        res, err = p.Exec(query, item.Player, nullableText(item.Type), startdate, returndate, nullableText(item.Memo), item.Createddate, item.Updateddate)
    }

    if err == nil {
        p.Result = &res

    } else {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
        p.Result = nil
    }

    return err
}

func (p *InjuryManager) Delete(id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "delete from injury_tb where i_id = ?"
    _, err := p.Exec(query, id)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }


    return err
}

func (p *InjuryManager) DeleteAll() error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "delete from injury_tb"
    _, err := p.Exec(query)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return err
}

func (p *InjuryManager) MakeQuery(initQuery string , postQuery string, initParams []interface{}, args []interface{}) (string, []interface{}) {
    var params []interface{}
    if initParams != nil {
        params = append(params, initParams...)
    }

    pos := 1

    var query strings.Builder
	query.WriteString(initQuery)

    for _, arg := range args {
        switch v := arg.(type) {
        case Where:
            item := v

            if strings.Contains(item.Column, "_") {
                query.WriteString(" and ")
            } else {
                query.WriteString(" and i_")
            }
            query.WriteString(item.Column)

            if item.Compare == "in" {
                query.WriteString(" in (")
                query.WriteString(strings.Trim(strings.Replace(fmt.Sprint(item.Value), " ", ", ", -1), "[]"))
                query.WriteString(")")
            } else if item.Compare == "not in" {
                query.WriteString(" not in (")
                query.WriteString(strings.Trim(strings.Replace(fmt.Sprint(item.Value), " ", ", ", -1), "[]"))
                query.WriteString(")")
            } else if item.Compare == "between" {
                if config.Database.Type == config.Postgresql {
                    query.WriteString(fmt.Sprintf(" between $%v and $%v", pos, pos + 1))
                    pos += 2
                } else {
                    query.WriteString(" between ? and ?")
                }

                s := item.Value.([2]string)
                params = append(params, s[0])
                params = append(params, s[1])
            } else {
                if config.Database.Type == config.Postgresql {
                    query.WriteString(" ")
                    query.WriteString(item.Compare)
                    query.WriteString(fmt.Sprintf(" $%v", pos))
                    pos++
                } else {
                    query.WriteString(" ")
                    query.WriteString(item.Compare)
                    query.WriteString(" ?")
                }
                if item.Compare == "like" {
                    params = append(params, "%" + item.Value.(string) + "%")
                } else {
                    params = append(params, item.Value)
                }
            }
        case Custom:
             item := v

            query.WriteString(" and ")
            query.WriteString(item.Query)
        }
    }

	query.WriteString(postQuery)

    return query.String(), params
}

func (p *InjuryManager) DeleteWhere(args []interface{}) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query, params := p.MakeQuery("delete from injury_tb where 1=1", "", nil, args)
    _, err := p.Exec(query, params...)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return err
}

func (p *InjuryManager) Update(item *Injury) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }


    if item.Createddate == "" {
       item.Createddate = "1000-01-01 00:00:00"
    }

    if item.Updateddate == "" {
       item.Updateddate = "1000-01-01 00:00:00"
    }


	query := "update injury_tb set i_player = ?, i_type = ?, i_startdate = ?, i_returndate = ?, i_memo = ?, i_createddate = ?, i_updateddate = ? where i_id = ?"
	_, err := p.Exec(query, item.Player, nullableText(item.Type), nullableDate(item.Startdate), nullableDate(item.Returndate), nullableText(item.Memo), item.Createddate, item.Updateddate, item.Id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }


    return err
}

func (p *InjuryManager) UpdateWhere(columns []injury.Params, args []interface{}) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    var initQuery strings.Builder
    var initParams []interface{}

    initQuery.WriteString("update injury_tb set ")
    for i, v := range columns {
        if i > 0 {
            initQuery.WriteString(", ")
        }

        if v.Column == injury.ColumnId {
        initQuery.WriteString("i_id = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == injury.ColumnPlayer {
        initQuery.WriteString("i_player = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == injury.ColumnType {
        initQuery.WriteString("i_type = ?")
        initParams = append(initParams, nullableText(fmt.Sprint(v.Value)))
        } else if v.Column == injury.ColumnStartdate {
        initQuery.WriteString("i_startdate = ?")
        initParams = append(initParams, nullableDate(fmt.Sprint(v.Value)))
        } else if v.Column == injury.ColumnReturndate {
        initQuery.WriteString("i_returndate = ?")
        initParams = append(initParams, nullableDate(fmt.Sprint(v.Value)))
        } else if v.Column == injury.ColumnMemo {
        initQuery.WriteString("i_memo = ?")
        initParams = append(initParams, nullableText(fmt.Sprint(v.Value)))
        } else if v.Column == injury.ColumnCreateddate {
        initQuery.WriteString("i_createddate = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == injury.ColumnUpdateddate {
        initQuery.WriteString("i_updateddate = ?")
        initParams = append(initParams, v.Value)
        } else {

        }
    }

    initQuery.WriteString(" where 1=1 ")

    query, params := p.MakeQuery(initQuery.String(), "", initParams, args)
    _, err := p.Exec(query, params...)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }


    return err
}

func (p *InjuryManager) GetIdentity() int64 {
    if !p.Conn.IsConnect() {
        return 0
    }

    id, err := (*p.Result).LastInsertId()

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
        return 0
    } else {
        return id
    }
}

func (p *Injury) InitExtra() {
    p.Extra = map[string]interface{}{

    }
}

func (p *InjuryManager) ReadRow(rows *sql.Rows) *Injury {
    var item Injury
    var err error



    if rows.Next() {
        var itype sql.NullString
        var startdate sql.NullString
        var returndate sql.NullString
        var memo sql.NullString
        err = rows.Scan(&item.Id, &item.Player, &itype, &startdate, &returndate, &memo, &item.Createddate, &item.Updateddate)
        item.Type = itype.String
        item.Startdate = startdate.String
        item.Returndate = returndate.String
        item.Memo = memo.String

        if item.Createddate == "0000-00-00 00:00:00" || item.Createddate == "1000-01-01 00:00:00" || item.Createddate == "9999-01-01 00:00:00" {
            item.Createddate = ""
        }

        if config.Database.Type == config.Postgresql {
            item.Createddate = strings.ReplaceAll(strings.ReplaceAll(item.Createddate, "T", " "), "Z", "")
        }

        if item.Updateddate == "0000-00-00 00:00:00" || item.Updateddate == "1000-01-01 00:00:00" || item.Updateddate == "9999-01-01 00:00:00" {
            item.Updateddate = ""
        }

        if config.Database.Type == config.Postgresql {
            item.Updateddate = strings.ReplaceAll(strings.ReplaceAll(item.Updateddate, "T", " "), "Z", "")
        }


    } else {
        return nil
    }

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
        return nil
    } else {

        item.InitExtra()

        return &item
    }
}

func (p *InjuryManager) ReadRows(rows *sql.Rows) []Injury {
    var items []Injury

    for rows.Next() {
        var item Injury
        var itype sql.NullString
        var startdate sql.NullString
        var returndate sql.NullString
        var memo sql.NullString

        err := rows.Scan(&item.Id, &item.Player, &itype, &startdate, &returndate, &memo, &item.Createddate, &item.Updateddate)
        if err != nil {
           if p.Log {
             log.Error().Str("error", err.Error()).Msg("SQL")
           }
           break
        }
        item.Type = itype.String
        item.Startdate = startdate.String
        item.Returndate = returndate.String
        item.Memo = memo.String


        if item.Createddate == "0000-00-00 00:00:00" || item.Createddate == "1000-01-01 00:00:00" || item.Createddate == "9999-01-01 00:00:00" {
            item.Createddate = ""
        }

        if config.Database.Type == config.Postgresql {
            item.Createddate = strings.ReplaceAll(strings.ReplaceAll(item.Createddate, "T", " "), "Z", "")
        }

        if item.Updateddate == "0000-00-00 00:00:00" || item.Updateddate == "1000-01-01 00:00:00" || item.Updateddate == "9999-01-01 00:00:00" {
            item.Updateddate = ""
        }

        if config.Database.Type == config.Postgresql {
            item.Updateddate = strings.ReplaceAll(strings.ReplaceAll(item.Updateddate, "T", " "), "Z", "")
        }


        item.InitExtra()

        items = append(items, item)
    }


     return items
}

func (p *InjuryManager) Get(id int64) *Injury {
    if !p.Conn.IsConnect() {
        return nil
    }

    var query strings.Builder
    query.WriteString(p.GetQuery())
    query.WriteString(" and i_id = ?")



    rows, err := p.Query(query.String(), id)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
       return nil
    }

    defer rows.Close()

    return p.ReadRow(rows)
}

func (p *InjuryManager) GetWhere(args []interface{}) *Injury {
    items := p.Find(args)
    if len(items) == 0 {
        return nil
    }

    return &items[0]
}

func (p *InjuryManager) Count(args []interface{}) int {
    if !p.Conn.IsConnect() {
        return 0
    }

    query, params := p.MakeQuery(p.GetQuerySelect(), p.GroupQuery, nil, args)
    rows, err := p.Query(query, params...)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
       return 0
    }

    defer rows.Close()

    if !rows.Next() {
        return 0
    }

    cnt := 0
    err = rows.Scan(&cnt)

    if err != nil {
        return 0
    } else {
        return cnt
    }
}

func (p *InjuryManager) FindAll() []Injury {
    return p.Find(nil)
}

func (p *InjuryManager) Find(args []interface{}) []Injury {
    if !p.Conn.IsConnect() {
        var items []Injury
        return items
    }

    var params []interface{}
    baseQuery := p.GetQuery()

    var query strings.Builder

    page := 0
    pagesize := 0
    orderby := ""

    pos := 1

    for _, arg := range args {
        switch v := arg.(type) {
        case PagingType:
            item := v
            page = item.Page
            pagesize = item.Pagesize
        case OrderingType:
            item := v
            orderby = item.Order
        case LimitType:
            item := v
            page = 1
            pagesize = item.Limit
        case OptionType:
            item := v
            if item.Limit > 0 {
                page = 1
                pagesize = item.Limit
            } else {
                page = item.Page
                pagesize = item.Pagesize
            }
            orderby = item.Order
        case Where:
            item := v

            if strings.Contains(item.Column, "_") {
                query.WriteString(" and ")
            } else {
                query.WriteString(" and i_")
            }
            query.WriteString(item.Column)

            if item.Compare == "in" {
                query.WriteString(" in (")
                query.WriteString(strings.Trim(strings.Replace(fmt.Sprint(item.Value), " ", ", ", -1), "[]"))
                query.WriteString(")")
            } else if item.Compare == "not in" {
                query.WriteString(" not in (")
                query.WriteString(strings.Trim(strings.Replace(fmt.Sprint(item.Value), " ", ", ", -1), "[]"))
                query.WriteString(")")
            } else if item.Compare == "between" {
                if config.Database.Type == config.Postgresql {
                    query.WriteString(fmt.Sprintf(" between $%v and $%v", pos, pos + 1))
                    pos += 2
                } else {
                    query.WriteString(" between ? and ?")
                }

                s := item.Value.([2]string)
                params = append(params, s[0])
                params = append(params, s[1])
            } else {
                if config.Database.Type == config.Postgresql {
                    query.WriteString(" ")
                    query.WriteString(item.Compare)
                    query.WriteString(fmt.Sprintf(" $%v", pos))
                    pos++
                } else {
                    query.WriteString(" ")
                    query.WriteString(item.Compare)
                    query.WriteString(" ?")
                }
                if item.Compare == "like" {
                    params = append(params, "%" + item.Value.(string) + "%")
                } else {
                    params = append(params, item.Value)
                }
            }
        case Custom:
             item := v

            query.WriteString(" and ")
            query.WriteString(item.Query)
        case Base:
             item := v

             baseQuery = item.Query
        }
    }

    query.WriteString(p.GroupQuery)

    startpage := (page - 1) * pagesize

    if page > 0 && pagesize > 0 {
        if orderby == "" {
            orderby = "i_id desc"
        } else {
            if !strings.Contains(orderby, "_") {
                if strings.ToUpper(orderby) != "RAND()" {
                    orderby = "i_" + orderby
                }
            }

        }
        query.WriteString(" order by ")
        query.WriteString(orderby)
        if config.Database.Type == config.Postgresql {
            query.WriteString(fmt.Sprintf(" limit $%v offset $%v", pos, pos + 1))
            params = append(params, pagesize)
            params = append(params, startpage)
        } else if config.Database.Type == config.Mysql {
            query.WriteString(" limit ? offset ?")
            params = append(params, pagesize)
            params = append(params, startpage)
        } else if config.Database.Type == config.Sqlserver {
            query.WriteString("OFFSET ? ROWS FETCH NEXT ? ROWS ONLY")
            params = append(params, startpage)
            params = append(params, pagesize)
        }
    } else {
        if orderby == "" {
            orderby = "i_id"
        } else {
            if !strings.Contains(orderby, "_") {
                if strings.ToUpper(orderby) != "RAND()" {
                    orderby = "i_" + orderby
                }
            }
        }
        query.WriteString(" order by ")
        query.WriteString(orderby)
    }

    rows, err := p.Query(baseQuery + query.String(), params...)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
        items := make([]Injury, 0)
        return items
    }

    defer rows.Close()

    return p.ReadRows(rows)
}

func (p *InjuryManager) GroupBy(name string, args []interface{}) []Groupby {
    if !p.Conn.IsConnect() {
        var items []Groupby
        return items
    }

    var params []interface{}
    baseQuery := p.GetQueryGroup(name)
    var query strings.Builder
    pos := 1

    for _, arg := range args {
        switch v := arg.(type) {
        case Where:
            item := v

            if strings.Contains(item.Column, "_") {
                query.WriteString(" and ")
            } else {
                query.WriteString(" and i_")
            }
            query.WriteString(item.Column)

            if item.Compare == "in" {
                query.WriteString(" in (")
                query.WriteString(strings.Trim(strings.Replace(fmt.Sprint(item.Value), " ", ", ", -1), "[]"))
                query.WriteString(")")
            } else if item.Compare == "not in" {
                query.WriteString(" not in (")
                query.WriteString(strings.Trim(strings.Replace(fmt.Sprint(item.Value), " ", ", ", -1), "[]"))
                query.WriteString(")")
            } else if item.Compare == "between" {
                if config.Database.Type == config.Postgresql {
                    query.WriteString(fmt.Sprintf(" between $%v and $%v", pos, pos + 1))
                    pos += 2
                } else {
                    query.WriteString(" between ? and ?")
                }

                s := item.Value.([2]string)
                params = append(params, s[0])
                params = append(params, s[1])
            } else {
                if config.Database.Type == config.Postgresql {
                    query.WriteString(" ")
                    query.WriteString(item.Compare)
                    query.WriteString(fmt.Sprintf(" $%v", pos))
                    pos++
                } else {
                    query.WriteString(" ")
                    query.WriteString(item.Compare)
                    query.WriteString(" ?")
                }
                if item.Compare == "like" {
                    params = append(params, "%" + item.Value.(string) + "%")
                } else {
                    params = append(params, item.Value)
                }
            }
        case Custom:
             item := v

            query.WriteString(" and ")
            query.WriteString(item.Query)
        case Base:
             item := v

             baseQuery = item.Query
        }
    }

    query.WriteString(" group by i_")
    query.WriteString(name)

    rows, err := p.Query(baseQuery + query.String(), params...)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
        var items []Groupby
        return items
    }

    defer rows.Close()

    var items []Groupby

    for rows.Next() {
        var item Groupby
        err := rows.Scan(&item.Value, &item.Count)
        if err != nil {
           if p.Log {
                log.Error().Str("error", err.Error()).Msg("SQL")
           }
           break
        }

        items = append(items, item)
    }

    return items
}
