package models

import (
    "fotstat/global/config"
    "fotstat/models/player"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"

    log "fotstat/global/log"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"

)

type Player struct {
            
    Id                int64 `json:"id"`         
    Team                int `json:"team"`         
    Name                string `json:"name"`         
    Number                int `json:"number"`         
    Position                string `json:"position"`         
    Createddate                string `json:"createddate"`         
    Updateddate                string `json:"updateddate"` 
    
    Extra                    map[string]interface{} `json:"extra"`
}

type PlayerManager struct {
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

func (c *Player) AddExtra(key string, value interface{}) {    
	c.Extra[key] = value     
}

func NewPlayerManager(conn *Connection) *PlayerManager {
    var item PlayerManager


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

func (p *PlayerManager) Close() {
    if p.Conn != nil {
        p.Conn.Close()
    }
}

func (p *PlayerManager) SetIndex(index string) {
    p.Index = index
}

func (p *PlayerManager) SetCountQuery(query string) {
    p.CountQuery = query
}

func (p *PlayerManager) SetSelectQuery(query string) {
    p.SelectQuery = query
}

func (p *PlayerManager) Exec(query string, params ...interface{}) (sql.Result, error) {
    if p.Log {
       if len(params) > 0 {
	       log.Debug().Str("query", query).Any("param", params).Msg("SQL")
       } else {
	       log.Debug().Str("query", query).Msg("SQL")
       }
    }

    return p.Conn.Exec(query, params...)
}

func (p *PlayerManager) Query(query string, params ...interface{}) (*sql.Rows, error) {
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

func (p *PlayerManager) GetQuery() string {
    if p.SelectQuery != "" {
        return p.SelectQuery    
    }

    var ret strings.Builder

    ret.WriteString("select p_id, p_team, p_name, p_number, p_position, p_createddate, p_updateddate from player_tb")

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

func (p *PlayerManager) GetQuerySelect() string {
    if p.CountQuery != "" {
        return p.CountQuery    
    }

    var ret strings.Builder
    
    ret.WriteString("select count(*) from player_tb")

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

func (p *PlayerManager) GetQueryGroup(name string) string {
    if p.SelectQuery != "" {
        return p.SelectQuery    
    }

    var ret strings.Builder
    ret.WriteString("select p_")
    ret.WriteString(name)
    ret.WriteString(", count(*) from player_tb ")

    if p.Index != "" {
        ret.WriteString(" use index(")
        ret.WriteString(p.Index)
        ret.WriteString(")")
    }

    ret.WriteString(" where 1=1 ")
    

    return ret.String()
}

func (p *PlayerManager) Truncate() error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }
    
    query := "truncate player_tb "
    _, err := p.Exec(query)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return nil
}

func (p *PlayerManager) Insert(item *Player) error {
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
        query = "insert into player_tb (p_id, p_team, p_name, p_number, p_position, p_createddate, p_updateddate) values (?, ?, ?, ?, ?, ?, ?)"
        res, err = p.Exec(query, item.Id, item.Team, item.Name, item.Number, item.Position, item.Createddate, item.Updateddate)
    } else {
        query = "insert into player_tb (p_team, p_name, p_number, p_position, p_createddate, p_updateddate) values (?, ?, ?, ?, ?, ?)"
        res, err = p.Exec(query, item.Team, item.Name, item.Number, item.Position, item.Createddate, item.Updateddate)
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

func (p *PlayerManager) Delete(id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "delete from player_tb where p_id = ?"
    _, err := p.Exec(query, id)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    
    return err
}

func (p *PlayerManager) DeleteAll() error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "delete from player_tb"
    _, err := p.Exec(query)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return err
}

func (p *PlayerManager) MakeQuery(initQuery string , postQuery string, initParams []interface{}, args []interface{}) (string, []interface{}) {
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
                query.WriteString(" and p_")
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

func (p *PlayerManager) DeleteWhere(args []interface{}) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query, params := p.MakeQuery("delete from player_tb where 1=1", "", nil, args)
    _, err := p.Exec(query, params...)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return err
}

func (p *PlayerManager) Update(item *Player) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }
    
    
    if item.Createddate == "" {
       item.Createddate = "1000-01-01 00:00:00"
    }
	
    if item.Updateddate == "" {
       item.Updateddate = "1000-01-01 00:00:00"
    }
	

	query := "update player_tb set p_team = ?, p_name = ?, p_number = ?, p_position = ?, p_createddate = ?, p_updateddate = ? where p_id = ?"
	_, err := p.Exec(query, item.Team, item.Name, item.Number, item.Position, item.Createddate, item.Updateddate, item.Id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }
    
        
    return err
}

func (p *PlayerManager) UpdateWhere(columns []player.Params, args []interface{}) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    var initQuery strings.Builder
    var initParams []interface{}

    initQuery.WriteString("update player_tb set ")
    for i, v := range columns {
        if i > 0 {
            initQuery.WriteString(", ")
        }

        if v.Column == player.ColumnId {
        initQuery.WriteString("p_id = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == player.ColumnTeam {
        initQuery.WriteString("p_team = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == player.ColumnName {
        initQuery.WriteString("p_name = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == player.ColumnNumber {
        initQuery.WriteString("p_number = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == player.ColumnPosition {
        initQuery.WriteString("p_position = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == player.ColumnCreateddate {
        initQuery.WriteString("p_createddate = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == player.ColumnUpdateddate {
        initQuery.WriteString("p_updateddate = ?")
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


func (p *PlayerManager) UpdateTeam(value int, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update player_tb set p_team = ? where p_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *PlayerManager) UpdateName(value string, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update player_tb set p_name = ? where p_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *PlayerManager) UpdateNumber(value int, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update player_tb set p_number = ? where p_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *PlayerManager) UpdatePosition(value string, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update player_tb set p_position = ? where p_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *PlayerManager) UpdateCreateddate(value string, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update player_tb set p_createddate = ? where p_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}

func (p *PlayerManager) UpdateUpdateddate(value string, id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

	query := "update player_tb set p_updateddate = ? where p_id = ?"
	_, err := p.Exec(query, value, id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }

    return err
}


*/

func (p *PlayerManager) GetIdentity() int64 {
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

func (p *Player) InitExtra() {
    p.Extra = map[string]interface{}{

    }
}

func (p *PlayerManager) ReadRow(rows *sql.Rows) *Player {
    var item Player
    var err error

    

    if rows.Next() {
        err = rows.Scan(&item.Id, &item.Team, &item.Name, &item.Number, &item.Position, &item.Createddate, &item.Updateddate)
        
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

func (p *PlayerManager) ReadRows(rows *sql.Rows) []Player {
    var items []Player

    for rows.Next() {
        var item Player
        

        err := rows.Scan(&item.Id, &item.Team, &item.Name, &item.Number, &item.Position, &item.Createddate, &item.Updateddate)
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

func (p *PlayerManager) Get(id int64) *Player {
    if !p.Conn.IsConnect() {
        return nil
    }

    var query strings.Builder
    query.WriteString(p.GetQuery())
    query.WriteString(" and p_id = ?")

    
    
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

func (p *PlayerManager) GetWhere(args []interface{}) *Player {
    items := p.Find(args)
    if len(items) == 0 {
        return nil
    }

    return &items[0]
}

func (p *PlayerManager) Count(args []interface{}) int {
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

func (p *PlayerManager) FindAll() []Player {
    return p.Find(nil)
}

func (p *PlayerManager) Find(args []interface{}) []Player {
    if !p.Conn.IsConnect() {
        var items []Player
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
                query.WriteString(" and p_")
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
            orderby = "p_id desc"
        } else {
            if !strings.Contains(orderby, "_") {                   
                if strings.ToUpper(orderby) != "RAND()" {
                    orderby = "p_" + orderby
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
            orderby = "p_id"
        } else {
            if !strings.Contains(orderby, "_") {
                if strings.ToUpper(orderby) != "RAND()" {
                    orderby = "p_" + orderby
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
        items := make([]Player, 0)
        return items
    }

    defer rows.Close()

    return p.ReadRows(rows)
}





func (p *PlayerManager) GroupBy(name string, args []interface{}) []Groupby {
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
                query.WriteString(" and p_")
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
    
    query.WriteString(" group by p_")
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
