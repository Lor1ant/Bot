package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// isoAlpha2 is the full ISO 3166-1 alpha-2 country-code set. Flags for all of
// these are fetched ONCE when the cabinet/mini-app comes up and then served by
// the bot itself. Rationale (privacy F1): the cabinet must never load flags
// from a third party (e.g. flagcdn) at runtime — that would leak every
// visitor's IP off-domain and add an external fingerprintable dependency.
const isoAlpha2 = "ad ae af ag ai al am ao aq ar as at au aw ax az ba bb bd be bf bg bh bi bj bl bm bn bo bq br bs bt bv bw by bz ca cc cd cf cg ch ci ck cl cm cn co cr cu cv cw cx cy cz de dj dk dm do dz ec ee eg eh er es et fi fj fk fm fo fr ga gb gd ge gf gg gh gi gl gm gn gp gq gr gs gt gu gw gy hk hm hn hr ht hu id ie il im in io iq ir is it je jm jo jp ke kg kh ki km kn kp kr kw ky kz la lb lc li lk lr ls lt lu lv ly ma mc md me mf mg mh mk ml mm mn mo mp mq mr ms mt mu mv mw mx my mz na nc ne nf ng ni nl no np nr nu nz om pa pe pf pg ph pk pl pm pn pr ps pt pw py qa re ro rs ru rw sa sb sc sd se sg sh si sj sk sl sm sn so sr ss st sv sx sy sz tc td tf tg th tj tk tl tm tn to tr tt tv tw tz ua ug um us uy uz va vc ve vg vi vn vu wf ws ye yt za zm zw"

// flagCDN is the upstream the flag SVGs are pulled from, once, at startup.
const flagCDN = "https://flagcdn.com/"

var errFlagFetch = errors.New("flag fetch failed")

func (a *App) flagDir() string { return filepath.Join(a.cfg.DataDir, "flags") }

// Flag returns the cached SVG bytes for a 2-letter ISO country code.
func (a *App) Flag(code string) ([]byte, bool) {
	code = strings.ToLower(strings.TrimSpace(code))
	if len(code) != 2 || code[0] < 'a' || code[0] > 'z' || code[1] < 'a' || code[1] > 'z' {
		return nil, false
	}
	a.flagMu.RLock()
	b, ok := a.flags[code]
	a.flagMu.RUnlock()
	if ok {
		return b, true
	}
	// not in memory yet — try the on-disk cache
	if disk, err := os.ReadFile(filepath.Join(a.flagDir(), code+".svg")); err == nil && len(disk) > 0 {
		a.flagMu.Lock()
		if a.flags == nil {
			a.flags = map[string][]byte{}
		}
		a.flags[code] = disk
		a.flagMu.Unlock()
		return disk, true
	}
	return nil, false
}

// CabinetFlag exposes Flag to the web layer (MiniProvider).
func (a *App) CabinetFlag(code string) ([]byte, bool) { return a.Flag(code) }

// ensureFlagsAsync downloads the full flag set once (idempotent), in the
// background, so the cabinet/mini-app serve flags from this server instead of a
// third-party CDN. Cached to {DataDir}/flags so restarts skip the network.
func (a *App) ensureFlagsAsync(ctx context.Context) {
	a.flagMu.Lock()
	if a.flagsStarted {
		a.flagMu.Unlock()
		return
	}
	a.flagsStarted = true
	if a.flags == nil {
		a.flags = map[string][]byte{}
	}
	a.flagMu.Unlock()
	go a.loadFlags(ctx)
}

func (a *App) loadFlags(ctx context.Context) {
	dir := a.flagDir()
	_ = os.MkdirAll(dir, 0o755)
	codes := strings.Fields(isoAlpha2)
	client := &http.Client{Timeout: 10 * time.Second}
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	var cntMu sync.Mutex
	var cached, fetched int
	for _, code := range codes {
		p := filepath.Join(dir, code+".svg")
		if disk, err := os.ReadFile(p); err == nil && len(disk) > 0 {
			a.flagMu.Lock()
			a.flags[code] = disk
			a.flagMu.Unlock()
			cntMu.Lock()
			cached++
			cntMu.Unlock()
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(code, p string) {
			defer wg.Done()
			defer func() { <-sem }()
			b, err := fetchFlag(ctx, client, code)
			if err != nil || len(b) == 0 {
				return
			}
			_ = os.WriteFile(p, b, 0o644)
			a.flagMu.Lock()
			a.flags[code] = b
			a.flagMu.Unlock()
			cntMu.Lock()
			fetched++
			cntMu.Unlock()
		}(code, p)
	}
	wg.Wait()
	a.log.Info("cabinet flags ready", "cached", cached, "downloaded", fetched, "total", len(codes))
}

func fetchFlag(ctx context.Context, client *http.Client, code string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, flagCDN+code+".svg", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errFlagFetch
	}
	return io.ReadAll(io.LimitReader(resp.Body, 256*1024))
}
