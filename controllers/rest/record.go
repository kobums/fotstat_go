package rest


import (
	"fotstat/controllers"

	"fotstat/models"
	"fotstat/models/record"

	"strings"
)

type RecordController struct {
	controllers.Controller
}

func (c *RecordController) Read(id int64) {
    
    
	conn := c.NewConnection()

	manager := models.NewRecordManager(conn)
	item := manager.Get(id)

    
    
    c.Set("item", item)
}

func (c *RecordController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewRecordManager(conn)

    var args []interface{}
    
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
    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }
}

func (c *RecordController) UpdateStats(item *models.Record) {
	conn := c.NewConnection()
	manager := models.NewRecordManager(conn)
	err := manager.UpdateWhere(
		[]record.Params{
			{Column: record.ColumnMin, Value: item.Min},
			{Column: record.ColumnGoal, Value: item.Goal},
			{Column: record.ColumnAssist, Value: item.Assist},
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

    
	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
    }
}

func (c *RecordController) Deletebatch(item *[]models.Record) {
    
    
    conn := c.NewConnection()

	manager := models.NewRecordManager(conn)

    for _, v := range *item {
        
    
	    err := manager.Delete(v.Id)
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}


