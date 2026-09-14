package app

import (
	"net/url"
	"strings"
)

func rewriteSubLink(src, override string) string {
	override = strings.TrimSpace(override)
	if override == "" || src == "" {
		return src
	}
	host := extractHost(override)
	if host == "" {
		return src
	}
	u, err := url.Parse(src)
	if err != nil || u.Host == "" {
		return src
	}
	if u.Host == host {
		return src
	}
	u.Host = host
	return u.String()
}

func extractHost(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "://") {

		if i := strings.Index(s, "/"); i >= 0 {
			s = s[:i]
		}
		return s
	}
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	return u.Host
}

func (a *App) subOverride() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.botCfg == nil {
		return ""
	}
	return a.botCfg.SubscriptionDomain
}

func (a *App) rewriteSub(src string) string {
	return rewriteSubLink(src, a.subOverride())
}
