package models

import (
    "fotstat/global/config"
    "fotstat/models/record"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"

    log "fotstat/global/log"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"

)

type Record struct {
            
    Id                int64 `json:"id"`         
    Quarter                int `json:"quarter"`         
    Player                int `json:"player"`         
    Min                int `json:"min"`         
    Goal                int `json:"goal"`         
    Assist                int `json:"assist"`         
    Createddate                string `json:"createddate"`         
    Updateddate                string `json:"updateddate"` 
    
    Extra                    map[string]interface{} `json:"extra"`
}

type RecordManager struct {
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

func (c *Record) AddExtra(key string, value interface{}) {    
	c.Extra[key] = value     
}

func NewRecordManager(conn *Connection) *RecordManager {
    var item RecordManager


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

func (p *RecordManager) Close() {
    if p.Conn != nil {
        p.Conn.Close()
    }
}

func (p *RecordManager) SetIndex(index string) {
    p.Index = index
}

func (p *RecordManager) SetCountQuery(query string) {
    p.CountQuery = query
}

func (p *RecordManager) SetSelectQuery(query string) {
    p.SelectQuery = query
}

func (p *RecordManager) Exec(query string, params ...interface{}) (sql.Result, error) {
    if p.Log {
       if len(params) > 0 {
	       log.Debug().Str("query", query).Any("param", params).Msg("SQL")
       } else {
	       log.Debug().Str("query", query).Msg("SQL")
       }
    }

    return p.Conn.Exec(query, params...)
}

func (p *RecordManager) Query(query string, params ...interface{}) (*sql.Rows, error) {
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

func (p *RecordManager) GetQuery() string {
    if p.SelectQuery != "" {
        return p.SelectQuery    
    }

    var ret strings.Builder

    ret.WriteString("select r_id, r_quarter, r_player, r_min, r_goal, r_assist, r_createddate, r_updateddate from record_tb")

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

func (p *RecordManager) GetQuerySelect() string {
    if p.CountQuery != "" {
        return p.CountQuery    
    }

    var ret strings.Builder
    
    ret.WriteString("select count(*) from record_tb")

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

func (p *RecordManager) GetQueryGroup(name string) string {
    if p.SelectQuery != "" {
        return p.SelectQuery    
    }

    var ret strings.Builder
    ret.WriteString("select r_")
    ret.WriteString(name)
    ret.WriteString(", count(*) from record_tb ")

    if p.Index != "" {
        ret.WriteString(" use index(")
        ret.WriteString(p.Index)
        ret.WriteString(")")
    }

    ret.WriteString(" where 1=1 ")
    

    return ret.String()
}

func (p *RecordManager) Truncate() error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }
    
    query := "truncate record_tb "
    _, err := p.Exec(query)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return nil
}

func (p *RecordManager) Insert(item *Record) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    if item.Createddate == "" {
        t := time.Now().UTC().Add(time.Hour * 9)
        //t := time.Now()
        item.Createddate = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
    }

    
    if item.Createddate == "" {
       item.Createddate = "1000-01-01 00:00:00"
    }
	
    if item.Updateddate == "" {
       item.Updateddate = "1000-01-01 00:00:00"
    }
	

    query := ""
    var res sql.Result
    var err error
    if item.Id > 0 {
        query = "insert into record_tb (r_id, r_quarter, r_player, r_min, r_goal, r_assist, r_createddate, r_updateddate) values (?, ?, ?, ?, ?, ?, ?, ?)"
        res, err = p.Exec(query, item.Id, item.Quarter, item.Player, item.Min, item.Goal, item.Assist, item.Createddate, item.Updateddate)
    } else {
        query = "insert into record_tb (r_quarter, r_player, r_min, r_goal, r_assist, r_createddate, r_updateddate) values (?, ?, ?, ?, ?, ?, ?)"
        res, err = p.Exec(query, item.Quarter, item.Player, item.Min, item.Goal, item.Assist, item.Createddate, item.Updateddate)
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

func (p *RecordManager) Delete(id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "delete from record_tb where r_id = ?"
    _, err := p.Exec(query, id)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    
    return err
}

func (p *RecordManager) DeleteAll() error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "delete from record_tb"
    _, err := p.Exec(query)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return err
}

func (p *RecordManager) MakeQuery(initQuery string , postQuery string, initParams []interface{}, args []interface{}) (string, []interface{}) {
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
                query.WriteString(" and r_")
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

func (p *RecordManager) DeleteWhere(args []interface{}) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query, params := p.MakeQuery("delete from record_tb where 1=1", "", nil, args)
    _, err := p.Exec(query, params...)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return err
}

func (p *RecordManager) Update(item *Record) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }
    
    
    if item.Createddate == "" {
       item.Createddate = "1000-01-01 00:00:00"
    }
	
    if item.Updateddate == "" {
       item.Updateddate = "1000-01-01 00:00:00"
    }
	

	query := "update record_tb set r_quarter = ?, r_player = ?, r_min = ?, r_goal = ?, r_assist = ?, r_createddate = ?, r_updateddate = ? where r_id = ?"
	_, err := p.Exec(query, item.Quarter, item.Player, item.Min, item.Goal, item.Assist, item.Createddate, item.Updateddate, item.Id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }
    
        
    return err
}

func (p *RecordManager) UpdateWhere(columns []record.Params, args []interface{}) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    var initQuery strings.Builder
    var initParams []interface{}

    initQuery.WriteString("update record_tb set ")
    for i, v := range columns {
        if i > 0 {
            initQuery.WriteString(", ")
        }

        if v.Column == record.ColumnId {
        initQuery.WriteString("r_id = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == record.ColumnQuarter {
        initQuery.WriteString("r_quarter = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == record.ColumnPlayer {
        initQuery.WriteString("r_player = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == record.ColumnMin {
        initQuery.WriteString("r_min = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == record.ColumnGoal {
        initQuery.WriteString("r_goal = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == record.ColumnAssist {
        initQuery.WriteString("r_assist = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == record.ColumnCreateddate {
        initQuery.WriteString("r_createddate = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == record.ColumnUpdateddate {
        initQuery.WriteString("r_updateddate = ?")
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

/*


func (p *RecordManager) UpdateQuarter(value int, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update record_tb set r_quarter = ? where r_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *RecordManager) UpdatePlayer(value int, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update record_tb set r_player = ? where r_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *RecordManager) UpdateMin(value int, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update record_tb set r_min = ? where r_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *RecordManager) UpdateGoal(value int, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update record_tb set r_goal = ? where r_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *RecordManager) UpdateAssist(value int, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update record_tb set r_assist = ? where r_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *RecordManager) UpdateCreateddate(value string, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update record_tb set r_createddate = ? where r_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *RecordManager) UpdateUpdateddate(value string, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update record_tb set r_updateddate = ? where r_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}


*/

func (p *RecordManager) GetIdentity() int64 {
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

func (p *Record) InitExtra() {
    p.Extra = map[string]interface{}{

    }
}

func (p *RecordManager) ReadRow(rows *sql.Rows) *Record {
    var item Record
    var err error

    

    if rows.Next() {
        err = rows.Scan(&item.Id, &item.Quarter, &item.Player, &item.Min, &item.Goal, &item.Assist, &item.Createddate, &item.Updateddate)
        
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

func (p *RecordManager) ReadRows(rows *sql.Rows) []Record {
    var items []Record

    for rows.Next() {
        var item Record
        

        err := rows.Scan(&item.Id, &item.Quarter, &item.Player, &item.Min, &item.Goal, &item.Assist, &item.Createddate, &item.Updateddate)
        if err != nil {
           if p.Log {
             log.Error().Str("error", err.Error()).Msg("SQL")
           }
           break
        }

        
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

func (p *RecordManager) Get(id int64) *Record {
    if !p.Conn.IsConnect() {
        return nil
    }

    var query strings.Builder
    query.WriteString(p.GetQuery())
    query.WriteString(" and r_id = ?")

    
    
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

func (p *RecordManager) GetWhere(args []interface{}) *Record {
    items := p.Find(args)
    if len(items) == 0 {
        return nil
    }

    return &items[0]
}

func (p *RecordManager) Count(args []interface{}) int {
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

func (p *RecordManager) FindAll() []Record {
    return p.Find(nil)
}

func (p *RecordManager) Find(args []interface{}) []Record {
    if !p.Conn.IsConnect() {
        var items []Record
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
                query.WriteString(" and r_")
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
            orderby = "r_id desc"
        } else {
            if !strings.Contains(orderby, "_") {                   
                if strings.ToUpper(orderby) != "RAND()" {
                    orderby = "r_" + orderby
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
            orderby = "r_id"
        } else {
            if !strings.Contains(orderby, "_") {
                if strings.ToUpper(orderby) != "RAND()" {
                    orderby = "r_" + orderby
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
        items := make([]Record, 0)
        return items
    }

    defer rows.Close()

    return p.ReadRows(rows)
}





func (p *RecordManager) GroupBy(name string, args []interface{}) []Groupby {
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
                query.WriteString(" and r_")
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
    
    query.WriteString(" group by r_")
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
