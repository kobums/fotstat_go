package rest

// 선수 상세 통계 API — GET /api/player/:id/stats?start=YYYY-MM-DD&end=YYYY-MM-DD
//
// 선수 상세 화면(웹 PlayerDetailPage · iOS PlayerDetailView)이 한 번의 호출로 그리는 데
// 필요한 것을 모두 내려준다. 그동안 클라이언트가 경기→쿼터→기록으로 fan-out 하던
// N+1 호출을 서버 조인으로 대체한다.
//
// 집계 규칙은 report.go aggregate() 및 웹 aggregateTeamStats.ts / playerMatchLog.ts /
// lib/training.ts 와 동일하다:
//   - 경기수 = 선수가 기록을 남긴 경기 수, 출전 시간·골·도움·카드 = 기록 합
//   - 결장 = 부상 기간에 걸친 팀 경기 수(오늘까지, 발생일 당일 제외)
//   - 스코어: 우리 팀 = 경기 내 전 선수 골 합, 상대 = 쿼터 awaygoals 합
//   - 훈련 참석률 = 기간 내 이미 열린(오늘까지) 훈련 중 참석 수. 미래 훈련은 분모 제외
//   - 부상 이력은 기간과 무관하게 전체(선수 내력) — 기간 반영은 결장 수가 담당

import (
	"regexp"
	"sort"

	"fotstat/controllers"
	"fotstat/models"
)

type PlayerStatsController struct {
	controllers.Controller
}

// PlayerLine 은 한 선수의 기간 집계. 웹 PlayerStat / iOS PlayerStats 와 필드가 같다.
type PlayerLine struct {
	Id          int64         `json:"id"`
	Name        string        `json:"name"`
	Number      int           `json:"number"`
	Position    string        `json:"position"`
	Games       int           `json:"games"`
	Min         int           `json:"min"`
	Goal        int           `json:"goal"`
	Assist      int           `json:"assist"`
	Yellow      int           `json:"yellow"`
	Red         int           `json:"red"`
	AbsentGames int           `json:"absentGames"`
	Training    *TrainingLine `json:"training"` // 기간 내 열린 훈련이 없으면 null
}

// TrainingLine 은 선수의 훈련 참석 집계.
type TrainingLine struct {
	Attended int `json:"attended"`
	Held     int `json:"held"`
	Rate     int `json:"rate"` // %, 반올림
	TotalMin int `json:"totalMin"`
}

// QuarterLine / MatchLine 은 경기별 기록. 웹 PlayerMatchLog / iOS PlayerMatchLog 와 같다.
type QuarterLine struct {
	QuarterId int64 `json:"quarterId"`
	Number    int   `json:"number"`
	Min       int   `json:"min"`
	Goal      int   `json:"goal"`
	Assist    int   `json:"assist"`
	Yellow    int   `json:"yellow"`
	Red       int   `json:"red"`
}

type MatchLine struct {
	MatchId   int64         `json:"matchId"`
	Opponent  string        `json:"opponent"`
	Matchdate string        `json:"matchdate"`
	Home      int           `json:"home"`
	Away      int           `json:"away"`
	Quarters  []QuarterLine `json:"quarters"`
	Min       int           `json:"min"`
	Goal      int           `json:"goal"`
	Assist    int           `json:"assist"`
	Yellow    int           `json:"yellow"`
	Red       int           `json:"red"`
}

// PlayerStatsResult 는 응답 item.
type PlayerStatsResult struct {
	Player     models.Player   `json:"player"`
	Start      string          `json:"start"`
	End        string          `json:"end"`
	MatchCount int             `json:"matchCount"` // 기간 내 진행된(쿼터가 있는) 팀 경기 수
	Summary    PlayerLine      `json:"summary"`
	Squad      []PlayerLine    `json:"squad"` // 스쿼드 전원(순위·팀 평균 계산용), 등번호→이름 순
	Matches    []MatchLine     `json:"matches"`
	Injuries   []models.Injury `json:"injuries"`
}

// playerStatsInput 은 집계에 필요한 원본 묶음 — 컨트롤러가 채우고 순수 함수가 소비한다.
type playerStatsInput struct {
	players     []models.Player
	matches     []models.Match // 기간 내 팀 경기
	quarters    []models.Quarter
	records     []models.Record
	injuries    []models.Injury   // 팀 전 선수의 전체 이력
	trainings   []models.Training // 기간 내 팀 훈련
	attendances []models.Attendance
}

var dayRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// Stats 는 GET /player/:id/stats 핸들러 본체.
func (c *PlayerStatsController) Stats(playerId int64) {
	conn := c.NewConnection()

	// 실패는 다른 JSON 라우트처럼 HTTP 200 + code:"error" 로 내려간다(라우터 참고).
	// 없는 선수와 남의 팀 선수를 구분하지 않고 같은 메시지를 준다 — id 존재 여부 노출 방지.
	user := requestUser(&c.Controller)
	if user == nil {
		c.Error(errForbidden)
		return
	}
	player := models.NewPlayerManager(conn).Get(playerId)
	if player == nil || !ownsTeam(conn, user, player.Team) {
		c.Error(errForbidden)
		return
	}

	start := c.Get("start")
	end := c.Get("end")
	if (start != "" && !dayRe.MatchString(start)) || (end != "" && !dayRe.MatchString(end)) {
		c.Set("code", "error")
		c.Set("message", "start/end must be YYYY-MM-DD")
		return
	}

	today := day(c.Now.Datetime())
	in := loadPlayerStatsInput(conn, player.Team, start, end)
	result := buildPlayerStats(in, *player, start, end, today)
	c.Set("item", result)
}

// loadPlayerStatsInput 은 report.go aggregate() 와 같은 조회에 훈련·참석을 더한다.
func loadPlayerStatsInput(conn *models.Connection, teamId int, start, end string) playerStatsInput {
	var in playerStatsInput

	in.players = models.NewPlayerManager(conn).Find([]interface{}{
		models.Where{Column: "team", Value: teamId, Compare: "="},
	})

	in.matches = models.NewMatchManager(conn).Find(append(
		[]interface{}{models.Where{Column: "team", Value: teamId, Compare: "="}},
		dateRangeArgs("matchdate", start, end)...,
	))
	matchIds := make([]int, 0, len(in.matches))
	for _, m := range in.matches {
		matchIds = append(matchIds, int(m.Id))
	}
	if len(matchIds) > 0 {
		in.quarters = models.NewQuarterManager(conn).Find([]interface{}{
			models.Where{Column: "match", Value: matchIds, Compare: "in"},
		})
	}
	quarterIds := make([]int, 0, len(in.quarters))
	for _, q := range in.quarters {
		quarterIds = append(quarterIds, int(q.Id))
	}
	if len(quarterIds) > 0 {
		in.records = models.NewRecordManager(conn).Find([]interface{}{
			models.Where{Column: "quarter", Value: quarterIds, Compare: "in"},
		})
	}

	playerIds := make([]int, 0, len(in.players))
	for _, p := range in.players {
		playerIds = append(playerIds, int(p.Id))
	}
	if len(playerIds) > 0 {
		in.injuries = models.NewInjuryManager(conn).Find([]interface{}{
			models.Where{Column: "player", Value: playerIds, Compare: "in"},
		})
	}

	in.trainings = models.NewTrainingManager(conn).Find(append(
		[]interface{}{models.Where{Column: "team", Value: teamId, Compare: "="}},
		dateRangeArgs("trainingdate", start, end)...,
	))
	trainingIds := make([]int, 0, len(in.trainings))
	for _, t := range in.trainings {
		trainingIds = append(trainingIds, int(t.Id))
	}
	if len(trainingIds) > 0 {
		in.attendances = models.NewAttendanceManager(conn).Find([]interface{}{
			models.Where{Column: "training", Value: trainingIds, Compare: "in"},
		})
	}
	return in
}

// dateRangeArgs 는 datetime 컬럼에 대한 포함 범위 조건(report.go aggregate 와 동일 규칙).
func dateRangeArgs(column, start, end string) []interface{} {
	switch {
	case start != "" && end != "":
		return []interface{}{models.Where{
			Column: column, Value: [2]string{start + " 00:00:00", end + " 23:59:59"}, Compare: "between",
		}}
	case start != "":
		return []interface{}{models.Where{Column: column, Value: start + " 00:00:00", Compare: ">="}}
	case end != "":
		return []interface{}{models.Where{Column: column, Value: end + " 23:59:59", Compare: "<="}}
	}
	return nil
}

// buildPlayerStats 는 순수 집계 — DB 없이 테스트 가능.
func buildPlayerStats(in playerStatsInput, player models.Player, start, end, today string) PlayerStatsResult {
	pid := int(player.Id)

	// 쿼터 → 경기, 경기별 상대 득점
	quarterToMatch := make(map[int]int, len(in.quarters))
	quarterById := make(map[int]models.Quarter, len(in.quarters))
	awayByMatch := make(map[int]int)
	playedMatches := make(map[int]struct{})
	for _, q := range in.quarters {
		quarterToMatch[int(q.Id)] = q.Match
		quarterById[int(q.Id)] = q
		awayByMatch[q.Match] += q.Awaygoals
		playedMatches[q.Match] = struct{}{}
	}

	// 기록 → 선수별 누적, 참여 경기, 경기별 우리 팀 득점, 이 선수의 쿼터 라인
	type acc struct{ min, goal, assist, yellow, red int }
	perPlayer := make(map[int]*acc)
	matchesByPlayer := make(map[int]map[int]struct{})
	homeByMatch := make(map[int]int)
	linesByMatch := make(map[int][]QuarterLine)
	for _, r := range in.records {
		mId, okm := quarterToMatch[r.Quarter]
		if !okm {
			continue // 기간 밖 쿼터의 기록은 무시
		}
		a := perPlayer[r.Player]
		if a == nil {
			a = &acc{}
			perPlayer[r.Player] = a
		}
		a.min += r.Min
		a.goal += r.Goal
		a.assist += r.Assist
		a.yellow += r.Yellowcard
		a.red += r.Redcard
		set := matchesByPlayer[r.Player]
		if set == nil {
			set = make(map[int]struct{})
			matchesByPlayer[r.Player] = set
		}
		set[mId] = struct{}{}
		homeByMatch[mId] += r.Goal
		if r.Player == pid {
			q := quarterById[r.Quarter]
			linesByMatch[mId] = append(linesByMatch[mId], QuarterLine{
				QuarterId: q.Id, Number: q.Number,
				Min: r.Min, Goal: r.Goal, Assist: r.Assist, Yellow: r.Yellowcard, Red: r.Redcard,
			})
		}
	}

	// 부상 이력 (선수별)
	injuriesByPlayer := make(map[int][]models.Injury)
	for _, inj := range in.injuries {
		injuriesByPlayer[inj.Player] = append(injuriesByPlayer[inj.Player], inj)
	}

	// 훈련 참석 — 이미 열린(오늘까지) 훈련만 분모
	heldIds := make(map[int]struct{})
	for _, t := range in.trainings {
		if d := day(t.Trainingdate); d != "" && d <= today {
			heldIds[int(t.Id)] = struct{}{}
		}
	}
	type tacc struct{ attended, totalMin int }
	trainingByPlayer := make(map[int]*tacc)
	for _, a := range in.attendances {
		if _, held := heldIds[a.Training]; !held {
			continue
		}
		t := trainingByPlayer[a.Player]
		if t == nil {
			t = &tacc{}
			trainingByPlayer[a.Player] = t
		}
		t.attended++
		t.totalMin += a.Min
	}
	trainingLine := func(playerId int) *TrainingLine {
		held := len(heldIds)
		if held == 0 {
			return nil
		}
		t := trainingByPlayer[playerId]
		line := &TrainingLine{Held: held}
		if t != nil {
			line.Attended = t.attended
			line.TotalMin = t.totalMin
		}
		line.Rate = int(float64(line.Attended)/float64(held)*100 + 0.5)
		return line
	}

	// 스쿼드 전원 집계
	squad := make([]PlayerLine, 0, len(in.players))
	for _, p := range in.players {
		id := int(p.Id)
		line := PlayerLine{
			Id: p.Id, Name: p.Name, Number: p.Number, Position: p.Position,
			Games:       len(matchesByPlayer[id]),
			AbsentGames: absentGames(injuriesByPlayer[id], in.matches, today),
			Training:    trainingLine(id),
		}
		if a := perPlayer[id]; a != nil {
			line.Min, line.Goal, line.Assist, line.Yellow, line.Red = a.min, a.goal, a.assist, a.yellow, a.red
		}
		squad = append(squad, line)
	}
	sort.SliceStable(squad, func(i, j int) bool {
		if squad[i].Number != squad[j].Number {
			return squad[i].Number < squad[j].Number
		}
		return squad[i].Name < squad[j].Name
	})

	// 이 선수의 요약 — 스쿼드에 없으면(이론상 불가) 0 기록
	summary := PlayerLine{Id: player.Id, Name: player.Name, Number: player.Number, Position: player.Position}
	for _, line := range squad {
		if line.Id == player.Id {
			summary = line
			break
		}
	}

	// 경기별 기록 (최신순, 쿼터는 번호순)
	matchById := make(map[int]models.Match, len(in.matches))
	for _, m := range in.matches {
		matchById[int(m.Id)] = m
	}
	matches := make([]MatchLine, 0, len(linesByMatch))
	for mId, lines := range linesByMatch {
		m, okm := matchById[mId]
		if !okm {
			continue
		}
		sort.Slice(lines, func(i, j int) bool { return lines[i].Number < lines[j].Number })
		ml := MatchLine{
			MatchId: m.Id, Opponent: m.Awayname, Matchdate: m.Matchdate,
			Home: homeByMatch[mId], Away: awayByMatch[mId], Quarters: lines,
		}
		for _, l := range lines {
			ml.Min += l.Min
			ml.Goal += l.Goal
			ml.Assist += l.Assist
			ml.Yellow += l.Yellow
			ml.Red += l.Red
		}
		matches = append(matches, ml)
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Matchdate != matches[j].Matchdate {
			return matches[i].Matchdate > matches[j].Matchdate
		}
		return matches[i].MatchId > matches[j].MatchId
	})

	// 부상 이력: 이 선수, 최신 발생순
	injuries := make([]models.Injury, 0, len(injuriesByPlayer[pid]))
	injuries = append(injuries, injuriesByPlayer[pid]...)
	sort.SliceStable(injuries, func(i, j int) bool { return injuries[i].Startdate > injuries[j].Startdate })

	return PlayerStatsResult{
		Player:     player,
		Start:      start,
		End:        end,
		MatchCount: len(playedMatches),
		Summary:    summary,
		Squad:      squad,
		Matches:    matches,
		Injuries:   injuries,
	}
}
