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

// validInjuryDates 는 복귀일이 있으면 발생일 이후인지 확인한다. returndate 가 비어
// 있으면(아직 부상 중) 통과. 프런트에서도 막지만 API 직접 호출 방어 목적.
func validInjuryDates(item *models.Injury) error {
	if item.Returndate != "" && item.Startdate != "" && item.Returndate < item.Startdate {
		return errors.New("returndate must be on or after startdate")
	}
	return nil
}

var errInjuryForbidden = errors.New("forbidden: player does not belong to your team")

// currentUser 는 JwtAuthRequired 미들웨어가 세팅한 요청 사용자. 없으면 nil.
func (c *InjuryController) currentUser() *models.User {
	if c.Context == nil {
		return nil
	}
	user, ok := c.Context.Locals("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}

// ownsPlayer 는 playerId 소속 팀이 요청 사용자의 소유인지 확인한다.
// 부상은 선수의 민감 정보라 모든 CRUD 에서 소유권을 검증한다 (IDOR 방지).
func (c *InjuryController) ownsPlayer(conn *models.Connection, playerId int) bool {
	user := c.currentUser()
	if user == nil || playerId == 0 {
		return false
	}
	player := models.NewPlayerManager(conn).Get(int64(playerId))
	if player == nil {
		return false
	}
	team := models.NewTeamManager(conn).Get(int64(player.Team))
	return team != nil && int64(team.User) == user.Id
}

func (c *InjuryController) Read(id int64) {


	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)
	item := manager.Get(id)

    if item != nil && !c.ownsPlayer(conn, item.Player) {
        c.Error(errInjuryForbidden)
        return
    }

    c.Set("item", item)
}

func (c *InjuryController) Index(page int, pagesize int) {


	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)

    var args []interface{}

    // 소유권 강제: 클라이언트 필터와 무관하게 요청 사용자 소유 팀의 선수로 범위를 제한한다
    user := c.currentUser()
    if user == nil {
        c.Error(errInjuryForbidden)
        return
    }
    args = append(args, models.Custom{Query: fmt.Sprintf("i_player in (select p_id from player_tb join team_tb on p_team = t_id where t_user = %d)", user.Id)})

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

    // 소유권 강제: Index 와 동일하게 요청 사용자 소유 팀으로 범위 제한
    user := c.currentUser()
    if user == nil {
        c.Error(errInjuryForbidden)
        return
    }
    args = append(args, models.Custom{Query: fmt.Sprintf("i_player in (select p_id from player_tb join team_tb on p_team = t_id where t_user = %d)", user.Id)})

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

    if !c.ownsPlayer(conn, item.Player) {
        c.Error(errInjuryForbidden)
        return
    }

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

    // Insert 와 동일한 검증을 배치에도 적용 — 전량 사전 검증 후 일괄 삽입해 부분 실패를 줄인다
    for i := 0; i < rows; i++ {
        if err := validInjuryDates(&((*item)[i])); err != nil {
            c.Error(err)
            return
        }
        if !c.ownsPlayer(conn, (*item)[i].Player) {
            c.Error(errInjuryForbidden)
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

func (c *InjuryController) Update(item *models.Injury) {

    if err := validInjuryDates(item); err != nil {
        c.Error(err)
        return
    }

	conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)

    // 기존 레코드의 선수와 변경 후 선수 모두 내 소유여야 한다
    // (타인 레코드 수정과 내 레코드를 타인 선수로 옮기는 것 둘 다 차단)
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errors.New("injury not found"))
        return
    }
    if !c.ownsPlayer(conn, existing.Player) || !c.ownsPlayer(conn, item.Player) {
        c.Error(errInjuryForbidden)
        return
    }

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

    existing := manager.Get(item.Id)
    if existing == nil {
        return
    }
    if !c.ownsPlayer(conn, existing.Player) {
        c.Error(errInjuryForbidden)
        return
    }

	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")
        c.Set("error", err)
    }
}

func (c *InjuryController) Deletebatch(item *[]models.Injury) {


    conn := c.NewConnection()

	manager := models.NewInjuryManager(conn)

    // Insertbatch 와 동일하게 전량 사전 검증 후 일괄 삭제 — 중간 실패로 일부만 지워지는 것을 방지
    for _, v := range *item {
        existing := manager.Get(v.Id)
        if existing == nil {
            continue  // 이미 없는 항목은 멱등 처리
        }
        if !c.ownsPlayer(conn, existing.Player) {
            c.Error(errInjuryForbidden)
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
