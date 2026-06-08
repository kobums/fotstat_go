# secrets/

Sign in with Apple 비공개 키(.p8)를 여기에 둡니다.

1. Apple Developer → Certificates, Identifiers & Profiles → **Keys** 에서
   "Sign in with Apple" 키를 만들고 `AuthKey_XXXXXXXXXX.p8` 를 다운로드합니다.
2. 이 디렉터리에 두고 파일명을 `AuthKey.p8` 로 바꾸거나,
   `.env.yml` 의 `apple.privateKeyPath` / 환경변수 `APPLE_PRIVATE_KEY_PATH` 를
   실제 파일명으로 맞춥니다.

⚠️ `.p8` 파일은 **절대 커밋하지 마세요** (`.gitignore` 로 차단되어 있습니다).
   키는 한 번만 다운로드 가능하니 안전한 곳에 백업해두세요.
