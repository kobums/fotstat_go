package rest


import (
	"fotstat/controllers"
	"fotstat/models"

    "strings"
)

type UserController struct {
	controllers.Controller
}

func (c *UserController) Read(id int64) {
    
    
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
	item := manager.Get(id)

    
    
    c.Set("item", item)
}

func (c *UserController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    var args []interface{}
    
    _email := c.Get("email")
    if _email != "" {
        args = append(args, models.Where{Column:"email", Value:_email, Compare:"like"})
    }
    _password := c.Get("password")
    if _password != "" {
        args = append(args, models.Where{Column:"password", Value:_password, Compare:"like"})
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
    
    _email := c.Get("email")
    if _email != "" {
        args = append(args, models.Where{Column:"email", Value:_email, Compare:"like"})
        
    }
    _password := c.Get("password")
    if _password != "" {
        args = append(args, models.Where{Column:"password", Value:_password, Compare:"like"})
        
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
	    err := manager.Insert(&((*item)[i]))
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}

func (c *UserController) Update(item *models.UserUpdate) {
    
    
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }
}

func (c *UserController) Delete(item *models.User) {
    
    
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

    for _, v := range *item {
        
    
	    err := manager.Delete(v.Id)
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}


