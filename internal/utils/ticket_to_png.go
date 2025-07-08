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
)

// TicketConfig содержит конфигурацию для генерации талона
type TicketConfig struct {
	Width          int    // Ширина изображения в пикселях
	Height         int    // Высота изображения в пикселях
	QRData         []byte // Данные для QR-кода
	FontPath       string // Путь к обычному шрифту
	BoldFontPath   string // Путь к жирному шрифту
	BackgroundPath string // Путь к фоновому изображению
}

// TicketData содержит данные для отображения на талоне
type TicketData struct {
	ServiceName  string
	TicketNumber string
	DateTime     time.Time
}

// resizeImage масштабирует изображение с сохранением пропорций и заполнением фона
func resizeImage(src image.Image, width, height int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	// Создаем новое изображение с нужными размерами
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Вычисляем коэффициенты масштабирования
	scaleX := float64(width) / float64(srcWidth)
	scaleY := float64(height) / float64(srcHeight)
	scale := scaleX
	if scaleY > scaleX {
		scale = scaleY
	}

	// Новые размеры после масштабирования
	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// Позиция для центрирования
	offsetX := (width - newWidth) / 2
	offsetY := (height - newHeight) / 2

	// Рисуем масштабированное изображение
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Вычисляем соответствующие координаты в исходном изображении
			srcX := int(float64(x-offsetX) / scale)
			srcY := int(float64(y-offsetY) / scale)

			if srcX >= 0 && srcX < srcWidth && srcY >= 0 && srcY < srcHeight {
				// Пиксель находится в пределах масштабированного изображения
				srcColor := src.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y)
				dst.Set(x, y, srcColor)
			}
			// Если пиксель вне изображения, оставляем прозрачным (по умолчанию)
		}
	}

	return dst
}

// wrapText разбивает текст на строки с учетом максимальной длины символов
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

		// Ищем последний пробел в пределах maxLength
		breakPoint := maxLength
		for i := maxLength - 1; i >= 0; i-- {
			if runes[i] == ' ' {
				breakPoint = i
				break
			}
		}

		// Если пробел не найден, разрываем по maxLength
		if breakPoint == maxLength && runes[maxLength-1] != ' ' {
			// Ищем пробел после maxLength
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
			// Убираем ведущие пробелы
			for len(runes) > 0 && runes[0] == ' ' {
				runes = runes[1:]
			}
		} else {
			break
		}
	}

	return lines
}

// createRoundedQRCode создает QR-код с закругленными краями
func createRoundedQRCode(data []byte, size int) (image.Image, error) {
	qrCode, err := qrcode.New(string(data), qrcode.Medium)
	if err != nil {
		return nil, err
	}

	qrImg := qrCode.Image(size)

	// Создаем новое изображение для закругленного QR-кода
	rounded := image.NewRGBA(image.Rect(0, 0, size, size))

	// Радиус закругления
	radius := size / 20

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Проверяем, находится ли пиксель в закругленных углах
			inCorner := false

			// Левый верхний угол
			if x < radius && y < radius {
				dx := radius - x
				dy := radius - y
				if dx*dx+dy*dy > radius*radius {
					inCorner = true
				}
			}
			// Правый верхний угол
			if x >= size-radius && y < radius {
				dx := x - (size - radius - 1)
				dy := radius - y
				if dx*dx+dy*dy > radius*radius {
					inCorner = true
				}
			}
			// Левый нижний угол
			if x < radius && y >= size-radius {
				dx := radius - x
				dy := y - (size - radius - 1)
				if dx*dx+dy*dy > radius*radius {
					inCorner = true
				}
			}
			// Правый нижний угол
			if x >= size-radius && y >= size-radius {
				dx := x - (size - radius - 1)
				dy := y - (size - radius - 1)
				if dx*dx+dy*dy > radius*radius {
					inCorner = true
				}
			}

			if !inCorner {
				rounded.Set(x, y, qrImg.At(x, y))
			}
		}
	}

	return rounded, nil
}

// GenerateTicketImage генерирует изображение талона с фоном, текстом и QR-кодом
func GenerateTicketImage(config TicketConfig, data TicketData, isColor bool) ([]byte, error) {
	// Загружаем фоновое изображение
	bgFile, err := os.Open(config.BackgroundPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия фонового изображения: %v", err)
	}
	defer bgFile.Close()

	// Декодируем фоновое изображение
	bgImg, _, err := image.Decode(bgFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования фонового изображения: %v", err)
	}

	// Создаем новое изображение с заданными размерами
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	// Масштабируем фоновое изображение с сохранением пропорций
	scaledBg := resizeImage(bgImg, config.Width, config.Height)
	draw.Draw(img, img.Bounds(), scaledBg, image.Point{}, draw.Src)

	// Загружаем обычный шрифт
	fontBytes, err := os.ReadFile(config.FontPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла шрифта: %v", err)
	}

	ttfFont, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга шрифта: %v", err)
	}

	// Загружаем жирный шрифт
	boldFontBytes, err := os.ReadFile(config.BoldFontPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла жирного шрифта: %v", err)
	}

	boldTtfFont, err := truetype.Parse(boldFontBytes)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга жирного шрифта: %v", err)
	}

	// Создаем контекст для рисования текста
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 255})) // Белый цвет

	// Вычисляем позиции для текста (пропорционально размеру изображения)
	labelSize := float64(config.Width) * 0.062   // Размер меток
	serviceSize := float64(config.Width) * 0.071 // Размер названия услуги
	numberSize := float64(config.Width) * 0.17   // Размер номера талона
	timeSize := float64(config.Width) * 0.062    // Размер времени

	// Рисуем заголовок "УСЛУГА" (обычный шрифт)
	c.SetFont(ttfFont)
	c.SetFontSize(labelSize)
	if isColor {
		c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 255})) // Белый цвет
	} else {
		c.SetSrc(image.NewUniform(color.RGBA{0, 0, 0, 255})) // Чёрный цвет
	}
	pt := freetype.Pt(config.Width/12, int(float64(config.Height)*0.11))
	_, err = c.DrawString("УСЛУГА", pt)
	if err != nil {
		return nil, fmt.Errorf("ошибка рисования заголовка: %v", err)
	}

	// Рисуем название услуги (жирный шрифт, с переносом по длине)
	c.SetFont(boldTtfFont)
	c.SetFontSize(serviceSize)
	serviceLines := wrapText(strings.ToUpper(data.ServiceName), 10) // Максимум 10 символов на строку, верхний регистр

	startY := float64(config.Height) * 0.18
	lineHeight := serviceSize * 1.2

	for i, line := range serviceLines {
		pt = freetype.Pt(config.Width/12, int(startY+float64(i)*lineHeight))
		_, err = c.DrawString(strings.TrimSpace(line), pt)
		if err != nil {
			return nil, fmt.Errorf("ошибка рисования текста услуги: %v", err)
		}
	}
	c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 255})) // Белый цвет

	// Рисуем "НОМЕР ТАЛОНА" (обычный шрифт)
	c.SetFont(ttfFont)
	c.SetFontSize(labelSize)
	pt = freetype.Pt(config.Width/12, int(float64(config.Height)*0.58))
	_, err = c.DrawString("НОМЕР ТАЛОНА", pt)
	if err != nil {
		return nil, fmt.Errorf("ошибка рисования заголовка номера: %v", err)
	}

	// Рисуем номер талона (жирный шрифт)
	c.SetFont(boldTtfFont)
	c.SetFontSize(numberSize)
	pt = freetype.Pt(config.Width/13, int(float64(config.Height)*0.7))
	_, err = c.DrawString(data.TicketNumber, pt)
	if err != nil {
		return nil, fmt.Errorf("ошибка рисования номера талона: %v", err)
	}

	// Рисуем информацию о времени
	c.SetFont(ttfFont)
	c.SetFontSize(labelSize)
	timeStartY := float64(config.Height) * 0.81

	// Заголовок "ВРЕМЯ"
	pt = freetype.Pt(config.Width/12, int(timeStartY))
	_, err = c.DrawString("ВРЕМЯ", pt)
	if err != nil {
		return nil, fmt.Errorf("ошибка рисования заголовка времени: %v", err)
	}

	// Дата и время (жирный шрифт)
	c.SetFont(boldTtfFont)
	c.SetFontSize(timeSize)

	// Дата
	pt = freetype.Pt(config.Width/12, int(timeStartY+float64(config.Height)*0.07))
	_, err = c.DrawString(data.DateTime.Format("02.01.2006"), pt)
	if err != nil {
		return nil, fmt.Errorf("ошибка рисования даты: %v", err)
	}

	// Время
	pt = freetype.Pt(config.Width/12, int(timeStartY+float64(config.Height)*0.12))
	_, err = c.DrawString(data.DateTime.Format("15:04:05"), pt)
	if err != nil {
		return nil, fmt.Errorf("ошибка рисования времени: %v", err)
	}

	// Генерируем QR-код с закругленными краями (увеличенный размер)
	qrSize := config.Width / 4 // Увеличенный размер QR-кода
	qrImg, err := createRoundedQRCode(config.QRData, qrSize)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания QR-кода: %v", err)
	}

	// Позиция QR-кода (правый нижний угол)
	qrX := config.Width - qrSize - config.Width/10
	qrY := config.Height - qrSize - config.Height/15
	qrRect := image.Rect(qrX, qrY, qrX+qrSize, qrY+qrSize)

	// Накладываем QR-код на изображение
	draw.Draw(img, qrRect, qrImg, image.Point{}, draw.Over)

	// Сохраняем изображение в буфер
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("ошибка кодирования PNG: %v", err)
	}

	return buf.Bytes(), nil
}

// GenerateTicketImageWithConfig генерирует талон с заданным фоном
func GenerateTicketImageWithConfig(baseSize int, qrData []byte, data TicketData, background string, isColor bool) ([]byte, error) {
	sqrt2 := 1.414
	width := int(float64(baseSize) / sqrt2)
	height := baseSize

	config := TicketConfig{
		Width:          width,
		Height:         height,
		QRData:         qrData,
		FontPath:       "assets/fonts/Arial.ttf",
		BoldFontPath:   "assets/fonts/Arial_bold.ttf",
		BackgroundPath: background,
	}

	return GenerateTicketImage(config, data, isColor)
}
