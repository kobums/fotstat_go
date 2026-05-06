<div align="center">

# ⚽ Fotstat (Football Statistics API Server)

**Go (Fiber) 기반의 축구 경기 통계 및 기록 관리 REST API 서버**

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ACD7?style=for-the-badge&logo=go&logoColor=white)](https://gofiber.io/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=for-the-badge&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![JWT](https://img.shields.io/badge/JWT-Auth-black?style=for-the-badge&logo=jsonwebtokens)](https://jwt.io/)

</div>

---

## 📖 소개

**Fotstat**은 축구 경기(Match), 팀(Team), 선수(Player), 쿼터(Quarter), 그리고 경기 기록(Record) 데이터를 효율적으로 관리하고 통계를 낼 수 있는 백엔드 API 서버입니다.  
안전한 JWT 기반 인증과 `bcrypt` 암호화를 지원하며, Go언어의 경량 웹 프레임워크인 Fiber를 활용하여 빠르고 확장성 있게 구축되었습니다.

---

## ✨ 주요 기능

| 기능 | 설명 |
|:---|:---|
| 🔒 **인증 (Auth)** | JWT(JSON Web Token) 및 bcrypt를 활용한 안전한 로그인 및 회원가입 (`/api/user`, `/api/jwt`) |
| ⚽ **경기 통계 도메인** | `User`, `Team`, `Player`, `Match`, `Quarter`, `Record` 도메인에 대한 완전한 CRUD REST API 제공 |
| 🛡️ **API 보호** | 인증된 사용자(Bearer Token)만 주요 데이터에 접근할 수 있도록 미들웨어 라우팅 분리 |
| 🗂️ **파일 업로드** | 경기/선수 관련 이미지 등 파일 업로드 기능 (`/api/upload`) |
| 🧪 **API 테스트 세팅** | `api_postman_collection.json` 제공 (토큰 자동 발급 및 전역 환경 설정 스크립트 포함) |

---

## 🏗️ 프로젝트 구조

```
fotstat/
├── main.go                 # 앱 엔트리포인트
├── services/
│   └── http.go             # Fiber HTTP 서버 설정 (CORS, 정적 파일 제공 등)
├── router/
│   ├── router.go           # 전체 라우터 초기화 (apiGroup 설정)
│   ├── auth.go             # JWT 인증 미들웨어 및 권한 검증 로직
│   └── routers/            # 각 도메인별 자동/수동 생성된 라우트 (user, match, team 등)
├── controllers/
│   ├── api/                # API 전용 컨트롤러 (파일 업로드 등)
│   └── rest/               # 각 모델별 비즈니스 로직 및 RESTful 컨트롤러
├── models/
│   ├── db.go               # MySQL 데이터베이스 커넥션 및 쿼리 매니저
│   └── {domain}.go         # user, team, match, player, quarter, record 모델 정의
├── global/
│   └── jwt/                # JWT 토큰 생성 및 bcrypt 패스워드 검증 유틸리티
├── api_postman_collection.json # 전체 API 테스트용 Postman 컬렉션
├── docker-compose.yml      # 백엔드 서버 Docker Compose 설정
└── .env.yml                # 개발/운영 환경 설정 파일
```

---

## 🚀 시작하기

### 1. 프로젝트 초기 설정
```bash
# 패키지 다운로드
go mod tidy

# 환경 설정 파일 준비 (.env.yml)
# DB 호스트, 포트, 비밀번호 등 환경에 맞게 수정
cp .env.yml.example .env.yml
```

### 2. 데이터베이스 구성
MySQL 서버가 실행 중이어야 합니다 (`Err: 61` 주의). `.env.yml`에 지정된 데이터베이스 정보에 맞게 스키마를 구성합니다.

### 3. 서버 실행
```bash
# 개발 모드 실행
go run main.go
# 또는 Make 명령어가 구성된 경우: make run
```
서버는 기본적으로 `http://localhost:8007`에서 실행됩니다.

---

## 🧪 API 테스트 (Postman)

프로젝트 루트에 포함된 **`api_postman_collection.json`** 파일을 Postman에 Import하여 사용하세요.

1. **회원가입**: `Auth > Create User` (`POST /api/user`)
2. **로그인**: `Auth > Login` (`GET /api/jwt`)
   - 로그인에 성공하면, 응답으로 받은 토큰이 자동으로 Postman 전역 변수 `{{jwt_token}}`에 저장됩니다.
3. **API 요청**: 이후 모든 `Match`, `Team`, `Player` 관련 API는 컬렉션의 Auth(Bearer Token)를 통해 자동으로 인증되어 원활한 테스트가 가능합니다.

---

## 🐳 Docker 지원

```bash
# 도커 이미지 빌드
make docker

# 도커 컨테이너 백그라운드 실행
docker-compose up -d
```
