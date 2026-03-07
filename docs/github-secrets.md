# GitHub Secrets 설정 가이드

## App Store (iOS) 배포 Secrets

| Secret 이름 | 설명 | 획득 방법 |
|---|---|---|
| `MATCH_PASSWORD` | Fastlane Match 암호화 비밀번호 | 팀 내 공유 |
| `MATCH_GIT_BASIC_AUTHORIZATION` | Match 인증서 저장소 접근 토큰 | `echo -n "user:token" \| base64` |
| `ASC_KEY_ID` | App Store Connect API 키 ID | App Store Connect → 사용자 및 접근 → 키 |
| `ASC_ISSUER_ID` | App Store Connect 발급자 ID | 위와 동일 페이지 |
| `ASC_PRIVATE_KEY` | App Store Connect API 개인키 (.p8) | 키 다운로드 후 내용 전체 복사 |

## Play Store (Android) 배포 Secrets

| Secret 이름 | 설명 | 획득 방법 |
|---|---|---|
| `ANDROID_KEYSTORE_BASE64` | 업로드 키스토어 Base64 | `base64 -i upload-keystore.jks` |
| `ANDROID_KEY_ALIAS` | 키스토어 별칭 | keytool 생성 시 지정한 값 |
| `ANDROID_KEY_PASSWORD` | 키 비밀번호 | keytool 생성 시 설정한 값 |
| `ANDROID_STORE_PASSWORD` | 키스토어 비밀번호 | keytool 생성 시 설정한 값 |
| `GOOGLE_PLAY_JSON_KEY` | Google Play API 서비스 계정 JSON | Google Play Console → API 액세스 |

## 공통 Secrets

| Secret 이름 | 설명 |
|---|---|
| `PROD_API_URL` | 프로덕션 API URL (예: https://api.finara.app) |

## 초기 설정 체크리스트

- [ ] Apple Developer 계정에서 App ID `com.aquaverse.app` 등록
- [ ] App Store Connect에서 앱 등록
- [ ] Fastlane Match용 인증서 저장소 생성 (private repo)
- [ ] `fastlane match init` 실행 후 `fastlane match appstore` 실행
- [ ] Google Play Console에서 앱 등록
- [ ] Google Play API 서비스 계정 생성 및 권한 부여
- [ ] Android 업로드 키스토어 생성:
  ```bash
  keytool -genkey -v -keystore upload-keystore.jks \
    -keyalg RSA -keysize 2048 -validity 10000 \
    -alias upload
  ```
