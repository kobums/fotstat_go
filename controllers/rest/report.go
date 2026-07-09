package rest

// 경기기록표 xlsx 리포트 생성.
// 웹 프론트가 exceljs로 만들던 '경기기록표' 양식(matchRecordSheet.ts)을 백엔드로 이관해
// 웹·iOS가 동일한 엔드포인트로 같은 파일을 내려받게 한다.
// 집계 규칙은 웹의 aggregateTeamStats.ts / lib/injury.ts 와 동일하게 맞춘다:
//   - 경기수 = 선수가 기록을 남긴 경기 수
//   - 총 경기시간 = 그 경기들의 쿼터 duration 합
//   - 출전 시간 = 기록된 min 합, 골/도움 = 기록 합
//   - 부상(결장) = 부상 기간에 걸친 팀 경기 수(오늘까지, 발생일 당일 제외)

import (
	"fmt"
	"sort"
	"strings"

	"fotstat/controllers"
	"fotstat/models"

	"github.com/xuri/excelize/v2"
)

type ReportController struct {
	controllers.Controller
}

// ── 원본 양식(matchRecordSheet.ts)에서 이식한 서식 상수 ──
const (
	reportFont      = "Malgun Gothic"
	reportSheetName = "경기기록표"
	fillTitle       = "FFFFCC" // 제목·이름 칸 연노랑
	fillHeader      = "CCFFFF" // 헤더 연하늘
	rowH            = 15.85
	headerH         = 16.25
)

// A(여백) + B~J 9개 컬럼 폭 — 원본 그대로.
var reportColWidths = []float64{3.6, 6.8, 7.8, 7.8, 7.8, 7.8, 7.8, 7.8, 7.8, 7.8}

var reportHeaders = []string{
	"이름", "총 경기수", "총 경기시간", "출전 시간", "포지션",
	"득점", "도움", "부상", "비고",
}

// playerStat 은 한 선수의 집계 행.
type playerStat struct {
	name        string
	number      int
	position    string
	games       int
	min         int
	goal        int
	assist      int
	totalTime   int
	absentGames int
}

// BuildMatchRecord 는 경기기록표 xlsx 바이트와 다운로드 파일명을 만든다.
// 실패 시 컨트롤러 Result 에 에러를 세팅하고 ok=false 를 반환한다(라우터가 JSON 에러 응답).
func (c *ReportController) BuildMatchRecord() (data []byte, filename string, ok bool) {
	conn := c.NewConnection()

	user := requestUser(&c.Controller)
	if user == nil {
		c.Code = 403
		c.Error(errForbidden)
		return nil, "", false
	}

	teamId := c.Geti("team")
	if !ownsTeam(conn, user, teamId) {
		c.Code = 403
		c.Error(errForbidden)
		return nil, "", false
	}
	team := models.NewTeamManager(conn).Get(int64(teamId))
	if team == nil {
		c.Code = 404
		c.Error(errNotFound)
		return nil, "", false
	}

	start := c.Get("start") // "YYYY-MM-DD" 또는 ""
	end := c.Get("end")
	today := c.Now.Datetime()
	if len(today) >= 10 {
		today = today[:10]
	}

	stats := c.aggregate(conn, teamId, start, end, today)

	title := recordSheetTitle(start, end)
	buf, err := renderMatchRecordSheet(stats, title)
	if err != nil {
		c.Code = 500
		c.Set("code", "error")
		c.Set("message", err.Error())
		return nil, "", false
	}

	return buf, safeFilename(fmt.Sprintf("%s %s.xlsx", team.Name, title)), true
}

// safeFilename 은 파일시스템 금지 문자를 제거해 다운로드 파일명으로 안전하게 만든다.
// 팀 이름에 '/' 등이 섞이면 클라이언트(특히 iOS)가 파일 경로를 만들 때 깨지므로
// 서버에서 미리 치환한다.
func safeFilename(name string) string {
	forbidden := `/\:*?"<>|`
	out := make([]rune, 0, len(name))
	for _, r := range name {
		if r < 0x20 || strings.ContainsRune(forbidden, r) {
			out = append(out, '_')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

// aggregate 는 팀·기간에 대한 선수별 집계를 로스터 순(등번호→이름)으로 반환한다.
func (c *ReportController) aggregate(conn *models.Connection, teamId int, start, end, today string) []playerStat {
	// 1) 기간 내 팀 경기
	matchArgs := []interface{}{
		models.Where{Column: "team", Value: teamId, Compare: "="},
	}
	if start != "" && end != "" {
		matchArgs = append(matchArgs, models.Where{
			Column: "matchdate", Value: [2]string{start + " 00:00:00", end + " 23:59:59"}, Compare: "between",
		})
	} else if start != "" {
		matchArgs = append(matchArgs, models.Where{Column: "matchdate", Value: start + " 00:00:00", Compare: ">="})
	} else if end != "" {
		matchArgs = append(matchArgs, models.Where{Column: "matchdate", Value: end + " 23:59:59", Compare: "<="})
	}
	matches := models.NewMatchManager(conn).Find(matchArgs)

	matchIds := make([]int, 0, len(matches))
	for _, m := range matches {
		matchIds = append(matchIds, int(m.Id))
	}

	// 2) 그 경기들의 쿼터 → 경기별 총 시간, 쿼터→경기 매핑
	minutesByMatch := make(map[int]int)
	quarterToMatch := make(map[int]int)
	quarterIds := make([]int, 0)
	if len(matchIds) > 0 {
		quarters := models.NewQuarterManager(conn).Find([]interface{}{
			models.Where{Column: "match", Value: matchIds, Compare: "in"},
		})
		for _, q := range quarters {
			minutesByMatch[q.Match] += q.Duration
			quarterToMatch[int(q.Id)] = q.Match
			quarterIds = append(quarterIds, int(q.Id))
		}
	}

	// 3) 그 쿼터들의 기록 → 선수별 min/goal/assist, 참여 경기 집합
	type acc struct{ min, goal, assist int }
	perPlayer := make(map[int]*acc)
	matchesByPlayer := make(map[int]map[int]struct{})
	if len(quarterIds) > 0 {
		records := models.NewRecordManager(conn).Find([]interface{}{
			models.Where{Column: "quarter", Value: quarterIds, Compare: "in"},
		})
		for _, r := range records {
			a := perPlayer[r.Player]
			if a == nil {
				a = &acc{}
				perPlayer[r.Player] = a
			}
			a.min += r.Min
			a.goal += r.Goal
			a.assist += r.Assist
			if mId, okm := quarterToMatch[r.Quarter]; okm {
				set := matchesByPlayer[r.Player]
				if set == nil {
					set = make(map[int]struct{})
					matchesByPlayer[r.Player] = set
				}
				set[mId] = struct{}{}
			}
		}
	}

	// 4) 팀 선수단 + 부상 이력
	players := models.NewPlayerManager(conn).Find([]interface{}{
		models.Where{Column: "team", Value: teamId, Compare: "="},
	})
	playerIds := make([]int, 0, len(players))
	for _, p := range players {
		playerIds = append(playerIds, int(p.Id))
	}
	injuriesByPlayer := make(map[int][]models.Injury)
	if len(playerIds) > 0 {
		injuries := models.NewInjuryManager(conn).Find([]interface{}{
			models.Where{Column: "player", Value: playerIds, Compare: "in"},
		})
		for _, inj := range injuries {
			injuriesByPlayer[inj.Player] = append(injuriesByPlayer[inj.Player], inj)
		}
	}

	// 5) 선수별 행 조립
	stats := make([]playerStat, 0, len(players))
	for _, p := range players {
		pid := int(p.Id)
		st := playerStat{
			name:     p.Name,
			number:   p.Number,
			position: p.Position,
		}
		if a := perPlayer[pid]; a != nil {
			st.min = a.min
			st.goal = a.goal
			st.assist = a.assist
		}
		for mId := range matchesByPlayer[pid] {
			st.games++
			st.totalTime += minutesByMatch[mId]
		}
		st.absentGames = absentGames(injuriesByPlayer[pid], matches, today)
		stats = append(stats, st)
	}

	// 원본 양식과 같은 로스터 순(등번호→이름)
	sort.SliceStable(stats, func(i, j int) bool {
		if stats[i].number != stats[j].number {
			return stats[i].number < stats[j].number
		}
		return stats[i].name < stats[j].name
	})
	return stats
}

// absentGames 는 부상 기간에 걸친(오늘까지 열린) 팀 경기 수를 센다.
// lib/injury.ts 의 injuryCoversMatch / absentGamesFor 와 동일 규칙.
func absentGames(spells []models.Injury, matches []models.Match, today string) int {
	if len(spells) == 0 {
		return 0
	}
	cnt := 0
	for _, m := range matches {
		d := day(m.Matchdate)
		if d == "" || d > today {
			continue // 아직 열리지 않은 경기는 결장 아님
		}
		for _, inj := range spells {
			if injuryCoversMatch(inj, m.Matchdate) {
				cnt++
				break
			}
		}
	}
	return cnt
}

// injuryCoversMatch: 발생일 다음 날부터 복귀일 당일까지의 경기를 부상 기간으로 본다.
// 복귀일 빈값("") = 아직 부상 중. 백엔드 record.go injuryConflict 와 같은 날짜 규칙.
func injuryCoversMatch(inj models.Injury, matchdate string) bool {
	start := day(inj.Startdate)
	if start == "" {
		return false
	}
	d := day(matchdate)
	if d == "" || d <= start {
		return false
	}
	end := day(inj.Returndate)
	return end == "" || d <= end
}

func day(date string) string {
	if len(date) >= 10 {
		return date[:10]
	}
	return date
}

// recordSheetTitle: 기간이 한 달 안이면 "M월 경기기록표", 부분 지정이면 범위, 없으면 전체.
// 웹 SeasonReportPage.recordSheetTitle 과 동일.
func recordSheetTitle(start, end string) string {
	if start != "" && end != "" && start[:7] == end[:7] {
		month := start[5:7]
		if month[0] == '0' {
			month = month[1:]
		}
		return fmt.Sprintf("%s월 경기기록표", month)
	}
	if start != "" || end != "" {
		s := "처음"
		if start != "" {
			s = dotDate(start)
		}
		e := "오늘"
		if end != "" {
			e = dotDate(end)
		}
		return fmt.Sprintf("%s~%s 경기기록표", s, e)
	}
	return "전체 경기기록표"
}

func dotDate(d string) string {
	out := make([]byte, 0, len(d))
	for i := 0; i < len(d); i++ {
		if d[i] == '-' {
			out = append(out, '.')
		} else {
			out = append(out, d[i])
		}
	}
	return string(out)
}

// renderMatchRecordSheet 은 집계 결과로 원본 양식 그대로의 xlsx 바이트를 만든다.
func renderMatchRecordSheet(players []playerStat, title string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	idx, err := f.NewSheet(reportSheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(idx)
	_ = f.DeleteSheet("Sheet1")
	sheet := reportSheetName

	// 컬럼 폭(A~J)
	for i, w := range reportColWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		if e := f.SetColWidth(sheet, col, col, w); e != nil {
			return nil, e
		}
	}

	medium := 2 // excelize border style: continuous medium
	double := 6 // double line

	box := func(fill string, fontSize float64, bold bool, bottom int) (int, error) {
		style := &excelize.Style{
			Font:      &excelize.Font{Family: reportFont, Size: fontSize, Bold: bold},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: medium},
				{Type: "right", Color: "000000", Style: medium},
			},
		}
		if bottom != 0 {
			style.Border = append(style.Border, excelize.Border{Type: "bottom", Color: "000000", Style: bottom})
		}
		if fill != "" {
			style.Fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{fill}}
		}
		return f.NewStyle(style)
	}

	// 제목: B1:J2 병합, 18pt bold, 연노랑, medium 박스
	titleStyle, err := box(fillTitle, 18, true, medium)
	if err != nil {
		return nil, err
	}
	if e := f.MergeCell(sheet, "B1", "J2"); e != nil {
		return nil, e
	}
	if e := f.SetCellValue(sheet, "B1", title); e != nil {
		return nil, e
	}
	if e := f.SetCellStyle(sheet, "B1", "J2", titleStyle); e != nil {
		return nil, e
	}
	_ = f.SetRowHeight(sheet, 1, rowH)
	_ = f.SetRowHeight(sheet, 2, rowH)

	// 헤더(3행): 이름만 11pt, 나머지 8pt. 연하늘, bottom medium
	headerName, err := box(fillHeader, 11, true, medium)
	if err != nil {
		return nil, err
	}
	headerOther, err := box(fillHeader, 8, true, medium)
	if err != nil {
		return nil, err
	}
	for i, h := range reportHeaders {
		cell, _ := excelize.CoordinatesToCellName(2+i, 3)
		if e := f.SetCellValue(sheet, cell, h); e != nil {
			return nil, e
		}
		st := headerOther
		if i == 0 {
			st = headerName
		}
		if e := f.SetCellStyle(sheet, cell, cell, st); e != nil {
			return nil, e
		}
	}
	_ = f.SetRowHeight(sheet, 3, headerH)

	// 선수 블록: 선수당 2행 세로 병합, 블록 사이 double, 마지막 블록 medium 마감
	// 스타일 캐시(이름/일반 × 하단 medium/double)
	nameMedium, err := box(fillTitle, 9, true, medium)
	if err != nil {
		return nil, err
	}
	nameDouble, err := box(fillTitle, 9, true, double)
	if err != nil {
		return nil, err
	}
	valMedium, err := box("", 11, false, medium)
	if err != nil {
		return nil, err
	}
	valDouble, err := box("", 11, false, double)
	if err != nil {
		return nil, err
	}

	for idx, p := range players {
		top := 4 + idx*2
		bottom := top + 1
		isLast := idx == len(players)-1

		values := []interface{}{
			p.name,
			p.games,
			fmt.Sprintf("%d'", p.totalTime),
			p.min,
			p.position,
			blankIfZero(p.goal),
			blankIfZero(p.assist),
			absentLabel(p.absentGames),
			"",
		}
		for i, v := range values {
			col := 2 + i
			topCell, _ := excelize.CoordinatesToCellName(col, top)
			botCell, _ := excelize.CoordinatesToCellName(col, bottom)
			if e := f.MergeCell(sheet, topCell, botCell); e != nil {
				return nil, e
			}
			if s, isStr := v.(string); !isStr || s != "" {
				if e := f.SetCellValue(sheet, topCell, v); e != nil {
					return nil, e
				}
			}
			var st int
			if i == 0 {
				if isLast {
					st = nameMedium
				} else {
					st = nameDouble
				}
			} else {
				if isLast {
					st = valMedium
				} else {
					st = valDouble
				}
			}
			if e := f.SetCellStyle(sheet, topCell, botCell, st); e != nil {
				return nil, e
			}
		}
		_ = f.SetRowHeight(sheet, top, rowH)
		_ = f.SetRowHeight(sheet, bottom, rowH)
	}

	out, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// blankIfZero: 원본 양식처럼 득점/도움이 0이면 빈칸.
func blankIfZero(n int) interface{} {
	if n == 0 {
		return ""
	}
	return n
}

func absentLabel(n int) string {
	if n > 0 {
		return fmt.Sprintf("결장%d", n)
	}
	return ""
}
