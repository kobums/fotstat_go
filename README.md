<div align="center">

# ⚽ Fotstat — API Server

**Go (Fiber) 기반 축구 경기 통계 및 기록 관리 REST API 서버**

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ACD7?style=flat-square&logo=go&logoColor=white)](https://gofiber.io/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat-square&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![Docker](https://img.shields.io/badge/Docker-지원-2496ED?style=flat-square&logo=docker&logoColor=white)](https://www.docker.com/)

</div>

---

## 개요

팀(Team), 선수(Player), 경기(Match), 쿼터(Quarter), 기록(Record) 데이터를 관리하는 RESTful API 서버입니다.
JWT 기반 인증, bcrypt 암호화, Docker 배포를 지원합니다.

---

## 기술 스택

| 구분 | 기술 |
|---|---|
| 언어 | Go 1.26 |
| 웹 프레임워크 | Fiber v2 |
| 데이터베이스 | MySQL 8.0 |
| 인증 | JWT + bcrypt |
| 로깅 | Zerolog |
| 배포 | Docker / Docker Compose |

---

## 프로젝트 구조

```
fotstat_go/
├── main.go
├── services/
│   └── http.go                  # Fiber 서버 설정 (CORS, 정적파일 등)
├── router/
│   ├── router.go                # 라우터 초기화 및 그룹 설정
│   ├── auth.go                  # JWT 미들웨어 및 인증 검증
│   └── routers/                 # 도메인별 라우트
│       ├── user.go
│       ├── team.go
│       ├── player.go
│       ├── match.go
│       ├── quarter.go           # PUT /quarter/awaygoals 포함
│       ├── record.go            # PUT /record/stats 포함
│       └── upload.go
├── controllers/
│   └── rest/                    # 도메인별 비즈니스 로직
│       ├── user.go
│       ├── team.go
│       ├── player.go
│       ├── match.go
│       ├── quarter.go
│       └── record.go
├── models/                      # 도메인 모델 및 DB 연결
│   ├── db.go
│   ├── cache.go
│   ├── user.go
│   ├── team.go
│   ├── player.go
│   ├── match.go
│   ├── quarter.go
│   └── record.go
├── global/
│   ├── config/                  # 환경 설정 로더
│   ├── jwt/                     # JWT 생성 및 bcrypt 유틸리티
│   ├── log/                     # 로거 초기화
│   └── setting/                 # 전역 설정 싱글톤
├── go-basic.sql                 # 기본 스키마 SQL
├── migration_add_quarter_duration.sql
├── migration_add_quarter_awaygoals.sql
├── docker-compose.yml
├── dockerfile
├── Makefile
└── api_postman_collection.json  # Postman 테스트 컬렉션
```

---

## API 엔드포인트

모든 엔드포인트는 `/api` 접두사를 사용하며, Auth를 제외한 모든 요청에 **Bearer Token** 인증이 필요합니다.

### Auth

| Method | Path | 설명 |
|---|---|---|
| `POST` | `/api/user` | 회원가입 |
| `GET` | `/api/jwt` | 로그인 (JWT 발급) |

### Team

| Method | Path | 설명 |
|---|---|---|
| `GET` | `/api/team?user={id}` | 팀 목록 조회 |
| `POST` | `/api/team` | 팀 생성 |
| `PUT` | `/api/team` | 팀 수정 |
| `DELETE` | `/api/team` | 팀 삭제 |

### Player

| Method | Path | 설명 |
|---|---|---|
| `GET` | `/api/player?team={id}` | 선수 목록 조회 |
| `POST` | `/api/player` | 선수 추가 |
| `PUT` | `/api/player` | 선수 수정 |
| `DELETE` | `/api/player` | 선수 삭제 |

### Match

| Method | Path | 설명 |
|---|---|---|
| `GET` | `/api/match?team={id}` | 경기 목록 조회 |
| `POST` | `/api/match` | 경기 생성 |
| `PUT` | `/api/match` | 경기 수정 |
| `DELETE` | `/api/match` | 경기 삭제 |

### Quarter

| Method | Path | 설명 |
|---|---|---|
| `GET` | `/api/quarter?match={id}` | 쿼터 목록 조회 |
| `POST` | `/api/quarter` | 쿼터 생성 |
| `PUT` | `/api/quarter/awaygoals` | 원정 골 수 부분 수정 |
| `PUT` | `/api/quarter` | 쿼터 전체 수정 |
| `DELETE` | `/api/quarter` | 쿼터 삭제 |

### Record

| Method | Path | 설명 |
|---|---|---|
| `GET` | `/api/record?quarter={id}` | 기록 목록 조회 |
| `GET` | `/api/record?player={id}` | 선수별 기록 조회 |
| `POST` | `/api/record` | 기록 생성 |
| `PUT` | `/api/record/stats` | 골/어시스트/출전시간 부분 수정 |
| `PUT` | `/api/record` | 기록 전체 수정 |
| `DELETE` | `/api/record` | 기록 삭제 |

> **부분 업데이트 엔드포인트** (`/awaygoals`, `/stats`): 전체 UPDATE 시 다른 필드가 0으로 덮어씌워지는 문제를 방지하기 위해 지정 컬럼만 업데이트합니다.

---

## 응답 형식

```json
// 목록 조회
{ "code": "success", "items": [...] }

// 단건 조회 / 생성 / 수정
{ "code": "success", "item": { ... } }

// 삭제 / 기타
{ "code": "success" }

// 오류
{ "code": "error", "message": "..." }
```

---

## 시작하기

### 요구사항

- Go 1.26+
- MySQL 8.0+

### 1. 환경 설정

```bash
cp .env.yml.example .env.yml
# .env.yml 에서 DB 접속 정보 수정
```

### 2. 데이터베이스 초기화

```bash
mysql -u root -p fotstat < go-basic.sql
mysql -u root -p fotstat < migration_add_quarter_duration.sql
mysql -u root -p fotstat < migration_add_quarter_awaygoals.sql
```

### 3. 실행

```bash
go mod tidy
make run
```

서버는 기본 포트 **8007**에서 실행됩니다.

---

## Docker 배포

```bash
# 백그라운드 실행
docker-compose up -d

# 로그 확인
docker-compose logs -f backend
```

---

## Postman 테스트

`api_postman_collection.json`을 Postman에서 Import 후:

1. `POST /api/user` — 회원가입
2. `GET /api/jwt` — 로그인 → 응답 토큰이 `{{jwt_token}}`에 자동 저장
3. 이후 모든 요청에 Bearer Token 자동 적용

---

## 빌드

```bash
make server     # 로컬 바이너리 빌드
make linux      # Linux 배포용 바이너리
make docker     # Docker 이미지 빌드
```
