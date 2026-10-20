package rest


import (
	"errors"
	"fotstat/controllers"
	"fotstat/global/jwt"
	"fotstat/models"

    "strings"
)

type UserController struct {
	controllers.Controller
}

// checkSignupEmail 은 가입 시 이메일이 비었거나, guest: 예약 접두사이거나,
// 이미 사용 중이면 에러를 돌려준다.
func checkSignupEmail(c *UserController, email string) error {
    email = strings.TrimSpace(email)
    if email == "" || strings.HasPrefix(email, "guest:") {
        return errors.New("invalid email")
    }
    conn := c.NewConnection()
    var args []interface{}
    args = append(args, models.Where{Column: "email", Value: email, Compare: "="})
    if models.NewUserManager(conn).GetWhere(args) != nil {
        return errors.New("email already in use")
    }
    return nil
}

func (c *UserController) Read(id int64) {

    // 본인 계정만 조회 가능 — 타인 이메일/비밀번호 해시 노출 차단
    user := requestUser(&c.Controller)
    if user == nil || user.Id != id {
        c.Error(errForbidden)
        return
    }

	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
	item := manager.Get(id)

    if item != nil {
        item.Password = ""   // 해시라도 응답에 노출하지 않는다
    }

    c.Set("item", item)
}

func (c *UserController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    var args []interface{}

    // 본인 계정으로 강제 스코프. password LIKE 검색 파라미터는 오라클로 악용될 수 있어 제거
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownUserScope(user))

    _email := c.Get("email")
    if _email != "" {
        args = append(args, models.Where{Column:"email", Value:_email, Compare:"like"})
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
                    str += ", u_" + strings.Trim(v, " ")                
                }
            }
        }
        
        args = append(args, models.Ordering(str))
    }
    
	items := manager.Find(args)
    for i := range items {
        items[i].Password = ""   // 해시라도 응답에 노출하지 않는다
    }
	c.Set("items", items)

    if page == 1 {
       total := manager.Count(args)
	   c.Set("total", total)
    }
}

func (c *UserController) Count() {


	conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    var args []interface{}

    // 본인 계정으로 강제 스코프 (Index 와 동일)
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownUserScope(user))

    _email := c.Get("email")
    if _email != "" {
        args = append(args, models.Where{Column:"email", Value:_email, Compare:"like"})

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

func (c *UserController) Insert(item *models.UserUpdate) {

    // 비로그인으로도 호출되는 가입 경로 — id 지정 삽입과 이메일 중복 가입을 막는다
    item.Id = 0
    item.Email = strings.TrimSpace(item.Email)
    if err := checkSignupEmail(c, item.Email); err != nil {
        c.Error(err)
        return
    }
    if item.Password == "" {
        c.Error(errors.New("password is required"))
        return
    }

    if item.Password != "" {
        hashed, err := jwt.GeneratePasswd(item.Password)
        if err == nil {
            item.Password = hashed
        }
    }
    

	conn := c.NewConnection()
    
	manager := models.NewUserManager(conn)
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

func (c *UserController) Insertbatch(item *[]models.UserUpdate) {  
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)
    
    
    
	conn := c.NewConnection()
    
	manager := models.NewUserManager(conn)

    for i := 0; i < rows; i++ {

        (*item)[i].Id = 0
        (*item)[i].Email = strings.TrimSpace((*item)[i].Email)
        if err := checkSignupEmail(c, (*item)[i].Email); err != nil {
            c.Error(err)
            return
        }
        if (*item)[i].Password == "" {
            c.Error(errors.New("password is required"))
            return
        }

        if (*item)[i].Password != "" {
            hashed, err := jwt.GeneratePasswd((*item)[i].Password)
            if err == nil {
                (*item)[i].Password = hashed
            }
        }
        
	    err := manager.Insert(&((*item)[i]))
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}

func (c *UserController) Update(item *models.UserUpdate) {

    // 본인 계정만 수정 가능
    user := requestUser(&c.Controller)
    if user == nil || user.Id != item.Id {
        c.Error(errForbidden)
        return
    }

	conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    // manager.Update 는 모든 컬럼을 덮어쓰므로, 비어 온 값은 현재 값으로 채워
    // 비밀번호가 빈 문자열로 지워지는 것을 막고, 이메일 변경 시에는 가입과 같은
    // 검증(빈 값·guest: 접두사·타 계정과 중복)을 적용한다.
    current := manager.Get(item.Id)
    if current == nil {
        c.Error(errForbidden)
        return
    }
    item.Email = strings.TrimSpace(item.Email)
    if item.Email == "" {
        item.Email = current.Email
    } else if item.Email != current.Email {
        if err := checkSignupEmail(c, item.Email); err != nil {
            c.Error(err)
            return
        }
    }
    if item.Createddate == "" {
        item.Createddate = current.Createddate
    }

    if item.Password != "" {
        hashed, err := jwt.GeneratePasswd(item.Password)
        if err != nil {
            c.Error(errors.New("password update failed"))
            return
        }
        item.Password = hashed
    } else {
        item.Password = current.Password
    }

    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }
}

func (c *UserController) Delete(item *models.User) {

    // 본인 계정만 삭제 가능 (정식 탈퇴 흐름은 /api/account 라우트 사용)
    user := requestUser(&c.Controller)
    if user == nil || user.Id != item.Id {
        c.Error(errForbidden)
        return
    }

    conn := c.NewConnection()

	manager := models.NewUserManager(conn)


	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
    }
}

func (c *UserController) Deletebatch(item *[]models.User) {
    
    
    conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    // 본인 계정만 삭제 가능
    user := requestUser(&c.Controller)
    for _, v := range *item {
        if user == nil || user.Id != v.Id {
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


