package rest

// 리소스 소유권 검증 공용 헬퍼 (IDOR 방지).
// 모든 도메인은 team_tb.t_user 를 뿌리로 하는 소유 체인을 가진다:
//   team ← player ← injury
//   team ← match ← quarter ← record
// 화면에서 타 팀 리소스를 숨기는 것과 별개로, API 직접 호출을 서버가 거부해야 한다.

import (
	"errors"
	"fmt"

	"fotstat/controllers"
	"fotstat/models"
)

var errForbidden = errors.New("forbidden: resource does not belong to you")
var errNotFound = errors.New("not found")

// requestUser 는 JwtAuthRequired 미들웨어가 세팅한 요청 사용자. 없으면 nil.
func requestUser(c *controllers.Controller) *models.User {
	if c.Context == nil {
		return nil
	}
	user, ok := c.Context.Locals("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}

func ownsTeam(conn *models.Connection, user *models.User, teamId int) bool {
	if user == nil || teamId == 0 {
		return false
	}
	team := models.NewTeamManager(conn).Get(int64(teamId))
	return team != nil && int64(team.User) == user.Id
}

func ownsPlayer(conn *models.Connection, user *models.User, playerId int) bool {
	if user == nil || playerId == 0 {
		return false
	}
	player := models.NewPlayerManager(conn).Get(int64(playerId))
	if player == nil {
		return false
	}
	return ownsTeam(conn, user, player.Team)
}

func ownsMatch(conn *models.Connection, user *models.User, matchId int) bool {
	if user == nil || matchId == 0 {
		return false
	}
	match := models.NewMatchManager(conn).Get(int64(matchId))
	if match == nil {
		return false
	}
	return ownsTeam(conn, user, match.Team)
}

func ownsQuarter(conn *models.Connection, user *models.User, quarterId int) bool {
	if user == nil || quarterId == 0 {
		return false
	}
	quarter := models.NewQuarterManager(conn).Get(int64(quarterId))
	if quarter == nil {
		return false
	}
	return ownsMatch(conn, user, quarter.Match)
}

// ownsRecordTarget 은 record 대상(쿼터, 선수)이 모두 요청 사용자 소유이면서
// 선수가 그 쿼터가 속한 경기의 팀 소속인지 확인한다. 같은 사용자가 여러 팀을
// 가질 때 다른 팀 선수의 기록이 섞여 통계가 오염되는 것을 막는다.
func ownsRecordTarget(conn *models.Connection, user *models.User, quarterId int, playerId int) bool {
	if user == nil || quarterId == 0 || playerId == 0 {
		return false
	}
	quarter := models.NewQuarterManager(conn).Get(int64(quarterId))
	if quarter == nil {
		return false
	}
	match := models.NewMatchManager(conn).Get(int64(quarter.Match))
	if match == nil || !ownsTeam(conn, user, match.Team) {
		return false
	}
	player := models.NewPlayerManager(conn).Get(int64(playerId))
	return player != nil && player.Team == match.Team
}

// Index/Count 용 강제 스코프 — 클라이언트 필터와 무관하게 AND 로 결합되어
// 요청 사용자 소유 범위 밖의 행은 조회되지 않는다. user.Id 는 JWT 유래 정수라 안전.

func ownTeamScope(user *models.User) models.Custom {
	return models.Custom{Query: fmt.Sprintf("t_user = %d", user.Id)}
}

func ownPlayerScope(user *models.User) models.Custom {
	return models.Custom{Query: fmt.Sprintf("p_team in (select t_id from team_tb where t_user = %d)", user.Id)}
}

func ownMatchScope(user *models.User) models.Custom {
	return models.Custom{Query: fmt.Sprintf("m_team in (select t_id from team_tb where t_user = %d)", user.Id)}
}

func ownQuarterScope(user *models.User) models.Custom {
	return models.Custom{Query: fmt.Sprintf("q_match in (select m_id from match_tb join team_tb on m_team = t_id where t_user = %d)", user.Id)}
}

func ownRecordScope(user *models.User) models.Custom {
	return models.Custom{Query: fmt.Sprintf("r_quarter in (select q_id from quarter_tb join match_tb on q_match = m_id join team_tb on m_team = t_id where t_user = %d)", user.Id)}
}

func ownUserScope(user *models.User) models.Custom {
	return models.Custom{Query: fmt.Sprintf("u_id = %d", user.Id)}
}

func ownInjuryScope(user *models.User) models.Custom {
	return models.Custom{Query: fmt.Sprintf("i_player in (select p_id from player_tb join team_tb on p_team = t_id where t_user = %d)", user.Id)}
}
