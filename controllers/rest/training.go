package rest


import (
	"fotstat/controllers"

	"fotstat/models"

    "strings"
)

type TrainingController struct {
	controllers.Controller
}

// 소유권 검증은 ownership.go 의 공용 헬퍼(requestUser/ownsTeam/ownsTraining/ownTrainingScope)를 사용한다.
// 훈련은 team 직속 리소스라 match 와 동일한 소유 체인(team ← training)을 가진다.

func (c *TrainingController) Read(id int64) {


	conn := c.NewConnection()

	manager := models.NewTrainingManager(conn)
	item := manager.Get(id)

    if item != nil && !ownsTeam(conn, requestUser(&c.Controller), item.Team) {
        c.Error(errForbidden)
        return
    }

    c.Set("item", item)
}

func (c *TrainingController) Index(page int, pagesize int) {


	conn := c.NewConnection()

	manager := models.NewTrainingManager(conn)

    var args []interface{}

    // 소유권 강제: 클라이언트 필터와 무관하게 요청 사용자 소유 팀으로 범위를 제한한다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownTrainingScope(user))

    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Where{Column:"team", Value:_team, Compare:"="})
    }
    _starttrainingdate := c.Get("starttrainingdate")
    _endtrainingdate := c.Get("endtrainingdate")
    if _starttrainingdate != "" && _endtrainingdate != "" {
        var v [2]string
        v[0] = _starttrainingdate
        v[1] = _endtrainingdate
        args = append(args, models.Where{Column:"trainingdate", Value:v, Compare:"between"})
    } else if  _starttrainingdate != "" {
        args = append(args, models.Where{Column:"trainingdate", Value:_starttrainingdate, Compare:">="})
    } else if  _endtrainingdate != "" {
        args = append(args, models.Where{Column:"trainingdate", Value:_endtrainingdate, Compare:"<="})
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
                    str += ", tr_" + strings.Trim(v, " ")
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

func (c *TrainingController) Count() {


	conn := c.NewConnection()

	manager := models.NewTrainingManager(conn)

    var args []interface{}

    // 소유권 강제: Index 와 동일하게 요청 사용자 소유 팀으로 범위 제한
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownTrainingScope(user))

    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Where{Column:"team", Value:_team, Compare:"="})
    }



    total := manager.Count(args)
	c.Set("total", total)
}

func (c *TrainingController) Insert(item *models.Training) {


	conn := c.NewConnection()

    if !ownsTeam(conn, requestUser(&c.Controller), item.Team) {
        c.Error(errForbidden)
        return
    }

	manager := models.NewTrainingManager(conn)
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

func (c *TrainingController) Insertbatch(item *[]models.Training) {
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)


	conn := c.NewConnection()

	manager := models.NewTrainingManager(conn)

    // Insert 와 동일한 검증을 배치에도 적용 — 전량 사전 검증 후 일괄 삽입해 부분 실패를 줄인다
    for i := 0; i < rows; i++ {
        if !ownsTeam(conn, requestUser(&c.Controller), (*item)[i].Team) {
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

func (c *TrainingController) Update(item *models.Training) {


	conn := c.NewConnection()

	manager := models.NewTrainingManager(conn)

    // 기존 레코드의 팀과 변경 후 팀 모두 내 소유여야 한다
    // (타인 레코드 수정과 내 레코드를 타인 팀으로 옮기는 것 둘 다 차단)
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errNotFound)
        return
    }
    if !ownsTeam(conn, requestUser(&c.Controller), existing.Team) || !ownsTeam(conn, requestUser(&c.Controller), item.Team) {
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

func (c *TrainingController) Delete(item *models.Training) {


    conn := c.NewConnection()

	manager := models.NewTrainingManager(conn)

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

func (c *TrainingController) Deletebatch(item *[]models.Training) {


    conn := c.NewConnection()

	manager := models.NewTrainingManager(conn)

    // 전량 사전 검증 후 일괄 삭제 — 중간 실패로 일부만 지워지는 것을 방지
    for _, v := range *item {
        existing := manager.Get(v.Id)
        if existing == nil {
            continue  // 이미 없는 항목은 멱등 처리
        }
        if !ownsTeam(conn, requestUser(&c.Controller), existing.Team) {
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
