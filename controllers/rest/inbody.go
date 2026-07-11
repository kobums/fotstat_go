package rest


import (
	"fotstat/controllers"

	"fotstat/models"

	"errors"
	"fmt"
	"strings"
)

type InbodyController struct {
	controllers.Controller
}

// validInbody 는 검사일 필수·측정치 음수 금지를 확인한다. 검사일 외 측정치는
// 전부 선택 입력(0 = 미측정 = DB NULL). 프런트에서도 막지만 API 직접 호출 방어 목적.
func validInbody(item *models.Inbody) error {
	if item.Testdate == "" {
		return errors.New("testdate is required")
	}
	if item.Height < 0 || item.Weight < 0 || item.Muscle < 0 || item.Fat < 0 ||
		item.Rightleg < 0 || item.Leftleg < 0 || item.Score < 0 {
		return errors.New("measurements must not be negative")
	}
	return nil
}

// 소유권 검증은 ownership.go 의 공용 헬퍼(requestUser/ownsPlayer/ownInbodyScope)를 사용한다.
// 인바디는 선수의 민감 정보라 모든 CRUD 에서 소유권을 검증한다 (IDOR 방지).

func (c *InbodyController) Read(id int64) {


	conn := c.NewConnection()

	manager := models.NewInbodyManager(conn)
	item := manager.Get(id)

    if item != nil && !ownsPlayer(conn, requestUser(&c.Controller), item.Player) {
        c.Error(errForbidden)
        return
    }

    c.Set("item", item)
}

func (c *InbodyController) Index(page int, pagesize int) {


	conn := c.NewConnection()

	manager := models.NewInbodyManager(conn)

    var args []interface{}

    // 소유권 강제: 클라이언트 필터와 무관하게 요청 사용자 소유 팀의 선수로 범위를 제한한다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownInbodyScope(user))

    _player := c.Geti("player")
    if _player != 0 {
        args = append(args, models.Where{Column:"player", Value:_player, Compare:"="})
    }
    // team 필터: 해당 팀 소속 선수들의 측정만 (player_tb 서브쿼리)
    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Custom{Query: fmt.Sprintf("ib_player in (select p_id from player_tb where p_team = %d)", _team)})
    }
    _starttestdate := c.Get("starttestdate")
    _endtestdate := c.Get("endtestdate")
    if _starttestdate != "" && _endtestdate != "" {
        var v [2]string
        v[0] = _starttestdate
        v[1] = _endtestdate
        args = append(args, models.Where{Column:"testdate", Value:v, Compare:"between"})
    } else if  _starttestdate != "" {
        args = append(args, models.Where{Column:"testdate", Value:_starttestdate, Compare:">="})
    } else if  _endtestdate != "" {
        args = append(args, models.Where{Column:"testdate", Value:_endtestdate, Compare:"<="})
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
                    str += ", ib_" + strings.Trim(v, " ")
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

func (c *InbodyController) Count() {


	conn := c.NewConnection()

	manager := models.NewInbodyManager(conn)

    var args []interface{}

    // 소유권 강제: Index 와 동일하게 요청 사용자 소유 팀으로 범위 제한
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownInbodyScope(user))

    _player := c.Geti("player")
    if _player != 0 {
        args = append(args, models.Where{Column:"player", Value:_player, Compare:"="})
    }
    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Custom{Query: fmt.Sprintf("ib_player in (select p_id from player_tb where p_team = %d)", _team)})
    }



    total := manager.Count(args)
	c.Set("total", total)
}

func (c *InbodyController) Insert(item *models.Inbody) {

    if err := validInbody(item); err != nil {
        c.Error(err)
        return
    }

	conn := c.NewConnection()

    if !ownsPlayer(conn, requestUser(&c.Controller), item.Player) {
        c.Error(errForbidden)
        return
    }

	manager := models.NewInbodyManager(conn)
	// (player, testdate) 는 UNIQUE — 같은 검사일에 다시 저장하면 삽입 대신
	// 갱신되어 단건 등록·시트 재저장 모두 마지막 값으로 수렴한다
	err := manager.Upsert(item)
    if err != nil {
        c.Set("code", "error")
        c.Set("error", err)
        return
    }

    id := manager.GetIdentity()
    c.Result["id"] = id
    item.Id = id
}

func (c *InbodyController) Insertbatch(item *[]models.Inbody) {
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)


	conn := c.NewConnection()

	manager := models.NewInbodyManager(conn)

    // 전량 사전 검증(입력값 + 소유권) 후 일괄 upsert — 부분 실패를 줄인다
    for i := 0; i < rows; i++ {
        if err := validInbody(&((*item)[i])); err != nil {
            c.Error(err)
            return
        }
        if !ownsPlayer(conn, requestUser(&c.Controller), (*item)[i].Player) {
            c.Error(errForbidden)
            return
        }
    }

    for i := 0; i < rows; i++ {

	    err := manager.Upsert(&((*item)[i]))
        if err != nil {
            c.Set("code", "error")
            c.Set("error", err)
            return
        }
    }
}

func (c *InbodyController) Update(item *models.Inbody) {

    if err := validInbody(item); err != nil {
        c.Error(err)
        return
    }

	conn := c.NewConnection()

	manager := models.NewInbodyManager(conn)

    // 기존 레코드의 선수와 변경 후 선수 모두 내 소유여야 한다
    // (타인 레코드 수정과 내 레코드를 타인 선수로 옮기는 것 둘 다 차단)
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errNotFound)
        return
    }
    if !ownsPlayer(conn, requestUser(&c.Controller), existing.Player) || !ownsPlayer(conn, requestUser(&c.Controller), item.Player) {
        c.Error(errForbidden)
        return
    }

    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")
        // uk_inbody_player_date 충돌 — 원본 DB 에러(제약명 등)를 클라이언트에 노출하지 않는다
        if strings.Contains(err.Error(), "Duplicate entry") {
            c.Set("error", "measurement already exists for this player and date")
        } else {
            c.Set("error", err)
        }
        return
    }
}

func (c *InbodyController) Delete(item *models.Inbody) {


    conn := c.NewConnection()

	manager := models.NewInbodyManager(conn)

    existing := manager.Get(item.Id)
    if existing == nil {
        return   // 이미 없음 — 멱등 처리
    }
    if !ownsPlayer(conn, requestUser(&c.Controller), existing.Player) {
        c.Error(errForbidden)
        return
    }

	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")
        c.Set("error", err)
    }
}

func (c *InbodyController) Deletebatch(item *[]models.Inbody) {


    conn := c.NewConnection()

	manager := models.NewInbodyManager(conn)

    // 전량 사전 검증 후 일괄 삭제 — 중간 실패로 일부만 지워지는 것을 방지
    for _, v := range *item {
        existing := manager.Get(v.Id)
        if existing == nil {
            continue  // 이미 없는 항목은 멱등 처리
        }
        if !ownsPlayer(conn, requestUser(&c.Controller), existing.Player) {
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
