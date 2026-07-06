package rest


import (
	"fotstat/controllers"

	"fotstat/models"
	"fotstat/models/record"

	"errors"
	"fmt"
	"strings"
)

type RecordController struct {
	controllers.Controller
}

// injuryConflict reports whether the given player is injured on the match date
// that the given quarter belongs to. Record 입력은 부상 기간에 걸치면 차단하되,
// 발생일 당일 경기는 허용한다(경기 중 부상 = 그날까지는 뛴 것). 즉 차단 범위는
// 발생일 다음 날부터 복귀일까지(i_returndate NULL 이면 아직 부상 중 = 계속 차단).
// 반환된 error 가 nil 이 아니면 그 선수는 해당 경기일에 부상 중이다.
func (c *RecordController) injuryConflict(conn *models.Connection, quarter int, player int) error {
	if quarter == 0 || player == 0 {
		return nil
	}

	q := models.NewQuarterManager(conn).Get(int64(quarter))
	if q == nil {
		return nil
	}

	m := models.NewMatchManager(conn).Get(int64(q.Match))
	if m == nil || m.Matchdate == "" {
		return nil
	}

	// m_matchdate 는 DATETIME, injury 날짜는 DATE 이므로 날짜 부분만 비교한다.
	matchdate := m.Matchdate
	if len(matchdate) >= 10 {
		matchdate = matchdate[:10]
	}

	injuryManager := models.NewInjuryManager(conn)
	cnt := injuryManager.Count([]interface{}{
		models.Where{Column: "player", Value: player, Compare: "="},
		models.Custom{Query: fmt.Sprintf("i_startdate < '%s'", matchdate)},
		models.Custom{Query: fmt.Sprintf("(i_returndate is null or i_returndate >= '%s')", matchdate)},
	})

	if cnt > 0 {
		return errors.New("injured player cannot be recorded for this match")
	}

	return nil
}

func (c *RecordController) Read(id int64) {
    
    
	conn := c.NewConnection()

	manager := models.NewRecordManager(conn)
	item := manager.Get(id)

    if item != nil && !ownsQuarter(conn, requestUser(&c.Controller), item.Quarter) {
        c.Error(errForbidden)
        return
    }

    c.Set("item", item)
}

func (c *RecordController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewRecordManager(conn)

    var args []interface{}

    // 소유권 강제: 요청 사용자 소유 팀의 기록만 조회된다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownRecordScope(user))

    _quarter := c.Geti("quarter")
    if _quarter != 0 {
        args = append(args, models.Where{Column:"quarter", Value:_quarter, Compare:"="})
    }
    _player := c.Geti("player")
    if _player != 0 {
        args = append(args, models.Where{Column:"player", Value:_player, Compare:"="})    
    }
    _min := c.Geti("min")
    if _min != 0 {
        args = append(args, models.Where{Column:"min", Value:_min, Compare:"="})    
    }
    _goal := c.Geti("goal")
    if _goal != 0 {
        args = append(args, models.Where{Column:"goal", Value:_goal, Compare:"="})    
    }
    _assist := c.Geti("assist")
    if _assist != 0 {
        args = append(args, models.Where{Column:"assist", Value:_assist, Compare:"="})    
    }
    _startcreateddate := c.Get("startcreateddate")
    _endcreateddate := c.Get("endcreateddate")
    if _startcreateddate != "" && _endcreateddate != "" {        
        var v [2]string
        v[0] = _startcreateddate
        v[1] = _endcreateddate  
        args = append(args, models.Where{Column:"createddate", Value:v, Compare:"between"})    
    } else if  _startcreateddate != "" {          
        args = append(args, models.Where{Column:"createddate", Value:_startcreateddate, Compare:">="})
    } else if  _endcreateddate != "" {          
        args = append(args, models.Where{Column:"createddate", Value:_endcreateddate, Compare:"<="})            
    }
    _startupdateddate := c.Get("startupdateddate")
    _endupdateddate := c.Get("endupdateddate")
    if _startupdateddate != "" && _endupdateddate != "" {        
        var v [2]string
        v[0] = _startupdateddate
        v[1] = _endupdateddate  
        args = append(args, models.Where{Column:"updateddate", Value:v, Compare:"between"})    
    } else if  _startupdateddate != "" {          
        args = append(args, models.Where{Column:"updateddate", Value:_startupdateddate, Compare:">="})
    } else if  _endupdateddate != "" {          
        args = append(args, models.Where{Column:"updateddate", Value:_endupdateddate, Compare:"<="})            
    }
    

    
    
    if page != 0 && pagesize != 0 {
        args = append(args, models.Paging(page, pagesize))
    }
    
    orderby := c.Get("orderby")
    if orderby == "" {
        if page != 0 && pagesize != 0 {
            orderby = "id desc"
            args = append(args, models.Ordering(orderby))
        }
    } else {
        orderbys := strings.Split(orderby, ",")

        str := ""
        for i, v := range orderbys {
            if i == 0 {
                str += v
            } else {
                if strings.Contains(v, "_") {                   
                    str += ", " + strings.Trim(v, " ")
                } else {
                    str += ", r_" + strings.Trim(v, " ")                
                }
            }
        }
        
        args = append(args, models.Ordering(str))
    }
    
	items := manager.Find(args)
	c.Set("items", items)

    if page == 1 {
       total := manager.Count(args)
	   c.Set("total", total)
    }
}

func (c *RecordController) Count() {
    
    
	conn := c.NewConnection()

	manager := models.NewRecordManager(conn)

    var args []interface{}

    // 소유권 강제: 요청 사용자 소유 팀의 기록만 조회된다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownRecordScope(user))

    _quarter := c.Geti("quarter")
    if _quarter != 0 {
        args = append(args, models.Where{Column:"quarter", Value:_quarter, Compare:"="})
    }
    _player := c.Geti("player")
    if _player != 0 {
        args = append(args, models.Where{Column:"player", Value:_player, Compare:"="})    
    }
    _min := c.Geti("min")
    if _min != 0 {
        args = append(args, models.Where{Column:"min", Value:_min, Compare:"="})    
    }
    _goal := c.Geti("goal")
    if _goal != 0 {
        args = append(args, models.Where{Column:"goal", Value:_goal, Compare:"="})    
    }
    _assist := c.Geti("assist")
    if _assist != 0 {
        args = append(args, models.Where{Column:"assist", Value:_assist, Compare:"="})    
    }
    _startcreateddate := c.Get("startcreateddate")
    _endcreateddate := c.Get("endcreateddate")

    if _startcreateddate != "" && _endcreateddate != "" {        
        var v [2]string
        v[0] = _startcreateddate
        v[1] = _endcreateddate  
        args = append(args, models.Where{Column:"createddate", Value:v, Compare:"between"})    
    } else if  _startcreateddate != "" {          
        args = append(args, models.Where{Column:"createddate", Value:_startcreateddate, Compare:">="})
    } else if  _endcreateddate != "" {          
        args = append(args, models.Where{Column:"createddate", Value:_endcreateddate, Compare:"<="})            
    }
    _startupdateddate := c.Get("startupdateddate")
    _endupdateddate := c.Get("endupdateddate")

    if _startupdateddate != "" && _endupdateddate != "" {        
        var v [2]string
        v[0] = _startupdateddate
        v[1] = _endupdateddate  
        args = append(args, models.Where{Column:"updateddate", Value:v, Compare:"between"})    
    } else if  _startupdateddate != "" {          
        args = append(args, models.Where{Column:"updateddate", Value:_startupdateddate, Compare:">="})
    } else if  _endupdateddate != "" {          
        args = append(args, models.Where{Column:"updateddate", Value:_endupdateddate, Compare:"<="})            
    }
    
    
    
    
    total := manager.Count(args)
	c.Set("total", total)
}

func (c *RecordController) Insert(item *models.Record) {
    
    
    

	conn := c.NewConnection()

    // 내 소유 팀의 쿼터에, 그 경기 팀 소속 선수의 기록만 생성 가능
    user := requestUser(&c.Controller)
    if !ownsRecordTarget(conn, user, item.Quarter, item.Player) {
        c.Error(errForbidden)
        return
    }

    if err := c.injuryConflict(conn, item.Quarter, item.Player); err != nil {
        c.Set("code", "error")
        c.Set("error", err.Error())
        return
    }

	manager := models.NewRecordManager(conn)
	err := manager.Insert(item)
    if err != nil {
        c.Set("code", "error")
        c.Set("error", err)
        return
    }

    id := manager.GetIdentity()
    c.Result["id"] = id
    item.Id = id
}

func (c *RecordController) Insertbatch(item *[]models.Record) {  
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)
    
    
    
	conn := c.NewConnection()
    
	manager := models.NewRecordManager(conn)

    // 전량 사전 검증(소유권·팀 일치 + 부상 충돌) 후 일괄 삽입
    user := requestUser(&c.Controller)
    for i := 0; i < rows; i++ {
        if !ownsRecordTarget(conn, user, (*item)[i].Quarter, (*item)[i].Player) {
            c.Error(errForbidden)
            return
        }
        if err := c.injuryConflict(conn, (*item)[i].Quarter, (*item)[i].Player); err != nil {
            c.Set("code", "error")
            c.Set("error", err.Error())
            return
        }
    }

    for i := 0; i < rows; i++ {

	    err := manager.Insert(&((*item)[i]))
        if err != nil {
            c.Set("code", "error")
            c.Set("error", err)
            return
        }
    }
}

func (c *RecordController) Update(item *models.Record) {




	conn := c.NewConnection()

	manager := models.NewRecordManager(conn)

    // 기존 기록과 변경 후 값(쿼터·선수) 모두 내 소유여야 한다
    user := requestUser(&c.Controller)
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errNotFound)
        return
    }
    if !ownsQuarter(conn, user, existing.Quarter) ||
        !ownsRecordTarget(conn, user, item.Quarter, item.Player) {
        c.Error(errForbidden)
        return
    }

    if err := c.injuryConflict(conn, item.Quarter, item.Player); err != nil {
        c.Set("code", "error")
        c.Set("error", err.Error())
        return
    }

    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }
}

func (c *RecordController) UpdateStats(item *models.Record) {
	conn := c.NewConnection()

	// 부상 기간 중인 선수는 기존 기록 수정도 차단(입력과 동일 정책).
	// body에는 quarter/player가 없으므로 기존 record를 id로 조회해 검증한다.
	existing := models.NewRecordManager(conn).Get(item.Id)
	if existing == nil {
		c.Error(errNotFound)
		return
	}
	if !ownsQuarter(conn, requestUser(&c.Controller), existing.Quarter) {
		c.Error(errForbidden)
		return
	}
	if err := c.injuryConflict(conn, existing.Quarter, existing.Player); err != nil {
		c.Set("code", "error")
		c.Set("error", err.Error())
		return
	}

	manager := models.NewRecordManager(conn)
	err := manager.UpdateWhere(
		[]record.Params{
			{Column: record.ColumnMin, Value: item.Min},
			{Column: record.ColumnGoal, Value: item.Goal},
			{Column: record.ColumnAssist, Value: item.Assist},
			{Column: record.ColumnYellowcard, Value: item.Yellowcard},
			{Column: record.ColumnRedcard, Value: item.Redcard},
		},
		[]interface{}{models.Where{Column: "id", Value: item.Id, Compare: "="}},
	)
	if err != nil {
		c.Set("code", "error")
		c.Set("error", err)
	}
}

func (c *RecordController) Delete(item *models.Record) {


    conn := c.NewConnection()

	manager := models.NewRecordManager(conn)

    existing := manager.Get(item.Id)
    if existing == nil {
        return   // 이미 없음 — 멱등 처리
    }
    if !ownsQuarter(conn, requestUser(&c.Controller), existing.Quarter) {
        c.Error(errForbidden)
        return
    }

	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
    }
}

func (c *RecordController) Deletebatch(item *[]models.Record) {
    
    
    conn := c.NewConnection()

	manager := models.NewRecordManager(conn)

    // 전량 사전 검증 후 일괄 삭제
    user := requestUser(&c.Controller)
    for _, v := range *item {
        existing := manager.Get(v.Id)
        if existing == nil {
            continue
        }
        if !ownsQuarter(conn, user, existing.Quarter) {
            c.Error(errForbidden)
            return
        }
    }

    for _, v := range *item {


	    err := manager.Delete(v.Id)
        if err != nil {
            c.Set("code", "error")
            c.Set("error", err)
            return
        }
    }
}


