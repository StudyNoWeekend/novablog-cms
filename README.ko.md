<div align="center">

<img src="docs/assets/novablog-logo.png" alt="NovaBlog CMS" width="280" />

# NovaBlog CMS

**취향과 직업에 맞춰 모든 것을 미리 준비해 주는 오픈소스 블로그 시스템 —— 콘텐츠는 CMS가, 외형은 테마가, 글쓰기는 AI가**

<p>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Redis-7+-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis" />
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
</p>

[공식 사이트](http://www.novablog.cn) · [무료 가입](http://www.novablog.cn/register) · [작가 블로그](http://czf.ditancafebar.cn) · [테마 저장소](https://github.com/StudyNoWeekend/novablog-web) · [Docker 이미지](https://github.com/StudyNoWeekend/novablog-cms/pkgs/container/novablog-cms) · [Skill 문서](https://github.com/StudyNoWeekend/novablog-cms/blob/main/frontend/public/docs/novablog-api.md)

</div>

<div align="center">

**언어**: [简体中文](README.md) · [繁體中文](README.zh-TW.md) · [English](README.en.md) · [日本語](README.ja.md) · 한국어

</div>

---

## 🎯 세 단계로 나만의 블로그 갖기

코드도, 서버 지식도 필요하지 않습니다:

1. **블로그 개설** —— [공식 호스팅 신청](http://www.novablog.cn/hosting#apply)에서 신청하면 배포 없이 바로 사용할 수 있습니다. 완전한 직접 관리를 원하면 아래 [배포](#-배포-두-가지-방법) 섹션을 따라 Docker로 직접 구축하세요;
2. **초기 설정 마법사** —— 계정 생성 → **크리에이터 방향 선택** → 스토리지 설정 → 기본 테마가 자동 설치됩니다. 클릭 몇 번이면 끝납니다;
3. **창작 시작** —— 글쓰기, 사진 업로드, 레시피 공개, 여행 기록… 선택한 방향에 따라 백오피스에 미리 활성화되는 모듈이 결정되며, 언제든 조정할 수 있습니다.

초기 설정이 끝나면 바로 접속 가능한 완전한 블로그가 완성됩니다: CMS가 블로그 페이지를 직접 호스팅하고, 방문자에게는 선택한 테마가 보이며, 콘텐츠 데이터는 그대로 유지됩니다.

### 취향 / 직업에 맞춘 커스터마이즈: 크리에이터 방향

초기 설정 마법사 2단계에서 **17장의 "크리에이터 방향" 카드**(중복 선택 가능)가 표시됩니다. 마음에 드는 것을 선택하면 해당 모듈이 자동으로 활성화되고, 하나도 선택하지 않으면 전부 활성화됩니다. 로그인 후 백오피스 상단에서 언제든 변경할 수 있습니다.

| 그룹 | 크리에이터 방향 |
|------|----------|
| 기술 · 디지털 | 기술 개발자 · 디지털 리뷰 |
| 사진 · 사운드 | 사진 애호가 · 영상 크리에이터 · 뮤지션 · 여행 브이로거 |
| 여행 · 라이프 | 여행 블로거 · 미식 애호가 · 피트니스 · 패션 코디 · 반려동물 일상 |
| 취미 · 컬렉션 | 독서 · 게이머 · 디자이너/일러스트레이터 · 핸드메이드 DIY |
| 특별 | 라이프스타일 · 올라운드 크리에이터(모든 모듈 한 번에 활성화) |

---

## 🧩 모듈 한눈에 보기

백오피스 모듈은 "시스템 설정 → 모듈 관리"에서 언제든 켜고 끌 수 있으며, 끄면 블로그와 백오피스 메뉴 양쪽에서 함께 숨겨집니다(게시글 모듈은 항상 켜져 있고, 미디어 라이브러리는 기반 인프라로 항상 사용 가능).

### ✍️ 콘텐츠 제작(취향에 맞게 선택)

| 모듈 | 제공 기능 |
|------|----------|
| 📝 게시글 관리 | Markdown / 리치 텍스트 듀얼 에디터, 카테고리와 태그, 초안/게시/게시 중단, SEO(슬러그·요약·커버·키워드 커스텀), 조회 통계, 공식 커뮤니티 원클릭 발행 |
| 📸 사진 포트폴리오 | 테마별로 정리한 사진집, EXIF 촬영 정보 자동 추출, 다양한 커버 스타일과 워터마크 보호 |
| 🗺️ 여행 가이드 | 구조화된 다일정 편집기, 아맵 / 구글 지도 위치 선택과 상세 페이지 지도 표시, 좋아요와 인기 추천 |
| 🍳 요리 레시피 | 식재료와 조리 단계 관리, 공식 커뮤니티 원클릭 발행 |
| 📖 독서 서재 | 책 수납, 별점과 서평, Douban 도서 정보 연동 지원 |
| 🎮 게임 라이브러리 | 게임 컬렉션, 플레이 시간과 한줄평, Bangumi 항목 연동 지원 |
| 💪 피트니스 | 운동 기록 관리와 블로그 노출 |
| 🎵 음악 재생목록 | 곡을 업로드해 재생목록 구성, 빌리빌리 링크 붙여넣기로 자동 파싱 재생, NetEase Cloud / QQ 음악 / Spotify 공개 재생목록 가져오기 |
| 🎥 영상 작품 | 영상과 다대다 앨범 관리, 빌리빌리 링크 파싱 임베드 |

### 👤 소개(개인 명함 만들기)

| 모듈 | 제공 기능 |
|------|----------|
| 프로필 | 닉네임·아바타·소개, 블로그 제목/설명/배경, 20+ 소셜 플랫폼 링크(브랜드 컬러 아이콘 내장) |
| 개인 장비 | 디지털/카메라 장비 목록, 사진 업로드 시 촬영 정보 자동 인식 |
| 프로젝트 경력 | 온라인 이력서형 프로젝트 경력: 분야, 역할, 기술 스택, 기간 |
| 기술 스택 | 스킬 태그 관리, 블로그에서는 공식 브랜드 아이콘으로 표시 |
| 오픈소스 작품 | GitHub 오픈소스 프로젝트 쇼케이스 카드, 공식 커뮤니티 원클릭 발행 |

### 🧰 인프라(기본 내장)

| 모듈 | 제공 기능 |
|------|----------|
| 대시보드 | 방문자/콘텐츠 지표, 모듈별 콘텐츠 카드, ECharts 추이 차트 |
| 댓글 관리 | 게시글·여행 가이드 등 여러 곳의 방문자 댓글 통합, 3단계 검토, 대댓글, IP 블랙리스트 |
| 미디어 라이브러리 | 이미지/영상/오디오 업로드(대용량 파일 분할), 폴더 관리, 크기 프리셋 워크벤치, 외부 이미지 가져오기와 검색 |
| 테마 마켓 | 공식 테마 탐색, 미리보기, 원클릭 설치 활성화(아래 테마 섹션 참조) |
| 시스템 설정 | 모듈 스위치, 오브젝트 스토리지(로컬/MinIO/알리바바 OSS/텐센트 COS 핫 스위치와 마이그레이션), CORS, 지도 Key, IP 블랙리스트, 커뮤니티 발행 |
| 접속 통계 | 블로그 방문자 IP 통계, ip2region 오프라인 지역 분석, 외부 서비스 불필요 |
| API 문서 / MCP 서비스 | 인증 불필요 공개 API 온라인 열람, Skill 문서 다운로드, MCP 키 관리(아래 AI 섹션 참조) |

---

## 🤖 AI에게 글쓰기 시키기: MCP + Skill

NovaBlog CMS에는 **MCP 서버**와 **Skill 형식 API 문서**가 내장되어 있어, Codex, ZCode, Claude Code, Cursor 같은 AI 에이전트가 여러분의 블로그에서 직접 콘텐츠를 만들 수 있습니다——말로 지시하면 AI가 실행합니다.

### 방법 1: MCP 연동(권장: 쓰기·게시·관리 모두 가능)

백오피스 "AI → MCP 서비스"에서 3단계로 연동합니다:

1. **키 발급**: 셀프서비스로 MCP API Key 발급(`nbt_mcp_` 접두사, 만료 기간 설정 가능, 레이트 리밋 내장);
2. **설정 붙여넣기**: 페이지가 자동 생성한 JSON을 복사해 AI 클라이언트의 MCP 설정에 붙여넣습니다:

```json
{
  "mcpServers": {
    "novablog": {
      "type": "http",
      "url": "https://여러분의-블로그-도메인/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer nbt_mcp_여러분의-키"
      }
    }
  }
}
```

3. **창작 지시**: AI에게 "OO 주제의 글을 써서 게시해 줘"라고 말하면 됩니다.

MCP 엔드포인트는 Streamable HTTP(`/api/v1/mcp`)이며, 현재 9개의 도구를 제공합니다:

| 도구 | 기능 |
|------|------|
| `create_article` | 글쓰기: 제목, Markdown 본문, 요약, 태그/카테고리(자동 생성), 커버, 상태 |
| `update_article` | 게시글 수정 |
| `set_article_status` | 게시 / 초안 전환 / 게시 중단 |
| `delete_article` | 게시글 삭제(실수 방지를 위해 명시적 confirm 필요) |
| `list_articles` / `get_article` | 게시글 목록과 상세 조회(키워드, 페이징) |
| `upload_image` | 이미지 업로드: http 링크 자동 전송 또는 base64, 참조 가능한 URL 반환 |
| `list_categories` / `list_tags` | 카테고리와 태그 조회 |

### 방법 2: Skill 문서(AI에게 모든 공개 API 익히기)

백오피스 "AI → API 문서" 페이지에서 **SKILL.md 형식의 API 레퍼런스**([novablog-api.md](frontend/public/docs/novablog-api.md))를 다운로드해 Claude Code / Cursor / ZCode 같은 AI 어시스턴트의 워크스페이스에 넣으면, 인증 불필요한 모든 공개 API의 경로·파라미터·응답 예시를 익혀 테마 제작, 연동, 대량 데이터 추출을 도와줍니다.

---

## 🏠 공식 생태계: NovaBlog 공식 사이트와 CMS의 관계

한 문장으로 말하면: **CMS는 여러분 자신의 블로그 핵심이고, 공식 NovaBlog 사이트는 여러분의 콘텐츠 커뮤니티입니다.**

- **NovaBlog CMS(본 프로젝트)**: 셀프호스팅 오픈소스 블로그 시스템. 콘텐츠와 외형을 관리하고 블로그 페이지를 직접 호스팅;
- **NovaBlog 공식 사이트**: 커뮤니티 포털——공식 계정 체계, 테마 마켓, 콘텐츠 커뮤니티. CMS의 테마 마켓과 커뮤니티 발행은 모두 공식 사이트에 연결;
- **공식 주소 통합 관리**: 출하 시 내장, 백오피스 "시스템 설정 → 커뮤니티 발행"에서 확인·변경 가능. 테마 마켓과 커뮤니티 발행이 같은 진입점을 공유하며 서버가 직접 연결합니다.

| 진입점 | 주소 | 용도 |
|------|------|------|
| 공식 사이트 | <http://www.novablog.cn> | 커뮤니티 포털 · 테마 마켓 |
| 무료 가입 | <http://www.novablog.cn/register> | 공식 계정 개설: 마켓 로그인, 호스팅 배포, 발행 키 |
| 공식 API | <http://api.novablog.cn> | 커뮤니티 발행과 테마 마켓의 통합 서비스 진입점 |

> 백오피스 "시스템 설정 → 커뮤니티 발행"의 공식 주소는 실제 설정을 따릅니다.

**커뮤니티 발행**: 게시글, 여행 가이드, 요리 레시피, 사진 포트폴리오, 오픈소스 작품 5종 콘텐츠를 원클릭으로 공식 커뮤니티에 동기화해 노출을 얻을 수 있습니다——생성/업데이트 후 자동 비동기 발행, 삭제 시 자동 게시 중단, 변경 없으면 자동 건너뜀, 실패 시 자동 재시도. 편집 페이지의 발행 버튼에서 수동 재발행과 동기화 상태 확인도 지원합니다.

---

## 🎨 테마: 옷 갈아입듯 외형 바꾸기

블로그 프런트엔드는 통째로 교체 가능한 "테마"입니다. 테마는 외형과 경험을, CMS는 콘텐츠와 호스팅을 담당합니다. 테마 전환은 포인터 하나만 업데이트하면 되어 초 단위로 적용되고, 재시작이 필요 없으며 콘텐츠는 그대로입니다.

### 방법 1: 공식 마켓 원클릭 설치(권장)

백오피스 "테마 마켓"에서 개성 넘치는 11개 이상의 공식 테마(기술 스타일, 프레임 라이프, 방랑 세계, 멜로디 노트 등)를 탐색·검색·별점·즐겨찾기할 수 있고, 선택만 하면:

- **활성화 전 미리보기**: 실제 렌더링 확인(`/preview`);
- **원클릭 설치**: 다운로드 → SHA256 검증 → 안전 압축 해제 검증 → `theme.json` 매니페스트 검증이 자동으로 진행;
- **버전 관리와 원클릭 롤백**. 테마는 `theme.json`으로 설정 항목을 선언하며 백오피스가 자동으로 시각화 설정 패널을 생성합니다. 코드를 건드리지 않고 설정을 바꿀 수 있습니다.

### 방법 2: 로컬 마운트로 자체 프런트엔드 사용

자체 프런트엔드를 쓰고 싶다면? `theme.json`과 `dist/`를 포함한 디렉터리를 CMS에 마운트만 하면 됩니다. 수정 없이 바로 연동: 베어메탈 배포는 `themes.frontend_dir` 설정, Docker 배포는 `./deploy.sh --frontend-dir <디렉터리>`.

### 🌍 테마는 오픈소스, 함께 만들어요

모든 공식 테마는 테마 저장소 [StudyNoWeekend/novablog-web](https://github.com/StudyNoWeekend/novablog-web)에서 오픈소스로 공개되어 있으며, **공식 + 커뮤니티 공동 유지보수** 오픈 모델을 채택합니다: 공식 팀이 공식 테마를 지속 제작·유지보수하고, 커뮤니티 작가는 규격에 맞춰 자체 테마를 개발해 Issue / PR을 제출하면 공식 마켓을 통해 모든 블로거에게 배포됩니다.

- 테마 패키지 규격: [novablog-web/docs/theme-spec.md](https://github.com/StudyNoWeekend/novablog-web/blob/main/docs/theme-spec.md)
- 패키징과 릴리스 가이드: [novablog-主题包规范与发布指南.md](novablog-主题包规范与发布指南.md)

---

## 🚀 배포: 두 가지 방법

### 방법 1: 공식 호스팅(배포 불필요, 초보자 추천)

[공식 사이트에서 무료 가입](http://www.novablog.cn/register)만 하면 호스팅 블로그가 개설됩니다. 서버도 Docker도 필요 없고, 가입 후 셀프호스팅과 완전히 동일한 초기 설정 마법사가 시작됩니다.

### 방법 2: Docker로 셀프호스팅

단일 이미지에 nginx 듀얼 엔트리, Go 백엔드, 관리자 프런트엔드가 내장되어 있습니다: `ghcr.io/studynoweekend/novablog-cms`(`linux/amd64` / `linux/arm64`). PostgreSQL 15+ 및 Redis 7+ 인스턴스는 직접 준비해야 합니다(스크립트가 연결성을 사전 점검):

```bash
cd deploy
./deploy.sh --name myblog                        # 대화형 질의로 나머지 설정 입력
./deploy.sh --name myblog --yes \
  --db-host 10.0.0.5 --db-password '***' \
  --redis-host 10.0.0.5 --redis-password '***'   # 비대화형 전체 기본값 배포
```

실행 후: 블로그 진입점 `http://<도메인>:80`, 백오피스 `http://<도메인>:8080/admin/`.

| 진입점 | 기본 포트 | 내용 |
|------|----------|------|
| 블로그 진입점 | `80` | 활성화된 테마 페이지(또는 로컬 마운트 프런트엔드) · `/preview` · `/api/v1` · `/files` |
| 백오피스 진입점 | `8080` | `/admin/` 관리자 백오피스 · `/api/v1` · `/files` |

자주 쓰는 운영 명령:

```bash
./deploy.sh --name myblog --status    # 컨테이너 상태 확인
./deploy.sh --name myblog --logs      # 앱 로그 추적
./deploy.sh --name myblog --down      # 중지 후 컨테이너 제거(마운트 데이터는 유지)
```

- `--name`은 필수이며, 같은 호스트에서 여러 블로그를 격리하는 키로 사용됩니다. 이름과 포트만 다르면 한 호스트에서 공존할 수 있습니다;
- 업그레이드는 `./deploy.sh --name myblog --version vX.Y.Z` 재실행, 마운트 데이터는 유지;
- 전체 파라미터, 외부 데이터베이스, 호스트 nginx 리버스 프록시 설명은 [deploy/README.md](deploy/README.md) 참조.

### 로컬 개발

```bash
# 1. 백엔드 실행
cd backend
cp config/config.example.yaml config/config.yaml   # postgres/redis 연결과 키 편집
make run                                            # :8111, 첫 실행 시 자동 마이그레이션

# 2. 관리자 프런트엔드 실행
cd ../frontend
pnpm install && pnpm dev                            # :5173, /api는 :8111로 프록시
```

브라우저에서 `http://localhost:5173`을 열면 초기 설정 마법사가 자동으로 시작됩니다.

---

## 🤝 기여

Issue와 Pull Request를 환영합니다:

1. 백엔드 커밋 전 `make fmt && make vet && make lint` 실행;
2. 프런트엔드 커밋 전 `pnpm lint` 실행;
3. API 추가 시 DTO와 API 문서를 함께 업데이트.

---

## 📜 라이선스

본 프로젝트는 [MIT](LICENSE) 라이선스로 공개되어 있습니다.

---

## 🙏 감사의 말

- [Gin](https://github.com/gin-gonic/gin) · [GORM](https://gorm.io) · [Vue.js](https://vuejs.org) · [Ant Design Vue](https://antdv.com) · [PostgreSQL](https://www.postgresql.org) · [Redis](https://redis.io)
- 그리고 이 프로젝트에 영감과 지원을 보내준 모든 커뮤니티 기여자분들

---

<p align="center">
  <sub>Made with ❤️ for independent bloggers</sub>
</p>
