package rest


import (
	"fotstat/controllers"
	
	"fotstat/models"

    "strings"
)

type TeamController struct {
	controllers.Controller
}

func (c *TeamController) Read(id int64) {
    
    
	conn := c.NewConnection()

	manager := models.NewTeamManager(conn)
	item := manager.Get(id)

    if item != nil {
        user := requestUser(&c.Controller)
        if user == nil || int64(item.User) != user.Id {
            c.Error(errForbidden)
            return
        }
    }

    c.Set("item", item)
}

func (c *TeamController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewTeamManager(conn)

    var args []interface{}

    // 소유권 강제: 클라이언트 필터와 무관하게 요청 사용자 소유 팀만 조회된다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownTeamScope(user))

    _user := c.Geti("user")
    if _user != 0 {
        args = append(args, models.Where{Column:"user", Value:_user, Compare:"="})
    }
    _name := c.Get("name")
    if _name != "" {
        args = append(args, models.Where{Column:"name", Value:_name, Compare:"="})

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
                    str += ", t_" + strings.Trim(v, " ")                
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

func (c *TeamController) Count() {
    
    
	conn := c.NewConnection()

	manager := models.NewTeamManager(conn)

    var args []interface{}

    // 소유권 강제: Index 와 동일
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownTeamScope(user))

    _user := c.Geti("user")
    if _user != 0 {
        args = append(args, models.Where{Column:"user", Value:_user, Compare:"="})
    }
    _name := c.Get("name")
    if _name != "" {
        args = append(args, models.Where{Column:"name", Value:_name, Compare:"="})


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

func (c *TeamController) Insert(item *models.Team) {

    // 팀 소유자는 서버가 요청 사용자로 강제 지정 — 타인 명의 팀 생성 차단
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    item.User = int(user.Id)

	conn := c.NewConnection()

	manager := models.NewTeamManager(conn)
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

func (c *TeamController) Insertbatch(item *[]models.Team) {  
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)

    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }

	conn := c.NewConnection()

	manager := models.NewTeamManager(conn)

    for i := 0; i < rows; i++ {
        (*item)[i].User = int(user.Id)   // 소유자 서버 강제 지정

	    err := manager.Insert(&((*item)[i]))
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}

func (c *TeamController) Update(item *models.Team) {

	conn := c.NewConnection()

	manager := models.NewTeamManager(conn)

    // 내 소유 팀만 수정 가능. 소유자 필드는 기존 값으로 고정해 이전(양도) 위조 차단
    user := requestUser(&c.Controller)
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errNotFound)
        return
    }
    if user == nil || int64(existing.User) != user.Id {
        c.Error(errForbidden)
        return
    }
    item.User = existing.User

    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }
}

func (c *TeamController) Delete(item *models.Team) {


    conn := c.NewConnection()

	manager := models.NewTeamManager(conn)

    user := requestUser(&c.Controller)
    existing := manager.Get(item.Id)
    if existing == nil {
        return   // 이미 없음 — 멱등 처리
    }
    if user == nil || int64(existing.User) != user.Id {
        c.Error(errForbidden)
        return
    }

	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
    }
}

func (c *TeamController) Deletebatch(item *[]models.Team) {
    
    
    conn := c.NewConnection()

	manager := models.NewTeamManager(conn)

    // 전량 사전 검증 후 일괄 삭제 — 중간 실패로 일부만 지워지는 것을 방지
    user := requestUser(&c.Controller)
    for _, v := range *item {
        existing := manager.Get(v.Id)
        if existing == nil {
            continue
        }
        if user == nil || int64(existing.User) != user.Id {
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


