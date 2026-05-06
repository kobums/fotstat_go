<div align="center">

# 🚀 fotstat (Golang Backend Boilerplate)

**Fiber 기반의 Go REST API 서버 기본 템플릿**

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ACD7?style=for-the-badge&logo=go&logoColor=white)](https://gofiber.io/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=for-the-badge&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)

</div>

---

## 📖 소개

**fotstat**은 새로운 Go 백엔드 프로젝트를 빠르게 시작하기 위한 기본 템플릿(Boilerplate)입니다.
Fiber 웹 프레임워크와 자동화된 코드 생성 도구를 활용하여 MVC 아키텍처 기반의 확장성 있는 서버를 구축할 수 있습니다.

---

## ✨ 주요 기능

| 기능 | 설명 |
|:---|:---|
| ⚙️ **코드 자동 생성** | `model.json` 정의를 기반으로 DB 모델 및 라우터 자동 생성 |
| 🗂️ **기본 컨트롤러** | 파일 업로드 기능 및 공통 컨트롤러 로직 포함 |
| 🔒 **설정 관리** | `.env.yml` 및 환경 변수를 통한 설정 (개발/운영 분리) |
| 🐳 **Docker 지원** | Multi-stage 빌드를 통한 경량 컨테이너 배포 지원 |

---

## 🏗️ 프로젝트 구조

```
fotstat/
├── main.go                 # 앱 엔트리포인트
├── services/
│   └── http.go             # Fiber HTTP 서버 설정 (CORS, TLS, 압축 등)
├── router/
│   ├── router.go           # 라우터 초기화
│   └── routers/            # 자동 생성된 라우트 및 기본 라우트 (upload.go)
├── controllers/
│   ├── controllers.go      # 공통 컨트롤러 베이스 로직
│   ├── api/                # API 전용 컨트롤러 (파일 업로드)
│   └── rest/               # 자동 생성된 RESTful 컨트롤러 디렉토리
├── models/
│   ├── db.go               # 데이터베이스 연결 및 설정
│   └── cache.go            # 인메모리 캐시 기능
├── global/                 # 전역 설정 및 유틸리티 (로깅, 시간, 이미지 처리 등)
├── dockerfile              # Multi-stage Docker 빌드
├── docker-compose.yml      # Docker Compose 설정
└── Makefile                # 빌드 및 실행 명령어
```

---

## 🚀 시작하기

### 1. 프로젝트 초기 설정
```bash
# 의존성 패키지 다운로드
go mod tidy

# 환경 설정 파일 복사 및 수정
cp .env.yml.example .env.yml

# 데이터베이스 스키마 생성
mysql -u db_user -p example_db < fotstat.sql
```

### 2. 코드 생성 및 실행
```bash
# 자동 생성 도구 실행 (도구 별도 필요)
# buildtool-model / buildtool-router 등

# 서버 빌드 및 실행
make run
```
서버는 기본적으로 `http://localhost:8007`에서 실행됩니다.

---

## 📝 환경 설정 (`.env.yml`)

```yaml
develop:
  database:
    type: mysql
    host: localhost
    port: 3306
    name: example_db
    user: db_user
    password: db_password
  port: 8007
  cors: [http://localhost:9007]
  documentRoot: ./webdata
  path: ./webdata
```

---

## 🐳 Docker 배포

```bash
# Docker 이미지 빌드
make docker

# Docker 컨테이너 실행
make dockerrun
```
