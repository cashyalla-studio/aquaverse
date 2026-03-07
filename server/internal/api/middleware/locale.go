package middleware

import (
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/labstack/echo/v4"
)

// LocaleMiddleware - Accept-Language 또는 X-Locale 헤더에서 로케일 추출
// 게시판 접근 시 반드시 유효한 로케일이어야 함
func LocaleMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 우선순위: X-Locale 헤더 > Accept-Language > 기본값(en-US)
			locale := domain.Locale(c.Request().Header.Get("X-Locale"))
			if locale == "" || !locale.IsValid() {
				locale = parseAcceptLanguage(c.Request().Header.Get("Accept-Language"))
			}
			c.Set(ContextKeyLocale, locale)
			return next(c)
		}
	}
}

func parseAcceptLanguage(header string) domain.Locale {
	if header == "" {
		return domain.LocaleENUS
	}
	// 간단한 파싱: "ko-KR,ko;q=0.9,en-US;q=0.8" → "ko"
	// 지원 로케일과 매핑
	mapping := map[string]domain.Locale{
		"ko":    domain.LocaleKO,
		"en-us": domain.LocaleENUS,
		"en-gb": domain.LocaleENGB,
		"en-au": domain.LocaleENAU,
		"en":    domain.LocaleENUS,
		"ja":    domain.LocaleJA,
		"zh-cn": domain.LocaleZHCN,
		"zh-tw": domain.LocaleZHTW,
		"zh":    domain.LocaleZHCN,
		"de":    domain.LocaleDE,
		"fr-fr": domain.LocaleFRFR,
		"fr-ca": domain.LocaleFRCA,
		"fr":    domain.LocaleFRFR,
		"es":    domain.LocaleES,
		"pt":    domain.LocalePT,
		"ar":    domain.LocaleAR,
		"he":    domain.LocaleHE,
	}
	// 첫 번째 언어 태그만 파싱
	for i, ch := range header {
		if ch == ',' || ch == ';' {
			tag := header[:i]
			if loc, ok := mapping[tag]; ok {
				return loc
			}
		}
	}
	if loc, ok := mapping[header]; ok {
		return loc
	}
	return domain.LocaleENUS
}
