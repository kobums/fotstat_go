package models

import (
    "fotstat/global/config"
    "fotstat/models/inbody"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"

    log "fotstat/global/log"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"

)

type Inbody struct {

    Id                int64 `json:"id"`
    Player                int `json:"player"`
    Testdate                string `json:"testdate"`
    Height                float64 `json:"height"`
    Weight                float64 `json:"weight"`
    Muscle                float64 `json:"muscle"`
    Fat                float64 `json:"fat"`
    Rightleg                float64 `json:"rightleg"`
    Leftleg                float64 `json:"leftleg"`
    Score                int `json:"score"`
    Createddate                string `json:"createddate"`
    Updateddate                string `json:"updateddate"`

    Extra                    map[string]interface{} `json:"extra"`
}

// nullableFloat maps 0 to SQL NULL — 인바디 측정치는 전부 양수라 0 은 "미측정"을 뜻한다.
// (API 응답에서는 반대로 NULL 이 0 으로 내려간다.)
func nullableFloat(value float64) interface{} {
    if value == 0 {
        return nil
    }
    return value
}

// nullableInt maps 0 to SQL NULL — nullableFloat 와 같은 규칙의 정수 컬럼용.
func nullableInt(value int) interface{} {
    if value == 0 {
        return nil
    }
    return value
}

// nullableFloatParam / nullableIntParam 은 UpdateWhere 처럼 interface{} 로
// 값을 받는 경로에서 0 = NULL 계약을 유지한다. 타입이 다르면 그대로 통과.
func nullableFloatParam(value interface{}) interface{} {
    if f, ok := value.(float64); ok {
        return nullableFloat(f)
    }
    return value
}

func nullableIntParam(value interface{}) interface{} {
    if n, ok := value.(int); ok {
        return nullableInt(n)
    }
    return value
}

type InbodyManager struct {
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

func (c *Inbody) AddExtra(key string, value interface{}) {    
	c.Extra[key] = value     
}

func NewInbodyManager(conn *Connection) *InbodyManager {
    var item InbodyManager


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

func (p *InbodyManager) Close() {
    if p.Conn != nil {
        p.Conn.Close()
    }
}

func (p *InbodyManager) SetIndex(index string) {
    p.Index = index
}

func (p *InbodyManager) SetCountQuery(query string) {
    p.CountQuery = query
}

func (p *InbodyManager) SetSelectQuery(query string) {
    p.SelectQuery = query
}

func (p *InbodyManager) Exec(query string, params ...interface{}) (sql.Result, error) {
    if p.Log {
       if len(params) > 0 {
	       log.Debug().Str("query", query).Any("param", params).Msg("SQL")
       } else {
	       log.Debug().Str("query", query).Msg("SQL")
       }
    }

    return p.Conn.Exec(query, params...)
}

func (p *InbodyManager) Query(query string, params ...interface{}) (*sql.Rows, error) {
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

func (p *InbodyManager) GetQuery() string {
    if p.SelectQuery != "" {
        return p.SelectQuery    
    }

    var ret strings.Builder

    ret.WriteString("select ib_id, ib_player, ib_testdate, ib_height, ib_weight, ib_muscle, ib_fat, ib_rightleg, ib_leftleg, ib_score, ib_createddate, ib_updateddate from inbody_tb")

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

func (p *InbodyManager) GetQuerySelect() string {
    if p.CountQuery != "" {
        return p.CountQuery    
    }

    var ret strings.Builder
    
    ret.WriteString("select count(*) from inbody_tb")

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

func (p *InbodyManager) GetQueryGroup(name string) string {
    if p.SelectQuery != "" {
        return p.SelectQuery    
    }

    var ret strings.Builder
    ret.WriteString("select ib_")
    ret.WriteString(name)
    ret.WriteString(", count(*) from inbody_tb ")

    if p.Index != "" {
        ret.WriteString(" use index(")
        ret.WriteString(p.Index)
        ret.WriteString(")")
    }

    ret.WriteString(" where 1=1 ")
    

    return ret.String()
}

func (p *InbodyManager) Truncate() error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }
    
    query := "truncate inbody_tb "
    _, err := p.Exec(query)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return nil
}

func (p *InbodyManager) Insert(item *Inbody) error {
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
        query = "insert into inbody_tb (ib_id, ib_player, ib_testdate, ib_height, ib_weight, ib_muscle, ib_fat, ib_rightleg, ib_leftleg, ib_score, ib_createddate, ib_updateddate) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
        res, err = p.Exec(query, item.Id, item.Player, nullableDate(item.Testdate), nullableFloat(item.Height), nullableFloat(item.Weight), nullableFloat(item.Muscle), nullableFloat(item.Fat), nullableFloat(item.Rightleg), nullableFloat(item.Leftleg), nullableInt(item.Score), item.Createddate, item.Updateddate)
    } else {
        query = "insert into inbody_tb (ib_player, ib_testdate, ib_height, ib_weight, ib_muscle, ib_fat, ib_rightleg, ib_leftleg, ib_score, ib_createddate, ib_updateddate) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
        res, err = p.Exec(query, item.Player, nullableDate(item.Testdate), nullableFloat(item.Height), nullableFloat(item.Weight), nullableFloat(item.Muscle), nullableFloat(item.Fat), nullableFloat(item.Rightleg), nullableFloat(item.Leftleg), nullableInt(item.Score), item.Createddate, item.Updateddate)
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

// Upsert 는 (ib_player, ib_testdate) UNIQUE 키 기준으로 없으면 삽입, 있으면
// 기존 행의 측정 컬럼 전부를 values() 로 덮어쓴다 (migration_013). 시트 일괄
// 입력을 재저장해도 행이 늘어나지 않고 마지막 값으로 수렴하며, 값을 지우고
// 저장하면 해당 컬럼이 NULL 로 갱신된다.
// last_insert_id(ib_id) 트릭으로 갱신 경로에서도 GetIdentity 가 기존 행 id 를 돌려준다.
func (p *InbodyManager) Upsert(item *Inbody) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    if item.Createddate == "" {
        t := time.Now().UTC().Add(time.Hour * 9)
        item.Createddate = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
    }

    if item.Updateddate == "" {
       item.Updateddate = "1000-01-01 00:00:00"
    }

    query := "insert into inbody_tb (ib_player, ib_testdate, ib_height, ib_weight, ib_muscle, ib_fat, ib_rightleg, ib_leftleg, ib_score, ib_createddate, ib_updateddate) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)" +
        " on duplicate key update" +
        " ib_id = last_insert_id(ib_id)," +
        " ib_height = values(ib_height)," +
        " ib_weight = values(ib_weight)," +
        " ib_muscle = values(ib_muscle)," +
        " ib_fat = values(ib_fat)," +
        " ib_rightleg = values(ib_rightleg)," +
        " ib_leftleg = values(ib_leftleg)," +
        " ib_score = values(ib_score)," +
        " ib_updateddate = current_timestamp"
    res, err := p.Exec(query, item.Player, nullableDate(item.Testdate), nullableFloat(item.Height), nullableFloat(item.Weight), nullableFloat(item.Muscle), nullableFloat(item.Fat), nullableFloat(item.Rightleg), nullableFloat(item.Leftleg), nullableInt(item.Score), item.Createddate, item.Updateddate)

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

func (p *InbodyManager) Delete(id int64) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "delete from inbody_tb where ib_id = ?"
    _, err := p.Exec(query, id)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    
    return err
}

func (p *InbodyManager) DeleteAll() error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query := "delete from inbody_tb"
    _, err := p.Exec(query)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return err
}

func (p *InbodyManager) MakeQuery(initQuery string , postQuery string, initParams []interface{}, args []interface{}) (string, []interface{}) {
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
                query.WriteString(" and ib_")
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

func (p *InbodyManager) DeleteWhere(args []interface{}) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    query, params := p.MakeQuery("delete from inbody_tb where 1=1", "", nil, args)
    _, err := p.Exec(query, params...)

    if err != nil {
       if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
       }
    }

    return err
}

func (p *InbodyManager) Update(item *Inbody) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }
    
    
    if item.Createddate == "" {
       item.Createddate = "1000-01-01 00:00:00"
    }
	
    if item.Updateddate == "" {
       item.Updateddate = "1000-01-01 00:00:00"
    }
	

	query := "update inbody_tb set ib_player = ?, ib_testdate = ?, ib_height = ?, ib_weight = ?, ib_muscle = ?, ib_fat = ?, ib_rightleg = ?, ib_leftleg = ?, ib_score = ?, ib_createddate = ?, ib_updateddate = ? where ib_id = ?"
	_, err := p.Exec(query, item.Player, nullableDate(item.Testdate), nullableFloat(item.Height), nullableFloat(item.Weight), nullableFloat(item.Muscle), nullableFloat(item.Fat), nullableFloat(item.Rightleg), nullableFloat(item.Leftleg), nullableInt(item.Score), item.Createddate, item.Updateddate, item.Id)

    if err != nil {
        if p.Log {
          log.Error().Str("error", err.Error()).Msg("SQL")
        }
    }
    
        
    return err
}

func (p *InbodyManager) UpdateWhere(columns []inbody.Params, args []interface{}) error {
    if !p.Conn.IsConnect() {
        return errors.New("Connection Error")
    }

    var initQuery strings.Builder
    var initParams []interface{}

    initQuery.WriteString("update inbody_tb set ")
    for i, v := range columns {
        if i > 0 {
            initQuery.WriteString(", ")
        }

        if v.Column == inbody.ColumnId {
        initQuery.WriteString("ib_id = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == inbody.ColumnPlayer {
        initQuery.WriteString("ib_player = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == inbody.ColumnTestdate {
        initQuery.WriteString("ib_testdate = ?")
        initParams = append(initParams, nullableDate(fmt.Sprint(v.Value)))
        } else if v.Column == inbody.ColumnHeight {
        initQuery.WriteString("ib_height = ?")
        initParams = append(initParams, nullableFloatParam(v.Value))
        } else if v.Column == inbody.ColumnWeight {
        initQuery.WriteString("ib_weight = ?")
        initParams = append(initParams, nullableFloatParam(v.Value))
        } else if v.Column == inbody.ColumnMuscle {
        initQuery.WriteString("ib_muscle = ?")
        initParams = append(initParams, nullableFloatParam(v.Value))
        } else if v.Column == inbody.ColumnFat {
        initQuery.WriteString("ib_fat = ?")
        initParams = append(initParams, nullableFloatParam(v.Value))
        } else if v.Column == inbody.ColumnRightleg {
        initQuery.WriteString("ib_rightleg = ?")
        initParams = append(initParams, nullableFloatParam(v.Value))
        } else if v.Column == inbody.ColumnLeftleg {
        initQuery.WriteString("ib_leftleg = ?")
        initParams = append(initParams, nullableFloatParam(v.Value))
        } else if v.Column == inbody.ColumnScore {
        initQuery.WriteString("ib_score = ?")
        initParams = append(initParams, nullableIntParam(v.Value))
        } else if v.Column == inbody.ColumnCreateddate {
        initQuery.WriteString("ib_createddate = ?")
        initParams = append(initParams, v.Value)
        } else if v.Column == inbody.ColumnUpdateddate {
        initQuery.WriteString("ib_updateddate = ?")
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


func (p *InbodyManager) GetIdentity() int64 {
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

func (p *Inbody) InitExtra() {
    p.Extra = map[string]interface{}{

    }
}

func (p *InbodyManager) ReadRow(rows *sql.Rows) *Inbody {
    var item Inbody
    var err error

    

    if rows.Next() {
        var testdate sql.NullString
        var height, weight, muscle, fat, rightleg, leftleg sql.NullFloat64
        var score sql.NullInt64
        err = rows.Scan(&item.Id, &item.Player, &testdate, &height, &weight, &muscle, &fat, &rightleg, &leftleg, &score, &item.Createddate, &item.Updateddate)
        item.Testdate = testdate.String
        item.Height = height.Float64
        item.Weight = weight.Float64
        item.Muscle = muscle.Float64
        item.Fat = fat.Float64
        item.Rightleg = rightleg.Float64
        item.Leftleg = leftleg.Float64
        item.Score = int(score.Int64)

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

func (p *InbodyManager) ReadRows(rows *sql.Rows) []Inbody {
    var items []Inbody

    for rows.Next() {
        var item Inbody
        var testdate sql.NullString
        var height, weight, muscle, fat, rightleg, leftleg sql.NullFloat64
        var score sql.NullInt64

        err := rows.Scan(&item.Id, &item.Player, &testdate, &height, &weight, &muscle, &fat, &rightleg, &leftleg, &score, &item.Createddate, &item.Updateddate)
        if err != nil {
           if p.Log {
             log.Error().Str("error", err.Error()).Msg("SQL")
           }
           break
        }
        item.Testdate = testdate.String
        item.Height = height.Float64
        item.Weight = weight.Float64
        item.Muscle = muscle.Float64
        item.Fat = fat.Float64
        item.Rightleg = rightleg.Float64
        item.Leftleg = leftleg.Float64
        item.Score = int(score.Int64)

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

func (p *InbodyManager) Get(id int64) *Inbody {
    if !p.Conn.IsConnect() {
        return nil
    }

    var query strings.Builder
    query.WriteString(p.GetQuery())
    query.WriteString(" and ib_id = ?")

    
    
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

func (p *InbodyManager) GetWhere(args []interface{}) *Inbody {
    items := p.Find(args)
    if len(items) == 0 {
        return nil
    }

    return &items[0]
}

func (p *InbodyManager) Count(args []interface{}) int {
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

func (p *InbodyManager) FindAll() []Inbody {
    return p.Find(nil)
}

func (p *InbodyManager) Find(args []interface{}) []Inbody {
    if !p.Conn.IsConnect() {
        var items []Inbody
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
                query.WriteString(" and ib_")
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
            orderby = "ib_id desc"
        } else {
            if !strings.Contains(orderby, "_") {                   
                if strings.ToUpper(orderby) != "RAND()" {
                    orderby = "ib_" + orderby
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
            orderby = "ib_id"
        } else {
            if !strings.Contains(orderby, "_") {
                if strings.ToUpper(orderby) != "RAND()" {
                    orderby = "ib_" + orderby
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
        items := make([]Inbody, 0)
        return items
    }

    defer rows.Close()

    return p.ReadRows(rows)
}





func (p *InbodyManager) GroupBy(name string, args []interface{}) []Groupby {
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
                query.WriteString(" and ib_")
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
    
    query.WriteString(" group by ib_")
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
