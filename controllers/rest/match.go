package rest


import (
	"fotstat/controllers"
	
	"fotstat/models"

    "strings"
)

type MatchController struct {
	controllers.Controller
}

func (c *MatchController) Read(id int64) {
    
    
	conn := c.NewConnection()

	manager := models.NewMatchManager(conn)
	item := manager.Get(id)

    if item != nil && !ownsTeam(conn, requestUser(&c.Controller), item.Team) {
        c.Error(errForbidden)
        return
    }

    c.Set("item", item)
}

func (c *MatchController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewMatchManager(conn)

    var args []interface{}

    // 소유권 강제: 요청 사용자 소유 팀의 경기만 조회된다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownMatchScope(user))

    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Where{Column:"team", Value:_team, Compare:"="})
    }
    _awayname := c.Get("awayname")
    if _awayname != "" {
        args = append(args, models.Where{Column:"awayname", Value:_awayname, Compare:"like"})
    }
    _startmatchdate := c.Get("startmatchdate")
    _endmatchdate := c.Get("endmatchdate")
    if _startmatchdate != "" && _endmatchdate != "" {        
        var v [2]string
        v[0] = _startmatchdate
        v[1] = _endmatchdate  
        args = append(args, models.Where{Column:"matchdate", Value:v, Compare:"between"})    
    } else if  _startmatchdate != "" {          
        args = append(args, models.Where{Column:"matchdate", Value:_startmatchdate, Compare:">="})
    } else if  _endmatchdate != "" {          
        args = append(args, models.Where{Column:"matchdate", Value:_endmatchdate, Compare:"<="})            
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
                    str += ", m_" + strings.Trim(v, " ")                
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

func (c *MatchController) Count() {
    
    
	conn := c.NewConnection()

	manager := models.NewMatchManager(conn)

    var args []interface{}

    // 소유권 강제: 요청 사용자 소유 팀의 경기만 조회된다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownMatchScope(user))

    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Where{Column:"team", Value:_team, Compare:"="})
    }
    _awayname := c.Get("awayname")
    if _awayname != "" {
        args = append(args, models.Where{Column:"awayname", Value:_awayname, Compare:"like"})
        
    }
    _startmatchdate := c.Get("startmatchdate")
    _endmatchdate := c.Get("endmatchdate")

    if _startmatchdate != "" && _endmatchdate != "" {        
        var v [2]string
        v[0] = _startmatchdate
        v[1] = _endmatchdate  
        args = append(args, models.Where{Column:"matchdate", Value:v, Compare:"between"})    
    } else if  _startmatchdate != "" {          
        args = append(args, models.Where{Column:"matchdate", Value:_startmatchdate, Compare:">="})
    } else if  _endmatchdate != "" {          
        args = append(args, models.Where{Column:"matchdate", Value:_endmatchdate, Compare:"<="})            
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

func (c *MatchController) Insert(item *models.Match) {

	conn := c.NewConnection()

    // 내 소유 팀의 경기만 생성 가능
    if !ownsTeam(conn, requestUser(&c.Controller), item.Team) {
        c.Error(errForbidden)
        return
    }

	manager := models.NewMatchManager(conn)
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

func (c *MatchController) Insertbatch(item *[]models.Match) {  
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)

	conn := c.NewConnection()

	manager := models.NewMatchManager(conn)

    // 전량 사전 검증 후 일괄 삽입
    user := requestUser(&c.Controller)
    for i := 0; i < rows; i++ {
        if !ownsTeam(conn, user, (*item)[i].Team) {
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

func (c *MatchController) Update(item *models.Match) {

	conn := c.NewConnection()

	manager := models.NewMatchManager(conn)

    // 기존 경기의 팀과 변경 후 팀 모두 내 소유여야 한다
    user := requestUser(&c.Controller)
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errNotFound)
        return
    }
    if !ownsTeam(conn, user, existing.Team) || !ownsTeam(conn, user, item.Team) {
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

func (c *MatchController) Delete(item *models.Match) {


    conn := c.NewConnection()

	manager := models.NewMatchManager(conn)

    existing := manager.Get(item.Id)
    if existing == nil {
        return   // 이미 없음 — 멱등 처리
    }
    if !ownsTeam(conn, requestUser(&c.Controller), existing.Team) {
        c.Error(errForbidden)
        return
    }

	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
    }
}

func (c *MatchController) Deletebatch(item *[]models.Match) {
    
    
    conn := c.NewConnection()

	manager := models.NewMatchManager(conn)

    // 전량 사전 검증 후 일괄 삭제
    user := requestUser(&c.Controller)
    for _, v := range *item {
        existing := manager.Get(v.Id)
        if existing == nil {
            continue
        }
        if !ownsTeam(conn, user, existing.Team) {
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


