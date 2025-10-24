package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/skip2/go-qrcode"
)

// TicketConfig содержит конфигурацию и данные для генерации талона
type TicketConfig struct {
	Width          int
	Height         int
	QRData         []byte
	FontPath       string
	BoldFontPath   string
	BackgroundPath string
	ServiceName    string
	TicketNumber   string
	DateTime       time.Time
	WaitingNumber  int
}

// ticketHTMLData содержит данные для HTML шаблона
type ticketHTMLData struct {
	Width            int
	Height           int
	BackgroundBase64 string
	QRCodeBase64     string
	ServiceName      string
	TicketNumber     string
	Date             string
	Time             string
	WaitingText      string
	IsColor          bool
	FontBase64       string
	BoldFontBase64   string
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        @font-face {
            font-family: 'CustomFont';
            src: url(data:font/truetype;base64,{{.FontBase64}}) format('truetype');
            font-weight: normal;
        }
        @font-face {
            font-family: 'CustomFont';
            src: url(data:font/truetype;base64,{{.BoldFontBase64}}) format('truetype');
            font-weight: bold;
        }
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'CustomFont', Arial, sans-serif;
            width: {{.Width}}px;
            height: {{.Height}}px;
            overflow: hidden;
        }
        .ticket {
            width: 100%;
            height: 100%;
            position: relative;
            background-image: url(data:image/png;base64,{{.BackgroundBase64}});
            background-size: cover;
            background-position: center;
        }
        .top-section {
            position: absolute;
            left: {{mul .Width 0.083}}px;
            top: {{mul .Height 0.08}}px;
            width: 55%;
            color: {{if .IsColor}}#ffffff{{else}}#000000{{end}};
        }
        .service-label {
            font-size: {{mul .Width 0.065}}px;
            font-weight: normal;
            margin-bottom: {{mul .Height 0.02}}px;
            letter-spacing: 0.5px;
        }
        .service-name {
            font-size: {{mul .Width 0.070}}px; /* Еще раз уменьшен шрифт */
            font-weight: bold;
            line-height: 1.15;
            text-transform: uppercase;
            word-wrap: break-word;
            color: {{if .IsColor}}#ffffff{{else}}#000000{{end}};
            margin-top: {{mul .Height 0.015}}px;
        }
        .ticket-number-section {
            position: absolute;
            left: {{mul .Width 0.083}}px;
            top: {{mul .Height 0.54}}px;
            color: #ffffff;
            z-index: 10;
        }
        .ticket-number-label {
            font-size: {{mul .Width 0.058}}px;
            font-weight: normal;
            margin-bottom: {{mul .Height 0.015}}px; /* Уменьшен отступ */
            letter-spacing: 0.5px;
        }
        .ticket-number {
            font-size: {{mul .Width 0.19}}px;
            font-weight: bold;
            line-height: 1;
            margin-top: {{mul .Height 0.01}}px;
        }
        .time-section {
            position: absolute;
            left: {{mul .Width 0.083}}px;
            top: {{mul .Height 0.76}}px;
            color: #ffffff;
            z-index: 10;
        }
        .time-label {
            font-size: {{mul .Width 0.058}}px;
            font-weight: normal;
            margin-bottom: {{mul .Height 0.01}}px; /* Уменьшен отступ */
            letter-spacing: 0.5px;
        }
        .date {
            font-size: {{mul .Width 0.065}}px;
            font-weight: bold;
            margin-bottom: {{mul .Height 0.01}}px; /* Отступ снизу для одинакового интервала */
        }
        .time {
            font-size: {{mul .Width 0.065}}px;
            font-weight: bold;
        }
        .qr-code {
            position: absolute;
            right: {{mul .Width 0.065}}px;
            bottom: {{mul .Height 0.11}}px;
            width: {{mul .Width 0.35}}px;
            height: {{mul .Width 0.35}}px;
            background: #ffffff;
            border-radius: {{mul .Width 0.03}}px;
            padding: {{mul .Width 0.015}}px;
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 10;
        }
        .qr-code img {
            width: 100%;
            height: 100%;
            display: block;
        }
        .waiting-text {
            position: absolute;
            bottom: {{mul .Height 0.035}}px;
            left: 50%;
            transform: translateX(-50%);
            font-size: {{mul .Width 0.028}}px;
            font-weight: normal;
            text-transform: uppercase;
            color: #ffffff;
            white-space: nowrap;
            letter-spacing: 0.3px;
        }
    </style>
</head>
<body>
    <div class="ticket">
        <div class="top-section">
            <div class="service-label">УСЛУГА</div>
            <div class="service-name">{{.ServiceName}}</div>
        </div>
        
        <div class="ticket-number-section">
            <div class="ticket-number-label">НОМЕР ТАЛОНА</div>
            <div class="ticket-number">{{.TicketNumber}}</div>
        </div>
        
        <div class="time-section">
            <div class="time-label">ВРЕМЯ</div>
            <div class="date">{{.Date}}</div>
            <div class="time">{{.Time}}</div>
        </div>
        
        <div class="qr-code">
            <img src="data:image/png;base64,{{.QRCodeBase64}}" alt="QR Code">
        </div>
        
        {{if .WaitingText}}
        <div class="waiting-text">{{.WaitingText}}</div>
        {{end}}
    </div>
</body>
</html>
`

// wrapText разбивает текст на строки с учетом максимальной длины символов
func wrapText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	var lines []string
	runes := []rune(text)

	for len(runes) > 0 {
		if len(runes) <= maxLength {
			lines = append(lines, string(runes))
			break
		}

		breakPoint := maxLength
		for i := maxLength - 1; i >= 0; i-- {
			if runes[i] == ' ' {
				breakPoint = i
				break
			}
		}

		if breakPoint == maxLength && runes[maxLength-1] != ' ' {
			for i := maxLength; i < len(runes); i++ {
				if runes[i] == ' ' {
					breakPoint = i
					break
				}
			}
			if breakPoint == maxLength {
				breakPoint = maxLength
			}
		}

		lines = append(lines, strings.TrimSpace(string(runes[:breakPoint])))
		if breakPoint < len(runes) {
			runes = runes[breakPoint:]
			for len(runes) > 0 && runes[0] == ' ' {
				runes = runes[1:]
			}
		} else {
			break
		}
	}

	return strings.Join(lines, "<br>")
}

// fileToBase64 преобразует файл в base64
func fileToBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// generateQRCode генерирует QR-код и возвращает его в base64
func generateQRCode(data []byte, size int) (string, error) {
	qrCode, err := qrcode.New(string(data), qrcode.Medium)
	if err != nil {
		return "", err
	}

	qrBytes, err := qrCode.PNG(size)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(qrBytes), nil
}

// ConvertToMonochrome применяет пороговое преобразование для конвертации в 1-битное монохромное изображение,
// сохраняя резкие края текста и QR-кодов.
func ConvertToMonochrome(src image.Image) image.Image {
	bounds := src.Bounds()
	// Создаем новое палитровое изображение для итогового 1-битного вывода.
	// Индекс 0 в палитре - Белый, индекс 1 - Черный.
	palette := color.Palette{color.White, color.Black}
	dst := image.NewPaletted(bounds, palette)

	// Пороговое значение (0-255). 128 - это 50% серого, хороший компромисс.
	const threshold = 128

	// Итерируемся по каждому пикселю
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Получаем исходный цвет и конвертируем его в оттенок серого,
			// используя стандартную формулу светимости для лучшего восприятия яркости.
			r, g, b, _ := src.At(x, y).RGBA()

			// Значения из RGBA() являются 16-битными (0-65535), поэтому мы сдвигаем их вправо на 8 бит,
			// чтобы получить 8-битное значение (0-255).
			gray := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)

			// Применяем порог
			if gray > threshold {
				dst.SetColorIndex(x, y, 0) // Белый
			} else {
				dst.SetColorIndex(x, y, 1) // Черный
			}
		}
	}
	return dst
}

// setDPI встраивает чанк pHYs в PNG-файл для установки DPI.
func setDPI(pngBytes []byte, dpi int) ([]byte, error) {
	// Конвертируем DPI в пиксели на метр
	ppm := uint32(float64(dpi) / 0.0254)

	// pHYs чанк имеет длину 9 байт
	// 4 байта: pixels per unit, X axis
	// 4 байта: pixels per unit, Y axis
	// 1 байт: unit specifier (1 = meters)
	pHYsData := make([]byte, 9)
	binary.BigEndian.PutUint32(pHYsData[0:4], ppm)
	binary.BigEndian.PutUint32(pHYsData[4:8], ppm)
	pHYsData[8] = 1 // Unit is meter

	// Создаем буфер для записи чанка
	chunk := new(bytes.Buffer)
	// Записываем тип чанка
	chunk.Write([]byte("pHYs"))
	// Записываем данные
	chunk.Write(pHYsData)

	// Вычисляем CRC32
	crc := crc32.NewIEEE()
	crc.Write(chunk.Bytes())
	crcSum := crc.Sum(nil)

	// Собираем полный чанк: Длина + Тип + Данные + CRC
	finalChunk := new(bytes.Buffer)
	// Длина данных (9)
	binary.Write(finalChunk, binary.BigEndian, uint32(9))
	// Тип и данные
	finalChunk.Write(chunk.Bytes())
	// CRC
	finalChunk.Write(crcSum)

	// Вставляем чанк в PNG. pHYs должен идти после IHDR и перед IDAT.
	// IHDR всегда первый чанк после 8-байтной сигнатуры PNG, и он имеет фиксированную длину 25 байт.
	// (4-длина, 4-тип, 13-данные, 4-crc)
	// Вставляем наш чанк сразу после IHDR.
	insertionPoint := 8 + 25
	if len(pngBytes) < insertionPoint {
		return nil, fmt.Errorf("PNG-файл слишком короткий для вставки DPI")
	}

	// Собираем новый PNG
	result := new(bytes.Buffer)
	result.Write(pngBytes[:insertionPoint])
	result.Write(finalChunk.Bytes())
	result.Write(pngBytes[insertionPoint:])

	return result.Bytes(), nil
}

// GenerateTicketImage генерирует изображение талона с фоном, текстом и QR-кодом
func GenerateTicketImage(config TicketConfig, isColor bool, dpi int) ([]byte, error) {
	// Преобразуем фоновое изображение в base64
	bgBase64, err := fileToBase64(config.BackgroundPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения фонового изображения: %v", err)
	}

	// Преобразуем шрифты в base64
	fontBase64, err := fileToBase64(config.FontPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения шрифта: %v", err)
	}

	boldFontBase64, err := fileToBase64(config.BoldFontPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения жирного шрифта: %v", err)
	}

	// Генерируем QR-код
	qrSize := int(float64(config.Width) * 0.28)
	qrBase64, err := generateQRCode(config.QRData, qrSize)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации QR-кода: %v", err)
	}

	// Подготавливаем данные для шаблона
	waitingText := ""
	if config.WaitingNumber > 0 {
		waitingText = strings.ToUpper(fmt.Sprintf("Перед вами %d человек в очереди", config.WaitingNumber))
	}

	data := ticketHTMLData{
		Width:            config.Width,
		Height:           config.Height,
		BackgroundBase64: bgBase64,
		QRCodeBase64:     qrBase64,
		ServiceName:      wrapText(strings.ToUpper(config.ServiceName), 12),
		TicketNumber:     config.TicketNumber,
		Date:             config.DateTime.Format("02.01.2006"),
		Time:             config.DateTime.Format("15:04:05"),
		WaitingText:      waitingText,
		IsColor:          isColor,
		FontBase64:       fontBase64,
		BoldFontBase64:   boldFontBase64,
	}

	// Создаем функции для шаблона
	funcMap := template.FuncMap{
		"mul": func(a interface{}, b float64) int {
			var aFloat float64
			switch v := a.(type) {
			case int:
				aFloat = float64(v)
			case float64:
				aFloat = v
			default:
				aFloat = 0
			}
			return int(aFloat * b)
		},
		"div": func(a interface{}, b float64) int {
			var aFloat float64
			switch v := a.(type) {
			case int:
				aFloat = float64(v)
			case float64:
				aFloat = v
			default:
				aFloat = 1
			}
			return int(aFloat / b)
		},
	}

	// Парсим и выполняем шаблон
	tmpl, err := template.New("ticket").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга шаблона: %v", err)
	}

	var htmlBuf bytes.Buffer
	err = tmpl.Execute(&htmlBuf, data)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения шаблона: %v", err)
	}

	// Преобразуем HTML в PNG с помощью chromedp
	rgbaPngBytes, err := htmlToPNG(htmlBuf.String(), config.Width, config.Height)
	if err != nil {
		return nil, fmt.Errorf("ошибка преобразования HTML в PNG: %v", err)
	}

	// Декодируем RGBA PNG
	img, err := png.Decode(bytes.NewReader(rgbaPngBytes))
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования RGBA PNG: %w", err)
	}

	// Применяем пороговое преобразование для конвертации в монохромный
	monoImage := ConvertToMonochrome(img)

	// Кодируем монохромное изображение обратно в PNG
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, monoImage); err != nil {
		return nil, fmt.Errorf("ошибка кодирования монохромного PNG: %w", err)
	}

	// Внедряем информацию о DPI в готовый PNG файл
	finalPngBytes, err := setDPI(pngBuf.Bytes(), dpi)
	if err != nil {
		return nil, fmt.Errorf("ошибка установки DPI: %w", err)
	}

	return finalPngBytes, nil
}

// htmlToPNG преобразует HTML в PNG используя headless Chrome
func htmlToPNG(htmlContent string, width, height int) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(width), int64(height)),
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),
		chromedp.Sleep(1000*time.Millisecond),
		chromedp.FullScreenshot(&buf, 100),
	)

	if err != nil {
		return nil, err
	}

	return buf, nil
}
