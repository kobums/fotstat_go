package rest


import (
	"fotstat/controllers"

	"fotstat/models"

    "errors"
    "fmt"
    "strings"
)

type InjuryController struct {
	controllers.Controller
}

// validDates 는 복귀일이 있으면 발생일 이후인지 확인한다. returndate 가 비어
// 있으면(아직 부상 중) 통과. 프런트에서도 막지만 API 직접 호출 방어 목적.
func validInjuryDates(item *models.Injury) error {
	if item.Returndate != "" && item.Startdate != "" && item.Returndate < item.Startdate {
		return errors.New("returndate must be on or after startdate")
	}
	return nil
}

func (c *InjuryController) Read(id int64) {


	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)
	item := manager.Get(id)



    c.Set("item", item)
}

func (c *InjuryController) Index(page int, pagesize int) {


	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)

    var args []interface{}

    _player := c.Geti("player")
    if _player != 0 {
        args = append(args, models.Where{Column:"player", Value:_player, Compare:"="})
    }
    // team 필터: 해당 팀 소속 선수들의 부상만 (player_tb 서브쿼리)
    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Custom{Query: fmt.Sprintf("i_player in (select p_id from player_tb where p_team = %d)", _team)})
    }
    // active=1 이면 아직 복귀 전(부상 중)인 기록만
    _active := c.Geti("active")
    if _active != 0 {
        args = append(args, models.Custom{Query: "i_returndate is null"})
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
                    str += ", i_" + strings.Trim(v, " ")
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

func (c *InjuryController) Count() {


	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)

    var args []interface{}

    _player := c.Geti("player")
    if _player != 0 {
        args = append(args, models.Where{Column:"player", Value:_player, Compare:"="})
    }
    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Custom{Query: fmt.Sprintf("i_player in (select p_id from player_tb where p_team = %d)", _team)})
    }
    _active := c.Geti("active")
    if _active != 0 {
        args = append(args, models.Custom{Query: "i_returndate is null"})
    }



    total := manager.Count(args)
	c.Set("total", total)
}

func (c *InjuryController) Insert(item *models.Injury) {

    if err := validInjuryDates(item); err != nil {
        c.Error(err)
        return
    }

	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)
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

func (c *InjuryController) Insertbatch(item *[]models.Injury) {
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)


	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)

    for i := 0; i < rows; i++ {

	    err := manager.Insert(&((*item)[i]))
        if err != nil {
            c.Set("code", "error")
            c.Set("error", err)
            return
        }
    }
}

func (c *InjuryController) Update(item *models.Injury) {

    if err := validInjuryDates(item); err != nil {
        c.Error(err)
        return
    }

	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)
    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")
        c.Set("error", err)
        return
    }
}

func (c *InjuryController) Delete(item *models.Injury) {


    conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)


	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")
        c.Set("error", err)
    }
}

func (c *InjuryController) Deletebatch(item *[]models.Injury) {


    conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)

    for _, v := range *item {


	    err := manager.Delete(v.Id)
        if err != nil {
            c.Set("code", "error")
            c.Set("error", err)
            return
        }
    }
}
