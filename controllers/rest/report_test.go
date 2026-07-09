package rest

import (
	"bytes"
	"testing"

	"fotstat/models"

	"github.com/xuri/excelize/v2"
)

func mkInjury(start, ret string) models.Injury {
	return models.Injury{Startdate: start, Returndate: ret}
}

func TestRecordSheetTitle(t *testing.T) {
	cases := []struct{ start, end, want string }{
		{"2026-06-01", "2026-06-30", "6월 경기기록표"},
		{"2026-11-01", "2026-11-05", "11월 경기기록표"},
		{"2026-06-01", "2026-07-15", "2026.06.01~2026.07.15 경기기록표"},
		{"", "", "전체 경기기록표"},
		{"2026-06-01", "", "2026.06.01~오늘 경기기록표"},
		{"", "2026-06-30", "처음~2026.06.30 경기기록표"},
	}
	for _, c := range cases {
		if got := recordSheetTitle(c.start, c.end); got != c.want {
			t.Errorf("recordSheetTitle(%q,%q)=%q want %q", c.start, c.end, got, c.want)
		}
	}
}

func TestSafeFilename(t *testing.T) {
	cases := []struct{ in, want string }{
		{"FC서울 6월 경기기록표.xlsx", "FC서울 6월 경기기록표.xlsx"},
		{"A/B FC 6월 경기기록표.xlsx", "A_B FC 6월 경기기록표.xlsx"},
		{`팀:*?"<>| 전체 경기기록표.xlsx`, "팀_______ 전체 경기기록표.xlsx"},
	}
	for _, c := range cases {
		if got := safeFilename(c.in); got != c.want {
			t.Errorf("safeFilename(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestInjuryCoversMatch(t *testing.T) {
	inj := mkInjury("2026-06-10", "2026-06-20")
	// 발생일 당일은 미포함, 다음 날부터 복귀일 당일까지 포함
	if injuryCoversMatch(inj, "2026-06-10 10:00:00") {
		t.Error("발생일 당일 경기는 부상 기간이 아님")
	}
	if !injuryCoversMatch(inj, "2026-06-11 10:00:00") {
		t.Error("발생일 다음 날은 부상 기간")
	}
	if !injuryCoversMatch(inj, "2026-06-20 10:00:00") {
		t.Error("복귀일 당일은 부상 기간")
	}
	if injuryCoversMatch(inj, "2026-06-21 10:00:00") {
		t.Error("복귀일 다음 날은 아님")
	}
	// 복귀일 없음 = 계속 부상 중
	open := mkInjury("2026-06-10", "")
	if !injuryCoversMatch(open, "2027-01-01 10:00:00") {
		t.Error("복귀일 빈값이면 이후 경기 계속 포함")
	}
}

func TestRenderMatchRecordSheet(t *testing.T) {
	players := []playerStat{
		{name: "홍길동", number: 7, position: "FW", games: 3, min: 210, goal: 4, assist: 2, totalTime: 240, absentGames: 0},
		{name: "김철수", number: 10, position: "MF", games: 2, min: 120, goal: 0, assist: 0, totalTime: 160, absentGames: 1},
	}
	buf, err := renderMatchRecordSheet(players, "6월 경기기록표")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	sheet := reportSheetName
	get := func(cell string) string {
		v, _ := f.GetCellValue(sheet, cell)
		return v
	}

	if got := get("B1"); got != "6월 경기기록표" {
		t.Errorf("title B1=%q", got)
	}
	if got := get("B3"); got != "이름" {
		t.Errorf("header B3=%q", got)
	}
	if got := get("J3"); got != "비고" {
		t.Errorf("header J3=%q", got)
	}
	// 첫 선수 블록(4행): 이름/경기수/총시간/출전/포지션/득점/도움/부상
	if got := get("B4"); got != "홍길동" {
		t.Errorf("B4=%q", got)
	}
	if got := get("C4"); got != "3" {
		t.Errorf("games C4=%q", got)
	}
	if got := get("D4"); got != "240'" {
		t.Errorf("totalTime D4=%q", got)
	}
	if got := get("E4"); got != "210" {
		t.Errorf("min E4=%q", got)
	}
	if got := get("G4"); got != "4" {
		t.Errorf("goal G4=%q", got)
	}
	// 둘째 선수(6행): 득점 0 → 빈칸, 부상 1 → 결장1
	if got := get("G6"); got != "" {
		t.Errorf("goal 0 should be blank, G6=%q", got)
	}
	if got := get("I6"); got != "결장1" {
		t.Errorf("absent I6=%q", got)
	}
}

