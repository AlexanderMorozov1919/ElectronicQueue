package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
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

// --- НОВАЯ ВСПОМОГАТЕЛЬНАЯ ФУНКЦИЯ ---
// drawRoundedRect рисует закрашенный прямоугольник с закругленными углами.
func drawRoundedRect(img *image.RGBA, r image.Rectangle, c color.Color, cornerRadius int) {
	// Рисуем центральный прямоугольник
	draw.Draw(img, image.Rect(r.Min.X+cornerRadius, r.Min.Y, r.Max.X-cornerRadius, r.Max.Y), &image.Uniform{c}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(r.Min.X, r.Min.Y+cornerRadius, r.Max.X, r.Max.Y-cornerRadius), &image.Uniform{c}, image.Point{}, draw.Src)

	// Рисуем круги в углах для создания закруглений
	// Верхний левый
	drawCircle(img, r.Min.X+cornerRadius, r.Min.Y+cornerRadius, cornerRadius, c)
	// Верхний правый
	drawCircle(img, r.Max.X-cornerRadius-1, r.Min.Y+cornerRadius, cornerRadius, c)
	// Нижний левый
	drawCircle(img, r.Min.X+cornerRadius, r.Max.Y-cornerRadius-1, cornerRadius, c)
	// Нижний правый
	drawCircle(img, r.Max.X-cornerRadius-1, r.Max.Y-cornerRadius-1, cornerRadius, c)
}

// drawCircle - вспомогательная функция для рисования круга.
func drawCircle(img *image.RGBA, x0, y0, r int, c color.Color) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				img.Set(x0+x, y0+y, c)
			}
		}
	}
}

// resizeImage масштабирует изображение с сохранением пропорций
func resizeImage(src image.Image, width, height int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	scaleX := float64(width) / float64(srcWidth)
	scaleY := float64(height) / float64(srcHeight)
	scale := scaleX
	if scaleY > scaleX {
		scale = scaleY
	}
	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)
	offsetX := (width - newWidth) / 2
	offsetY := (height - newHeight) / 2
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x-offsetX) / scale)
			srcY := int(float64(y-offsetY) / scale)
			if srcX >= 0 && srcX < srcWidth && srcY >= 0 && srcY < srcHeight {
				srcColor := src.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y)
				dst.Set(x, y, srcColor)
			}
		}
	}
	return dst
}

// wrapText разбивает текст на строки
func wrapText(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
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
	return lines
}

// --- ОСНОВНАЯ ФУНКЦИЯ С ИЗМЕНЕНИЯМИ ---
// GenerateTicketImage генерирует изображение талона с фоном, текстом и QR-кодом
func GenerateTicketImage(config TicketConfig, isColor bool) ([]byte, error) {
	bgFile, err := os.Open(config.BackgroundPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия фонового изображения: %v", err)
	}
	defer bgFile.Close()

	bgImg, _, err := image.Decode(bgFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования фонового изображения: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))
	scaledBg := resizeImage(bgImg, config.Width, config.Height)
	draw.Draw(img, img.Bounds(), scaledBg, image.Point{}, draw.Src)

	fontBytes, err := os.ReadFile(config.FontPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла шрифта: %v", err)
	}
	ttfFont, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга шрифта: %v", err)
	}

	boldFontBytes, err := os.ReadFile(config.BoldFontPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла жирного шрифта: %v", err)
	}
	boldTtfFont, err := truetype.Parse(boldFontBytes)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга жирного шрифта: %v", err)
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 255}))

	// Размеры шрифтов, пропорциональные ширине
	labelSize := float64(config.Width) * 0.062
	// *** ИЗМЕНЕНИЕ: Уменьшен размер шрифта для услуги, чтобы избежать переноса ***
	serviceSize := float64(config.Width) * 0.068
	numberSize := float64(config.Width) * 0.17
	timeSize := float64(config.Width) * 0.062
	waitingSize := float64(config.Width) * 0.035

	// --- Рисуем УСЛУГА ---
	textColor := image.NewUniform(color.RGBA{0, 0, 0, 255}) // Черный для текста на белом фоне
	c.SetFont(ttfFont)
	c.SetFontSize(labelSize)
	c.SetSrc(textColor)
	pt := freetype.Pt(config.Width/12, int(float64(config.Height)*0.11))
	_, _ = c.DrawString("УСЛУГА", pt)

	// --- Рисуем название услуги ---
	c.SetFont(boldTtfFont)
	c.SetFontSize(serviceSize)
	// *** ИЗМЕНЕНИЕ: Увеличена длина строки, чтобы предотвратить перенос ***
	serviceLines := wrapText(strings.ToUpper(config.ServiceName), 20)
	startY := float64(config.Height) * 0.18
	lineHeight := serviceSize * 1.2
	for i, line := range serviceLines {
		pt = freetype.Pt(config.Width/12, int(startY+float64(i)*lineHeight))
		_, _ = c.DrawString(strings.TrimSpace(line), pt)
	}

	// --- Рисуем НОМЕР ТАЛОНА ---
	whiteColor := image.NewUniform(color.RGBA{255, 255, 255, 255})
	c.SetSrc(whiteColor)
	c.SetFont(ttfFont)
	c.SetFontSize(labelSize)
	pt = freetype.Pt(config.Width/12, int(float64(config.Height)*0.57))
	_, _ = c.DrawString("НОМЕР ТАЛОНА", pt)

	// --- Рисуем номер талона (A008) ---
	c.SetFont(boldTtfFont)
	c.SetFontSize(numberSize)
	pt = freetype.Pt(config.Width/13, int(float64(config.Height)*0.69))
	_, _ = c.DrawString(config.TicketNumber, pt)

	// --- Рисуем ВРЕМЯ ---
	c.SetFont(ttfFont)
	c.SetFontSize(labelSize)
	timeStartY := float64(config.Height) * 0.78
	pt = freetype.Pt(config.Width/12, int(timeStartY))
	_, _ = c.DrawString("ВРЕМЯ", pt)

	// --- Рисуем дату и время ---
	c.SetFont(boldTtfFont)
	c.SetFontSize(timeSize)
	pt = freetype.Pt(config.Width/12, int(timeStartY+float64(config.Height)*0.06))
	_, _ = c.DrawString(config.DateTime.Format("02.01.2006"), pt)
	pt = freetype.Pt(config.Width/12, int(timeStartY+float64(config.Height)*0.11))
	_, _ = c.DrawString(config.DateTime.Format("15:04:05"), pt)

	// --- *** НАЧАЛО ИЗМЕНЕНИЙ: ЛОГИКА QR-КОДА *** ---
	// 1. Рассчитываем размер и положение белого фона для QR-кода
	// Увеличиваем размер фона
	qrBgSize := int(float64(config.Width) / 2.8)
	qrBgCornerRadius := qrBgSize / 10
	qrBgX := config.Width - qrBgSize - config.Width/15
	qrBgY := int(float64(config.Height) * 0.6)
	qrBgRect := image.Rect(qrBgX, qrBgY, qrBgX+qrBgSize, qrBgY+qrBgSize)

	// 2. Рисуем белый закругленный прямоугольник на основном изображении
	drawRoundedRect(img, qrBgRect, color.White, qrBgCornerRadius)

	// 3. Рассчитываем размер и положение самого QR-кода (он будет меньше фона)
	// Увеличиваем QR-код, оставляя небольшие отступы
	margin := int(float64(qrBgSize) * 0.08)
	qrCodeSize := qrBgSize - (2 * margin)
	qrCodeX := qrBgX + margin
	qrCodeY := qrBgY + margin

	// 4. Генерируем стандартный (не закругленный) QR-код
	qrCode, err := qrcode.New(string(config.QRData), qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания QR-кода: %v", err)
	}
	qrCode.DisableBorder = true // Убираем собственную рамку QR-кода
	qrImg := qrCode.Image(qrCodeSize)

	// 5. Накладываем QR-код на белый фон
	draw.Draw(img, image.Rect(qrCodeX, qrCodeY, qrCodeX+qrCodeSize, qrCodeY+qrCodeSize), qrImg, image.Point{}, draw.Over)
	// --- *** КОНЕЦ ИЗМЕНЕНИЙ: ЛОГИКА QR-КОДА *** ---

	// --- Рисуем надпись об очереди ---
	if config.WaitingNumber > 0 {
		c.SetFont(ttfFont)
		c.SetFontSize(waitingSize)
		c.SetSrc(whiteColor)
		queueText := strings.ToUpper(fmt.Sprintf("Перед вами %d человек в очереди", config.WaitingNumber))

		face := truetype.NewFace(ttfFont, &truetype.Options{Size: waitingSize, DPI: 72})
		bounds, _ := font.BoundString(face, queueText)
		textWidthPixels := int(bounds.Max.X-bounds.Min.X) >> 6
		textY := float64(config.Height) * 0.96
		textX := (config.Width - textWidthPixels) / 2

		pt = freetype.Pt(textX, int(textY))
		_, _ = c.DrawString(queueText, pt)
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("ошибка кодирования PNG: %v", err)
	}

	return buf.Bytes(), nil
}
