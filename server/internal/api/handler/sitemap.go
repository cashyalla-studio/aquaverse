package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type SitemapHandler struct {
	fishSvc *service.FishService
}

func NewSitemapHandler(fishSvc *service.FishService) *SitemapHandler {
	return &SitemapHandler{fishSvc: fishSvc}
}

func (h *SitemapHandler) Sitemap(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	filter := domain.FishFilter{
		Locale: domain.LocaleENUS,
		Page:   1,
		Limit:  1000,
	}
	result, err := h.fishSvc.List(ctx, filter)
	if err != nil {
		result = nil
	}

	baseURL := "https://finara.app"

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>` + baseURL + `/</loc><changefreq>daily</changefreq><priority>1.0</priority></url>
  <url><loc>` + baseURL + `/fish</loc><changefreq>weekly</changefreq><priority>0.9</priority></url>
  <url><loc>` + baseURL + `/marketplace</loc><changefreq>daily</changefreq><priority>0.8</priority></url>
  <url><loc>` + baseURL + `/community</loc><changefreq>daily</changefreq><priority>0.7</priority></url>
  <url><loc>` + baseURL + `/species/reptile</loc><changefreq>weekly</changefreq><priority>0.8</priority></url>
  <url><loc>` + baseURL + `/species/amphibian</loc><changefreq>weekly</changefreq><priority>0.8</priority></url>
  <url><loc>` + baseURL + `/species/insect</loc><changefreq>weekly</changefreq><priority>0.8</priority></url>
`

	if result != nil {
		for _, f := range result.Items {
			xml += fmt.Sprintf(`  <url><loc>%s/fish/%d</loc><changefreq>monthly</changefreq><priority>0.6</priority><lastmod>%s</lastmod></url>
`, baseURL, f.ID, time.Now().Format("2006-01-02"))
		}
	}

	xml += `</urlset>`

	c.Response().Header().Set("Content-Type", "application/xml")
	return c.String(http.StatusOK, xml)
}
