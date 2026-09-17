package rest

import (
	"encoding/json"
	"testing"

	"fotstat/models"
)

// 고정 시나리오 (웹 playerMatchLog.test.ts / aggregateTeamStats.test.ts 와 같은 모양):
//
//	경기 1(06-01, A팀): Q11(15', 실점1), Q12(15', 실점0)
//	경기 2(06-08, B팀): Q21(20', 실점2)
//	경기 3(06-15, C팀): 쿼터 없음, today(06-14) 이후의 예정 경기
//	선수 100: Q11 15' 1G 1A 🟨, Q12 15' 1G 0A, Q21 20' 0G 0A 🟥
//	선수 200: Q11 15' 0G 0A, Q21 20' 1G 0A
//	선수 300: 기록 없음, 06-02~06-10 부상(경기 2 결장)
//	훈련: 06-03(열림), 06-05(열림), 06-30(미래) — 100은 2회 참석(60+50분), 200은 1회
func fixture() (playerStatsInput, []models.Player) {
	players := []models.Player{
		{Id: 100, Team: 1, Name: "홍길동", Number: 10, Position: "ST"},
		{Id: 200, Team: 1, Name: "김철수", Number: 7, Position: "CM"},
		{Id: 300, Team: 1, Name: "이영희", Number: 1, Position: "GK"},
	}
	in := playerStatsInput{
		players: players,
		matches: []models.Match{
			{Id: 1, Team: 1, Awayname: "A팀", Matchdate: "2026-06-01 10:00:00"},
			{Id: 2, Team: 1, Awayname: "B팀", Matchdate: "2026-06-08 10:00:00"},
			{Id: 3, Team: 1, Awayname: "C팀", Matchdate: "2026-06-15 10:00:00"},
		},
		quarters: []models.Quarter{
			{Id: 11, Match: 1, Number: 1, Duration: 15, Awaygoals: 1},
			{Id: 12, Match: 1, Number: 2, Duration: 15, Awaygoals: 0},
			{Id: 21, Match: 2, Number: 1, Duration: 20, Awaygoals: 2},
		},
		records: []models.Record{
			{Id: 1, Quarter: 11, Player: 100, Min: 15, Goal: 1, Assist: 1, Yellowcard: 1},
			{Id: 2, Quarter: 12, Player: 100, Min: 15, Goal: 1},
			{Id: 3, Quarter: 21, Player: 100, Min: 20, Redcard: 1},
			{Id: 4, Quarter: 11, Player: 200, Min: 15},
			{Id: 5, Quarter: 21, Player: 200, Min: 20, Goal: 1},
			{Id: 6, Quarter: 999, Player: 100, Min: 90, Goal: 9}, // 기간 밖 쿼터 — 무시돼야 함
		},
		injuries: []models.Injury{
			{Id: 1, Player: 300, Type: "발목", Startdate: "2026-06-02", Returndate: "2026-06-10"},
			{Id: 2, Player: 100, Type: "햄스트링", Startdate: "2026-05-01", Returndate: "2026-05-20"},
			{Id: 3, Player: 100, Type: "무릎", Startdate: "2026-05-25", Returndate: ""},
		},
		trainings: []models.Training{
			{Id: 1, Team: 1, Trainingdate: "2026-06-03 18:00:00"},
			{Id: 2, Team: 1, Trainingdate: "2026-06-05 18:00:00"},
			{Id: 3, Team: 1, Trainingdate: "2026-06-30 18:00:00"},
		},
		attendances: []models.Attendance{
			{Id: 1, Training: 1, Player: 100, Min: 60},
			{Id: 2, Training: 2, Player: 100, Min: 50},
			{Id: 3, Training: 1, Player: 200, Min: 60},
			{Id: 4, Training: 3, Player: 200, Min: 60}, // 미래 훈련 — 분모·분자 모두 제외
		},
	}
	return in, players
}

// 경기 3(06-15)이 아직 열리지 않은 시점 — 미래 경기는 결장에서 제외돼야 한다.
const fixtureToday = "2026-06-14"

func TestBuildPlayerStats_Summary(t *testing.T) {
	in, players := fixture()
	r := buildPlayerStats(in, players[0], "2026-06-01", "2026-06-30", fixtureToday)

	s := r.Summary
	if s.Id != 100 || s.Games != 2 || s.Min != 50 || s.Goal != 2 || s.Assist != 1 || s.Yellow != 1 || s.Red != 1 {
		t.Errorf("summary = %+v", s)
	}
	// 부상 3(무릎, 05-25~진행중)이 경기 1·2를 덮는다 → 결장 2 (경기 3은 today 이후라 제외)
	if s.AbsentGames != 2 {
		t.Errorf("absentGames = %d want 2", s.AbsentGames)
	}
	if r.MatchCount != 2 {
		t.Errorf("matchCount = %d want 2 (쿼터 있는 경기만)", r.MatchCount)
	}
	if r.Start != "2026-06-01" || r.End != "2026-06-30" {
		t.Errorf("range echo = %q..%q", r.Start, r.End)
	}
}

func TestBuildPlayerStats_Squad(t *testing.T) {
	in, players := fixture()
	r := buildPlayerStats(in, players[0], "", "", fixtureToday)

	if len(r.Squad) != 3 {
		t.Fatalf("squad len = %d", len(r.Squad))
	}
	// 등번호 순: 1(이영희), 7(김철수), 10(홍길동)
	if r.Squad[0].Id != 300 || r.Squad[1].Id != 200 || r.Squad[2].Id != 100 {
		t.Errorf("squad order = %d,%d,%d", r.Squad[0].Id, r.Squad[1].Id, r.Squad[2].Id)
	}
	gk := r.Squad[0]
	if gk.Games != 0 || gk.Goal != 0 || gk.AbsentGames != 1 {
		t.Errorf("기록 없는 선수도 0 기록 + 결장으로 포함돼야: %+v", gk)
	}
	if r.Squad[1].Goal != 1 || r.Squad[1].Games != 2 {
		t.Errorf("김철수 = %+v", r.Squad[1])
	}
}

func TestBuildPlayerStats_Matches(t *testing.T) {
	in, players := fixture()
	r := buildPlayerStats(in, players[0], "", "", fixtureToday)

	if len(r.Matches) != 2 {
		t.Fatalf("matches len = %d", len(r.Matches))
	}
	// 최신순
	m2, m1 := r.Matches[0], r.Matches[1]
	if m2.MatchId != 2 || m1.MatchId != 1 {
		t.Errorf("order = %d,%d", m2.MatchId, m1.MatchId)
	}
	// 경기 1: 우리 팀 골 = 선수100 2 + 선수200 0 = 2, 상대 1. 쿼터 2개 번호순.
	if m1.Opponent != "A팀" || m1.Home != 2 || m1.Away != 1 || m1.Min != 30 || m1.Goal != 2 || m1.Assist != 1 || m1.Yellow != 1 {
		t.Errorf("match1 = %+v", m1)
	}
	if len(m1.Quarters) != 2 || m1.Quarters[0].QuarterId != 11 || m1.Quarters[1].QuarterId != 12 || m1.Quarters[0].Number != 1 {
		t.Errorf("match1 quarters = %+v", m1.Quarters)
	}
	// 경기 2: 우리 팀 골 = 선수200 1, 상대 2, 선수100 레드 1
	if m2.Home != 1 || m2.Away != 2 || m2.Red != 1 || m2.Min != 20 {
		t.Errorf("match2 = %+v", m2)
	}
}

func TestBuildPlayerStats_InjuriesAndTraining(t *testing.T) {
	in, players := fixture()
	r := buildPlayerStats(in, players[0], "2026-06-01", "2026-06-30", fixtureToday)

	// 부상 이력은 기간과 무관하게 전체, 최신 발생순
	if len(r.Injuries) != 2 || r.Injuries[0].Id != 3 || r.Injuries[1].Id != 2 {
		t.Errorf("injuries = %+v", r.Injuries)
	}
	// 훈련: 열린 훈련 2회(06-30 제외). 100: 2/2 100% 110분, 200: 1/2 50%, 300: 0/2 0%
	if s := r.Summary.Training; s == nil || s.Attended != 2 || s.Held != 2 || s.Rate != 100 || s.TotalMin != 110 {
		t.Errorf("training(100) = %+v", s)
	}
	var p200, p300 *TrainingLine
	for _, l := range r.Squad {
		switch l.Id {
		case 200:
			p200 = l.Training
		case 300:
			p300 = l.Training
		}
	}
	if p200 == nil || p200.Attended != 1 || p200.Rate != 50 {
		t.Errorf("training(200) = %+v", p200)
	}
	if p300 == nil || p300.Attended != 0 || p300.Rate != 0 || p300.Held != 2 {
		t.Errorf("training(300) = %+v", p300)
	}
}

func TestBuildPlayerStats_NoTrainingsIsNull(t *testing.T) {
	in, players := fixture()
	in.trainings = nil
	in.attendances = nil
	r := buildPlayerStats(in, players[0], "", "", fixtureToday)
	if r.Summary.Training != nil {
		t.Errorf("훈련이 없으면 training=null 이어야: %+v", r.Summary.Training)
	}
	// 미래 훈련만 있어도 null
	in.trainings = []models.Training{{Id: 9, Team: 1, Trainingdate: "2027-01-01 10:00:00"}}
	r = buildPlayerStats(in, players[0], "", "", fixtureToday)
	if r.Summary.Training != nil {
		t.Errorf("미래 훈련만 있으면 training=null 이어야: %+v", r.Summary.Training)
	}
}

func TestBuildPlayerStats_EmptyInput(t *testing.T) {
	p := models.Player{Id: 5, Team: 1, Name: "신입", Number: 99}
	r := buildPlayerStats(playerStatsInput{players: []models.Player{p}}, p, "", "", fixtureToday)
	if r.Summary.Games != 0 || r.MatchCount != 0 || len(r.Matches) != 0 || len(r.Injuries) != 0 || len(r.Squad) != 1 {
		t.Errorf("empty = %+v", r)
	}
	// 빈 슬라이스가 JSON null 이 아니라 [] 로 나가야 클라이언트가 안전하다
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"matches", "injuries", "squad"} {
		if string(m[k]) == "null" {
			t.Errorf("%s 가 null 로 직렬화됨", k)
		}
	}
}

func TestDateRangeArgs(t *testing.T) {
	if got := dateRangeArgs("matchdate", "", ""); got != nil {
		t.Errorf("빈 범위는 조건 없음: %+v", got)
	}
	both := dateRangeArgs("matchdate", "2026-06-01", "2026-06-30")
	w, ok := both[0].(models.Where)
	if !ok || w.Compare != "between" || w.Value.([2]string)[1] != "2026-06-30 23:59:59" {
		t.Errorf("between = %+v", both)
	}
	only := dateRangeArgs("trainingdate", "2026-06-01", "")
	w = only[0].(models.Where)
	if w.Compare != ">=" || w.Value != "2026-06-01 00:00:00" {
		t.Errorf(">= = %+v", only)
	}
}

func TestDayRe(t *testing.T) {
	for _, ok := range []string{"2026-06-01", "1999-12-31"} {
		if !dayRe.MatchString(ok) {
			t.Errorf("%q should match", ok)
		}
	}
	for _, bad := range []string{"2026-6-1", "2026-06-01 10:00:00", "abc", "2026-06-01'"} {
		if dayRe.MatchString(bad) {
			t.Errorf("%q should not match", bad)
		}
	}
}
