package rest


import (
	"fotstat/controllers"

	"fotstat/models"

	"errors"
	"fmt"
	"strings"
)

type AttendanceController struct {
	controllers.Controller
}

// trainingInjuryConflict reports whether the given player is injured on the
// training date. record 의 injuryConflict 와 동일 정책 — 부상 기간에 걸치면 참석
// 입력을 차단하되, 발생일 당일 훈련은 허용한다(훈련 중 부상 = 그날까지는 뛴 것).
// 즉 차단 범위는 발생일 다음 날부터 복귀일까지(i_returndate NULL 이면 계속 차단).
func (c *AttendanceController) trainingInjuryConflict(conn *models.Connection, training int, player int) error {
	if training == 0 || player == 0 {
		return nil
	}

	t := models.NewTrainingManager(conn).Get(int64(training))
	if t == nil || t.Trainingdate == "" {
		return nil
	}

	// tr_trainingdate 는 DATETIME, injury 날짜는 DATE 이므로 날짜 부분만 비교한다.
	trainingdate := t.Trainingdate
	if len(trainingdate) >= 10 {
		trainingdate = trainingdate[:10]
	}

	injuryManager := models.NewInjuryManager(conn)
	cnt := injuryManager.Count([]interface{}{
		models.Where{Column: "player", Value: player, Compare: "="},
		models.Custom{Query: fmt.Sprintf("i_startdate < '%s'", trainingdate)},
		models.Custom{Query: fmt.Sprintf("(i_returndate is null or i_returndate >= '%s')", trainingdate)},
	})

	if cnt > 0 {
		return errors.New("injured player cannot attend this training")
	}

	return nil
}

func (c *AttendanceController) Read(id int64) {


	conn := c.NewConnection()

	manager := models.NewAttendanceManager(conn)
	item := manager.Get(id)

    if item != nil && !ownsTraining(conn, requestUser(&c.Controller), item.Training) {
        c.Error(errForbidden)
        return
    }

    c.Set("item", item)
}

func (c *AttendanceController) Index(page int, pagesize int) {


	conn := c.NewConnection()

	manager := models.NewAttendanceManager(conn)

    var args []interface{}

    // 소유권 강제: 요청 사용자 소유 팀의 훈련 참석만 조회된다
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownAttendanceScope(user))

    _training := c.Geti("training")
    if _training != 0 {
        args = append(args, models.Where{Column:"training", Value:_training, Compare:"="})
    }
    _player := c.Geti("player")
    if _player != 0 {
        args = append(args, models.Where{Column:"player", Value:_player, Compare:"="})
    }
    // team 필터: 해당 팀 훈련의 참석만 (training_tb 서브쿼리) — 팀 단위 집계용
    _team := c.Geti("team")
    if _team != 0 {
        args = append(args, models.Custom{Query: fmt.Sprintf("a_training in (select tr_id from training_tb where tr_team = %d)", _team)})
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
                    str += ", a_" + strings.Trim(v, " ")
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

func (c *AttendanceController) Count() {


	conn := c.NewConnection()

	manager := models.NewAttendanceManager(conn)

    var args []interface{}

    // 소유권 강제: Index 와 동일하게 요청 사용자 소유 범위로 제한
    user := requestUser(&c.Controller)
    if user == nil {
        c.Error(errForbidden)
        return
    }
    args = append(args, ownAttendanceScope(user))

    _training := c.Geti("training")
    if _training != 0 {
        args = append(args, models.Where{Column:"training", Value:_training, Compare:"="})
    }
    _player := c.Geti("player")
    if _player != 0 {
        args = append(args, models.Where{Column:"player", Value:_player, Compare:"="})
    }



    total := manager.Count(args)
	c.Set("total", total)
}

func (c *AttendanceController) Insert(item *models.Attendance) {


	conn := c.NewConnection()

    // 내 소유 팀의 훈련에, 그 팀 소속 선수의 참석만 생성 가능
    user := requestUser(&c.Controller)
    if !ownsAttendanceTarget(conn, user, item.Training, item.Player) {
        c.Error(errForbidden)
        return
    }

    if err := c.trainingInjuryConflict(conn, item.Training, item.Player); err != nil {
        c.Set("code", "error")
        c.Set("error", err.Error())
        return
    }

	manager := models.NewAttendanceManager(conn)
	// (training, player) 는 UNIQUE — 이미 행이 있으면 삽입 대신 갱신되어
	// 참석 체크 중복 호출(레이스·재전송)에도 행이 늘지 않는다
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

func (c *AttendanceController) Insertbatch(item *[]models.Attendance) {
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)



	conn := c.NewConnection()

	manager := models.NewAttendanceManager(conn)

    // 전량 사전 검증(소유권·팀 일치 + 부상 충돌) 후 일괄 삽입
    user := requestUser(&c.Controller)
    for i := 0; i < rows; i++ {
        if !ownsAttendanceTarget(conn, user, (*item)[i].Training, (*item)[i].Player) {
            c.Error(errForbidden)
            return
        }
        if err := c.trainingInjuryConflict(conn, (*item)[i].Training, (*item)[i].Player); err != nil {
            c.Set("code", "error")
            c.Set("error", err.Error())
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

func (c *AttendanceController) Update(item *models.Attendance) {


	conn := c.NewConnection()

	manager := models.NewAttendanceManager(conn)

    // 기존 참석과 변경 후 값(훈련·선수) 모두 내 소유여야 한다
    user := requestUser(&c.Controller)
    existing := manager.Get(item.Id)
    if existing == nil {
        c.Error(errNotFound)
        return
    }
    if !ownsTraining(conn, user, existing.Training) ||
        !ownsAttendanceTarget(conn, user, item.Training, item.Player) {
        c.Error(errForbidden)
        return
    }

    if err := c.trainingInjuryConflict(conn, item.Training, item.Player); err != nil {
        c.Set("code", "error")
        c.Set("error", err.Error())
        return
    }

    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")
        // uq_attendance_training_player 충돌 — 원본 DB 에러(제약명 등)를 클라이언트에 노출하지 않는다
        if strings.Contains(err.Error(), "Duplicate entry") {
            c.Set("error", "attendance already exists for this training and player")
        } else {
            c.Set("error", err)
        }
        return
    }
}

func (c *AttendanceController) Delete(item *models.Attendance) {


    conn := c.NewConnection()

	manager := models.NewAttendanceManager(conn)

    existing := manager.Get(item.Id)
    if existing == nil {
        return   // 이미 없음 — 멱등 처리
    }
    if !ownsTraining(conn, requestUser(&c.Controller), existing.Training) {
        c.Error(errForbidden)
        return
    }

	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")
        c.Set("error", err)
    }
}

func (c *AttendanceController) Deletebatch(item *[]models.Attendance) {


    conn := c.NewConnection()

	manager := models.NewAttendanceManager(conn)

    // 전량 사전 검증 후 일괄 삭제
    user := requestUser(&c.Controller)
    for _, v := range *item {
        existing := manager.Get(v.Id)
        if existing == nil {
            continue
        }
        if !ownsTraining(conn, user, existing.Training) {
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
