package v1

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/majcek210/monsee/internal/service"
)

// cdataEscape prevents CDATA injection by splitting any ]]> sequence so it
// cannot prematurely close a <![CDATA[...]]> section.
func cdataEscape(s string) string {
	return strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>")
}

type RSSHandler struct {
	incidents *service.IncidentService
	settings  *service.SettingsService
}

func NewRSSHandler(incidents *service.IncidentService, settings *service.SettingsService) *RSSHandler {
	return &RSSHandler{incidents: incidents, settings: settings}
}

// GetFeed returns an RSS 2.0 feed of recent incidents.
func (h *RSSHandler) GetFeed(c fiber.Ctx) error {
	cfg, err := h.settings.Get(c.Context())
	if err != nil {
		return err
	}

	incs, err := h.incidents.List(c.Context(), "")
	if err != nil {
		return err
	}

	baseURL := fmt.Sprintf("%s://%s", c.Scheme(), c.Hostname())

	var sb strings.Builder
	for _, inc := range incs {
		pubDate := inc.CreatedAt.UTC().Format(time.RFC1123Z)
		fmt.Fprintf(&sb, `
    <item>
      <title><![CDATA[%s]]></title>
      <link>%s/incidents/%s</link>
      <guid>%s/incidents/%s</guid>
      <pubDate>%s</pubDate>
      <description><![CDATA[Status: %s | Severity: %s]]></description>
    </item>`, cdataEscape(inc.Title), baseURL, inc.ID, baseURL, inc.ID, pubDate, inc.Status, inc.Severity)
	}
	items := sb.String()

	feed := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title><![CDATA[%s — Incidents]]></title>
    <link>%s</link>
    <description>Recent incidents for %s</description>
    <lastBuildDate>%s</lastBuildDate>%s
  </channel>
</rss>`, cdataEscape(cfg.SiteTitle), baseURL, html.EscapeString(cfg.SiteTitle), time.Now().UTC().Format(time.RFC1123Z), items)

	c.Set("Content-Type", "application/rss+xml; charset=utf-8")
	return c.SendString(feed)
}
