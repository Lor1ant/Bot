// SPDX-License-Identifier: PolyForm-Noncommercial-1.0.0

// Package csqtt — минимальный клиент веб-панели сервера CSQTT. Панель
// авторизует логином/паролем и отдаёт куку сессии на сутки, дальше все
// операции идут через /api/* с этой кукой. Пароль клиенту генерирует сама
// панель (POST /api/clients), готовой ссылки csqtt:// панель не отдаёт —
// её собирает BuildLink ровно по той же схеме, что и браузер панели.
package csqtt

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ClientInfo — запись клиента в панели (GET /api/clients). Поля, которые
// выдаче не нужны (трафик, устройства), не разбираем: JSON панель отдаёт
// с лишними ключами, и они молча игнорируются.
type ClientInfo struct {
	Password  string `json:"password"`
	Name      string `json:"name"`
	Expires   int64  `json:"expires"`
	Active    bool   `json:"active"`
	VKHashes  string `json:"vk_hashes"`
	DTLSPort  int    `json:"dtls_port"`
	WGPort    int    `json:"wg_port"`
	LocalPort int    `json:"local_port"`
}

// CreateRequest — создание клиента (POST /api/clients). Hash и порты обычно
// копируются из клиента-шаблона, чтобы новый доступ работал в том же
// режиме (WDTT/CSQTT), что и существующие.
type CreateRequest struct {
	Name      string `json:"name"`
	Days      int    `json:"days"`
	Hash      string `json:"hash"`
	DTLSPort  int    `json:"dtls_port,omitempty"`
	WGPort    int    `json:"wg_port,omitempty"`
	LocalPort int    `json:"local_port,omitempty"`
}

// CreateResult — ответ панели на создание: пароль сгенерирован ею же.
type CreateResult struct {
	Password  string `json:"password"`
	Expires   int64  `json:"expires"`
	DTLSPort  int    `json:"dtls_port"`
	WGPort    int    `json:"wg_port"`
	LocalPort int    `json:"local_port"`
	VKHashes  string `json:"vk_hashes"`
}

type Client struct {
	base string
	user string
	pass string
	http *http.Client

	// mu защищает cookie: выдачу могут параллельно дёргать несколько
	// пользователей, а логин в панель строго по одному.
	mu     sync.Mutex
	cookie string
}

func New(baseURL, user, pass string) *Client {
	// Панель обычно висит на IP:порт с самоподписанным сертификатом
	// (как https://94.183.238.47:46002/) — системные CA его не примут.
	// Это соединение «бот → своя панель», а не браузер пользователя,
	// поэтому проверку цепочки отключаем; логин/пароль всё равно идут по TLS.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	return &Client{
		base: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		user: strings.TrimSpace(user),
		pass: pass,
		http: &http.Client{Timeout: 15 * time.Second, Transport: transport},
	}
}

// Login авторизуется в панели и запоминает куку сессии. Панель принимает
// логин и пароль открытым текстом (её caesar_decode пропускает значения без
// префикса «c1:» как есть) и отдаёт csqtt_session на 24 часа.
func (c *Client) Login(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	body, err := json.Marshal(map[string]string{"user": c.user, "pass": c.pass})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/login", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("/api/login: %s", resp.Status)
	}
	for _, ck := range resp.Cookies() {
		if ck.Name == "csqtt_session" && ck.Value != "" {
			c.cookie = ck.Value
			return nil
		}
	}
	return fmt.Errorf("/api/login: сессия не выдана")
}

// do выполняет запрос к панели; на 401 перелогинивается и повторяет его
// один раз — сессия живёт сутки, но панель или сервер могли перезапуститься.
func (c *Client) do(ctx context.Context, method, path string, body, out any, relogin bool) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.mu.Lock()
	cookie := c.cookie
	c.mu.Unlock()
	if cookie != "" {
		req.Header.Set("Cookie", "csqtt_session="+cookie)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized && relogin && cookie != "" {
		if err := c.Login(ctx); err != nil {
			return err
		}
		return c.do(ctx, method, path, body, out, false)
	}
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(payload))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("%s: %s", path, msg)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(payload, out)
}

func (c *Client) ListClients(ctx context.Context) ([]ClientInfo, error) {
	var out []ClientInfo
	if err := c.do(ctx, http.MethodGet, "/api/clients", nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateClient(ctx context.Context, req CreateRequest) (*CreateResult, error) {
	var out CreateResult
	if err := c.do(ctx, http.MethodPost, "/api/clients", req, &out, true); err != nil {
		return nil, err
	}
	if out.Password == "" {
		return nil, fmt.Errorf("/api/clients: панель не вернула пароль")
	}
	return &out, nil
}

// BuildLink собирает ссылку импорта csqtt:// по той же схеме, что и сама
// панель (buildCsqttLink в её веб-интерфейсе): host — только имя хоста,
// без порта; hashes перечисляются через «+».
func BuildLink(host, password string, peerPort int, rawHashes string) string {
	var sb strings.Builder
	sb.WriteString("csqtt://connect?v=2&host=")
	sb.WriteString(url.QueryEscape(host))
	sb.WriteString("&peer=")
	sb.WriteString(strconv.Itoa(peerPort))
	sb.WriteString("&password=")
	sb.WriteString(url.QueryEscape(password))
	var hashes []string
	for _, h := range strings.Split(rawHashes, ",") {
		if h = strings.TrimSpace(h); h != "" {
			hashes = append(hashes, url.QueryEscape(h))
		}
	}
	if len(hashes) > 0 {
		sb.WriteString("&hashes=")
		sb.WriteString(strings.Join(hashes, "+"))
	}
	return sb.String()
}
