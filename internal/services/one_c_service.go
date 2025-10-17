package services

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"ElectronicQueue/internal/config"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

// OneCService handles communication with the 1C proxy service.
type OneCService struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewOneCService creates a new instance of OneCService.
// If a WireGuard config path is provided in the main .env config (1C_CONF),
// it creates a tunneled HTTP client. Otherwise, it uses a default client.
func NewOneCService(cfg *config.Config) (*OneCService, error) {
	var httpClient *http.Client
	var err error

	if cfg.OneCConf != "" {
		// Если указан конфиг WireGuard, создаем клиент, который работает через туннель
		httpClient, err = createWireGuardClient(cfg.OneCConf)
		if err != nil {
			return nil, fmt.Errorf("failed to create WireGuard client for 1C service: %w", err)
		}
	} else {
		// HTTP-клиент по умолчанию без туннеля
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &OneCService{
		baseURL:    cfg.OneCURL,
		apiKey:     cfg.OneCAPIKey,
		httpClient: httpClient,
	}, nil
}

// executeRequest handles making an authorized request to the 1C service.
func (s *OneCService) executeRequest(url string) (interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for 1C service: %w", err)
	}

	if s.apiKey != "" {
		// API ключ для 1С прокси-сервиса (1C.py)
		req.Header.Set("Authorization", "Basic "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to 1C service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("1C service returned non-OK status: %d - %s", resp.StatusCode, string(body))
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response from 1C service: %w", err)
	}

	return data, nil
}

// GetSchedule fetches a patient's appointment time by phone number from the 1C service.
func (s *OneCService) GetSchedule(phone string) (interface{}, error) {
	// Remove the leading '+' if it exists
	cleanedPhone := strings.TrimPrefix(phone, "+")
	reqURL := fmt.Sprintf("%s/getschedule?phone=%s", s.baseURL, cleanedPhone)
	return s.executeRequest(reqURL)
}

// GetDoctorSchedule fetches the general schedule from the 1C service.
func (s *OneCService) GetDoctorSchedule() (interface{}, error) {
	reqURL := s.baseURL + "/getdoctorschedule"
	return s.executeRequest(reqURL)
}

// --- WireGuard Client Creation ---

type wgInterface struct {
	PrivateKey string
	Address    []netip.Prefix
}

type wgPeer struct {
	PublicKey  string
	AllowedIPs []netip.Prefix
	Endpoint   string
}

// parseWgConf читает конфигурационный файл WireGuard.
func parseWgConf(path string) (*wgInterface, *wgPeer, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	iface := &wgInterface{}
	peer := &wgPeer{}
	inPeerSection := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if strings.EqualFold(line, "[Interface]") {
			inPeerSection = false
			continue
		}
		if strings.EqualFold(line, "[Peer]") {
			inPeerSection = true
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if !inPeerSection {
			switch strings.ToLower(key) {
			case "privatekey":
				iface.PrivateKey = value
			case "address":
				addrs := strings.Split(value, ",")
				for _, addrStr := range addrs {
					prefix, err := netip.ParsePrefix(strings.TrimSpace(addrStr))
					if err != nil {
						return nil, nil, fmt.Errorf("invalid Address %q: %w", addrStr, err)
					}
					iface.Address = append(iface.Address, prefix)
				}
			}
		} else {
			switch strings.ToLower(key) {
			case "publickey":
				peer.PublicKey = value
			case "allowedips":
				ips := strings.Split(value, ",")
				for _, ipStr := range ips {
					prefix, err := netip.ParsePrefix(strings.TrimSpace(ipStr))
					if err != nil {
						return nil, nil, fmt.Errorf("invalid AllowedIPs %q: %w", ipStr, err)
					}
					peer.AllowedIPs = append(peer.AllowedIPs, prefix)
				}
			case "endpoint":
				peer.Endpoint = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return iface, peer, nil
}

// base64ToHex конвертирует ключ WireGuard из base64 в hex для UAPI.
func base64ToHex(b64 string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	if len(decoded) != 32 {
		return "", fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(decoded))
	}
	return hex.EncodeToString(decoded), nil
}

// createWireGuardClient создает HTTP-клиент, который работает через туннель WireGuard.
func createWireGuardClient(confPath string) (*http.Client, error) {
	// 1. Парсинг конфигурации
	iface, peer, err := parseWgConf(confPath)
	if err != nil {
		return nil, fmt.Errorf("could not parse WireGuard config: %w", err)
	}

	var ifaceAddrs []netip.Addr
	for _, prefix := range iface.Address {
		ifaceAddrs = append(ifaceAddrs, prefix.Addr())
	}

	// 2. Создание TUN-устройства и сетевого стека
	// Используем публичный DNS-сервер в качестве запасного
	dns, _ := netip.ParseAddr("8.8.8.8")
	tun, tnet, err := netstack.CreateNetTUN(ifaceAddrs, []netip.Addr{dns}, 1420)
	if err != nil {
		return nil, fmt.Errorf("failed to create netstack TUN: %w", err)
	}

	// 3. Создание устройства WireGuard
	// LogLevelError используется, чтобы избежать спама в логах. Для отладки можно использовать LogLevelVerbose.
	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "(1C-wireguard) "))

	// 4. Сборка строки конфигурации UAPI
	privateKeyHex, err := base64ToHex(iface.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKeyHex, err := base64ToHex(peer.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid peer public key: %w", err)
	}

	var allowedIPsStr []string
	for _, ip := range peer.AllowedIPs {
		allowedIPsStr = append(allowedIPsStr, ip.String())
	}

	uapiConfig := fmt.Sprintf(
		"private_key=%s\n"+
			"public_key=%s\n"+
			"endpoint=%s\n"+
			"allowed_ip=%s\n"+
			"persistent_keepalive_interval=25",
		privateKeyHex,
		publicKeyHex,
		peer.Endpoint,
		strings.Join(allowedIPsStr, ","),
	)

	// 5. Применение конфигурации и запуск устройства
	if err := dev.IpcSet(uapiConfig); err != nil {
		return nil, fmt.Errorf("failed to set UAPI config: %w", err)
	}

	if err := dev.Up(); err != nil {
		return nil, fmt.Errorf("failed to bring up WireGuard device: %w", err)
	}

	// 6. Создание HTTP-клиента, который использует сетевой стек туннеля
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: tnet.DialContext,
		},
		Timeout: 20 * time.Second, // Увеличим таймаут для VPN-соединений
	}

	return client, nil
}
