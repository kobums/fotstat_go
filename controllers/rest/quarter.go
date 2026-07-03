package rest


import (
	"fotstat/controllers"

	"fotstat/models"
	"fotstat/models/quarter"

	"strings"
)

type QuarterController struct {
	controllers.Controller
}

func (c *QuarterController) Read(id int64) {
    
    
	conn := c.NewConnection()

	manager := models.NewQuarterManager(conn)
	item := manager.Get(id)

    if item != nil && !ownsMatch(conn, requestUser(&c.Controller), item.Match) {
        c.Error(errForbidden)
        return
    }

    c.Set("item", item)
}

func (c *QuarterController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewQuarterManager(conn)

    var args []interface{}

    // 소유권 강제: 요청 사용자 소유 팀의 경기 쿼터만 조회된다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownQuarterScope(user))

    _match := c.Geti("match")
    if _match != 0 {
        args = append(args, models.Where{Column:"match", Value:_match, Compare:"="})
    }
    _number := c.Geti("number")
    if _number != 0 {
        args = append(args, models.Where{Column:"number", Value:_number, Compare:"="})    
    }
    _duration := c.Geti("duration")
    if _duration != 0 {
        args = append(args, models.Where{Column:"duration", Value:_duration, Compare:"="})    
    }
    _awaygoals := c.Geti("awaygoals")
    if _awaygoals != 0 {
        args = append(args, models.Where{Column:"awaygoals", Value:_awaygoals, Compare:"="})    
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
                    str += ", q_" + strings.Trim(v, " ")                
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

func (c *QuarterController) Count() {
    
    
	conn := c.NewConnection()

	manager := models.NewQuarterManager(conn)

    var args []interface{}

    // 소유권 강제: 요청 사용자 소유 팀의 경기 쿼터만 조회된다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownQuarterScope(user))

    _match := c.Geti("match")
    if _match != 0 {
        args = append(args, models.Where{Column:"match", Value:_match, Compare:"="})
    }
    _number := c.Geti("number")
    if _number != 0 {
        args = append(args, models.Where{Column:"number", Value:_number, Compare:"="})    
    }
    _duration := c.Geti("duration")
    if _duration != 0 {
        args = append(args, models.Where{Column:"duration", Value:_duration, Compare:"="})    
    }
    _awaygoals := c.Geti("awaygoals")
    if _awaygoals != 0 {
        args = append(args, models.Where{Column:"awaygoals", Value:_awaygoals, Compare:"="})    
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

func (c *QuarterController) Insert(item *models.Quarter) {

	conn := c.NewConnection()

    // 내 소유 팀의 경기에만 쿼터 생성 가능
    if !ownsMatch(conn, requestUser(&c.Controller), item.Match) {
        c.Error(errForbidden)
        return
    }

	manager := models.NewQuarterManager(conn)
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

func (c *QuarterController) Insertbatch(item *[]models.Quarter) {  
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)

	conn := c.NewConnection()

	manager := models.NewQuarterManager(conn)

    // 전량 사전 검증 후 일괄 삽입
    user := requestUser(&c.Controller)
    for i := 0; i < rows; i++ {
        if !ownsMatch(conn, user, (*item)[i].Match) {
            c.Error(errForbidden)
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

func (c *QuarterController) Update(item *models.Quarter) {

	conn := c.NewConnection()

	manager := models.NewQuarterManager(conn)

    // 기존 쿼터의 경기와 변경 후 경기 모두 내 소유여야 한다
    user := requestUser(&c.Controller)
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errNotFound)
        return
    }
    if !ownsMatch(conn, user, existing.Match) || !ownsMatch(conn, user, item.Match) {
        c.Error(errForbidden)
        return
    }

    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }
}

func (c *QuarterController) UpdateAwaygoals(item *models.Quarter) {
	conn := c.NewConnection()
	manager := models.NewQuarterManager(conn)

    // 부분 업데이트도 기존 쿼터 기준으로 소유권 검증
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errNotFound)
        return
    }
    if !ownsMatch(conn, requestUser(&c.Controller), existing.Match) {
        c.Error(errForbidden)
        return
    }

	err := manager.UpdateWhere(
		[]quarter.Params{{Column: quarter.ColumnAwaygoals, Value: item.Awaygoals}},
		[]interface{}{models.Where{Column: "id", Value: item.Id, Compare: "="}},
	)
	if err != nil {
		c.Set("code", "error")
		c.Set("error", err)
	}
}

func (c *QuarterController) Delete(item *models.Quarter) {


    conn := c.NewConnection()

	manager := models.NewQuarterManager(conn)

    existing := manager.Get(item.Id)
    if existing == nil {
        return   // 이미 없음 — 멱등 처리
    }
    if !ownsMatch(conn, requestUser(&c.Controller), existing.Match) {
        c.Error(errForbidden)
        return
    }

	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
    }
}

func (c *QuarterController) Deletebatch(item *[]models.Quarter) {
    
    
    conn := c.NewConnection()

	manager := models.NewQuarterManager(conn)

    // 전량 사전 검증 후 일괄 삭제
    user := requestUser(&c.Controller)
    for _, v := range *item {
        existing := manager.Get(v.Id)
        if existing == nil {
            continue
        }
        if !ownsMatch(conn, user, existing.Match) {
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


