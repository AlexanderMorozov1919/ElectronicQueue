// путь до файла: D:\vs\go\ElectronicQueue\internal\utils\audio.go
package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// WAVHeader представляет заголовок WAV файла
type WAVHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

// WAVData представляет данные WAV файла
type WAVData struct {
	Header WAVHeader
	Data   []byte
}

// GenerateAnnouncementWav создает WAV файл с озвучкой талона
// ИЗМЕНЕНО: Добавлен параметр backgroundMusicEnabled
func GenerateAnnouncementWav(ticketNumber, windowNumber, audioDir string, backgroundMusicEnabled bool) ([]byte, error) {
	// Парсим номер талона
	letter, number, err := parseTicketNumber(ticketNumber)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга номера талона: %v", err)
	}

	// Создаем последовательность файлов для воспроизведения
	var audioFiles []string

	// 1. Клиент_номер.wav
	audioFiles = append(audioFiles, filepath.Join(audioDir, "Klient_nomer.wav"))

	// 2. Буква талона
	audioFiles = append(audioFiles, filepath.Join(audioDir, fmt.Sprintf("%s.wav", letter)))

	// 3. Номер талона (разбиваем на составляющие)
	numberFiles, err := getNumberFiles(number, audioDir)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения файлов для номера: %v", err)
	}
	audioFiles = append(audioFiles, numberFiles...)

	// 4. Подойдите_к_окну_номер.wav
	audioFiles = append(audioFiles, filepath.Join(audioDir, "Podoidite_k_oknu_nomer.wav"))

	// 5. Номер окна
	windowFiles, err := getNumberFiles(windowNumber, audioDir)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения файлов для номера окна: %v", err)
	}
	audioFiles = append(audioFiles, windowFiles...)

	// Загружаем и объединяем основные аудиофайлы
	mainAudio, err := concatenateWavFiles(audioFiles)
	if err != nil {
		return nil, fmt.Errorf("ошибка объединения аудиофайлов: %v", err)
	}

	// ИЗМЕНЕНО: Блок микширования теперь условный
	if backgroundMusicEnabled {
		// Загружаем фоновую музыку
		backgroundFile := filepath.Join(audioDir, "The_Time_Is_Now.wav")
		backgroundAudio, err := loadWavFile(backgroundFile)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки фоновой музыки: %v", err)
		}

		// Микшируем основную дорожку с фоновой
		result, err := mixAudioTracks(mainAudio, backgroundAudio)
		if err != nil {
			return nil, fmt.Errorf("ошибка микширования аудиодорожек: %v", err)
		}

		return result, nil
	}

	// ИЗМЕНЕНО: Если музыка отключена, возвращаем только основную дорожку
	// Создаем результирующий WAV файл из основного аудио
	var result bytes.Buffer
	if err := binary.Write(&result, binary.LittleEndian, mainAudio.Header); err != nil {
		return nil, fmt.Errorf("ошибка записи заголовка основного аудио: %w", err)
	}
	if _, err := result.Write(mainAudio.Data); err != nil {
		return nil, fmt.Errorf("ошибка записи данных основного аудио: %w", err)
	}

	return result.Bytes(), nil
}

// parseTicketNumber парсит номер талона в формате "A007" или "C21"
func parseTicketNumber(ticket string) (string, string, error) {
	re := regexp.MustCompile(`^([A-Z])(\d+)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(ticket))

	if len(matches) != 3 {
		return "", "", fmt.Errorf("неверный формат номера талона: %s", ticket)
	}

	letter := matches[1]
	number := matches[2]

	// Удаляем ведущие нули
	numberInt, err := strconv.Atoi(number)
	if err != nil {
		return "", "", fmt.Errorf("ошибка преобразования номера: %v", err)
	}

	return letter, strconv.Itoa(numberInt), nil
}

// getNumberFiles возвращает список файлов для озвучки числа
func getNumberFiles(number, audioDir string) ([]string, error) {
	num, err := strconv.Atoi(number)
	if err != nil {
		return nil, fmt.Errorf("ошибка преобразования номера: %v", err)
	}

	if num < 1 || num > 99 {
		return nil, fmt.Errorf("номер должен быть от 1 до 99")
	}

	var files []string

	if num <= 20 {
		// Для чисел 1-20 есть отдельные файлы
		files = append(files, filepath.Join(audioDir, fmt.Sprintf("%d.wav", num)))
	} else {
		// Для чисел 21-99 разбиваем на десятки и единицы
		tens := (num / 10) * 10
		ones := num % 10

		files = append(files, filepath.Join(audioDir, fmt.Sprintf("%d.wav", tens)))

		if ones > 0 {
			files = append(files, filepath.Join(audioDir, fmt.Sprintf("%d.wav", ones)))
		}
	}

	return files, nil
}

// loadWavFile загружает WAV файл
func loadWavFile(filename string) (*WAVData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла %s: %v", filename, err)
	}
	defer file.Close()

	var header WAVHeader
	err = binary.Read(file, binary.LittleEndian, &header)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения заголовка WAV: %v", err)
	}

	// Проверяем формат
	if string(header.ChunkID[:]) != "RIFF" || string(header.Format[:]) != "WAVE" {
		return nil, fmt.Errorf("неверный формат файла: %s", filename)
	}

	data := make([]byte, header.Subchunk2Size)
	_, err = io.ReadFull(file, data)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения данных WAV: %v", err)
	}

	return &WAVData{
		Header: header,
		Data:   data,
	}, nil
}

// concatenateWavFiles объединяет несколько WAV файлов
func concatenateWavFiles(filenames []string) (*WAVData, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("список файлов пуст")
	}

	// Загружаем первый файл для получения параметров
	firstFile, err := loadWavFile(filenames[0])
	if err != nil {
		return nil, err
	}

	var concatenatedData []byte
	concatenatedData = append(concatenatedData, firstFile.Data...)

	// Объединяем остальные файлы
	for _, filename := range filenames[1:] {
		wavData, err := loadWavFile(filename)
		if err != nil {
			return nil, err
		}

		// Проверяем совместимость параметров
		if wavData.Header.SampleRate != firstFile.Header.SampleRate ||
			wavData.Header.NumChannels != firstFile.Header.NumChannels ||
			wavData.Header.BitsPerSample != firstFile.Header.BitsPerSample {
			return nil, fmt.Errorf("несовместимые параметры аудио в файле %s", filename)
		}

		concatenatedData = append(concatenatedData, wavData.Data...)
	}

	// Создаем новый заголовок
	newHeader := firstFile.Header
	newHeader.Subchunk2Size = uint32(len(concatenatedData))
	newHeader.ChunkSize = 36 + newHeader.Subchunk2Size

	return &WAVData{
		Header: newHeader,
		Data:   concatenatedData,
	}, nil
}

// mixAudioTracks микширует две аудиодорожки
func mixAudioTracks(main, background *WAVData) ([]byte, error) {
	// Нормализуем параметры фоновой дорожки под основную
	normalizedBackground, err := normalizeWavData(background, main.Header.SampleRate, main.Header.NumChannels, main.Header.BitsPerSample)
	if err != nil {
		return nil, fmt.Errorf("ошибка нормализации фоновой дорожки: %v", err)
	}

	background = normalizedBackground

	// Микшируем в float32 для точности
	mainSamples := make([]float32, 0)
	backgroundSamples := make([]float32, 0)

	// Конвертируем основную дорожку
	if main.Header.BitsPerSample == 32 && main.Header.AudioFormat == 3 {
		mainSamples = bytesToFloat32Samples(main.Data)
	} else if main.Header.BitsPerSample == 16 && main.Header.AudioFormat == 1 {
		mainSamples = int16SamplesToFloat32(bytesToInt16Samples(main.Data))
	}

	// Конвертируем фоновую дорожку
	if background.Header.BitsPerSample == 32 && background.Header.AudioFormat == 3 {
		backgroundSamples = bytesToFloat32Samples(background.Data)
	} else if background.Header.BitsPerSample == 16 && background.Header.AudioFormat == 1 {
		backgroundSamples = int16SamplesToFloat32(bytesToInt16Samples(background.Data))
	}

	// Определяем длину результирующих семплов
	maxSamples := len(mainSamples)
	if len(backgroundSamples) > maxSamples {
		maxSamples = len(backgroundSamples)
	}

	mixedSamples := make([]float32, maxSamples)

	// Микшируем семплы
	for i := 0; i < maxSamples; i++ {
		var mainSample, backgroundSample float32

		if i < len(mainSamples) {
			mainSample = mainSamples[i]
		}
		if i < len(backgroundSamples) {
			backgroundSample = backgroundSamples[i]
		}

		// Микшируем семплы (фоновая дорожка с уменьшенной громкостью)
		mixedSamples[i] = mainSample + backgroundSample*0.25

		// Предотвращаем клиппинг
		if mixedSamples[i] > 1.0 {
			mixedSamples[i] = 1.0
		} else if mixedSamples[i] < -1.0 {
			mixedSamples[i] = -1.0
		}
	}

	// Создаем новый заголовок
	newHeader := main.Header

	// Конвертируем результат в нужный формат
	var mixedData []byte
	if main.Header.BitsPerSample == 32 && main.Header.AudioFormat == 3 {
		mixedData = float32SamplesToBytes(mixedSamples)
	} else {
		mixedData = int16SamplesToBytes(float32SamplesToInt16(mixedSamples))
	}

	newHeader.Subchunk2Size = uint32(len(mixedData))
	newHeader.ChunkSize = 36 + newHeader.Subchunk2Size

	// Создаем результирующий WAV файл
	var result bytes.Buffer
	binary.Write(&result, binary.LittleEndian, newHeader)
	result.Write(mixedData)

	return result.Bytes(), nil
}

// normalizeWavData нормализует параметры WAV файла
func normalizeWavData(wav *WAVData, targetSampleRate uint32, targetChannels uint16, targetBitsPerSample uint16) (*WAVData, error) {
	// Если параметры уже совпадают, возвращаем исходные данные
	if wav.Header.SampleRate == targetSampleRate &&
		wav.Header.NumChannels == targetChannels &&
		wav.Header.BitsPerSample == targetBitsPerSample {
		return wav, nil
	}

	// Конвертируем исходные данные в float32 семплы
	var sourceSamples []float32

	if wav.Header.BitsPerSample == 32 && wav.Header.AudioFormat == 3 {
		// 32-bit float
		sourceSamples = bytesToFloat32Samples(wav.Data)
	} else if wav.Header.BitsPerSample == 16 && wav.Header.AudioFormat == 1 {
		// 16-bit PCM
		sourceSamples = int16SamplesToFloat32(bytesToInt16Samples(wav.Data))
	} else {
		return nil, fmt.Errorf("неподдерживаемый формат аудио: %d бит, формат %d", wav.Header.BitsPerSample, wav.Header.AudioFormat)
	}

	var targetSamples []float32

	// Обработка каналов
	if wav.Header.NumChannels == 2 && targetChannels == 1 {
		// Конвертируем стерео в моно
		targetSamples = stereoToMonoFloat32(sourceSamples)
	} else if wav.Header.NumChannels == 1 && targetChannels == 2 {
		// Конвертируем моно в стерео
		targetSamples = monoToStereoFloat32(sourceSamples)
	} else if wav.Header.NumChannels == targetChannels {
		targetSamples = sourceSamples
	} else {
		return nil, fmt.Errorf("неподдерживаемое количество каналов: %d -> %d", wav.Header.NumChannels, targetChannels)
	}

	// Ресэмплинг (простая реализация)
	if wav.Header.SampleRate != targetSampleRate {
		targetSamples = resampleAudioFloat32(targetSamples, wav.Header.SampleRate, targetSampleRate, targetChannels)
	}

	// Создаем новый заголовок
	newHeader := wav.Header
	newHeader.SampleRate = targetSampleRate
	newHeader.NumChannels = targetChannels
	newHeader.BitsPerSample = targetBitsPerSample
	newHeader.ByteRate = targetSampleRate * uint32(targetChannels) * uint32(targetBitsPerSample/8)
	newHeader.BlockAlign = targetChannels * (targetBitsPerSample / 8)

	// Конвертируем в нужный формат
	var newData []byte
	if targetBitsPerSample == 32 {
		newHeader.AudioFormat = 3 // IEEE float
		newData = float32SamplesToBytes(targetSamples)
	} else if targetBitsPerSample == 16 {
		newHeader.AudioFormat = 1 // PCM
		newData = int16SamplesToBytes(float32SamplesToInt16(targetSamples))
	} else {
		return nil, fmt.Errorf("неподдерживаемая целевая битность: %d", targetBitsPerSample)
	}

	newHeader.Subchunk2Size = uint32(len(newData))
	newHeader.ChunkSize = 36 + newHeader.Subchunk2Size

	return &WAVData{
		Header: newHeader,
		Data:   newData,
	}, nil
}

// bytesToFloat32Samples конвертирует байты в 32-битные float семплы
func bytesToFloat32Samples(data []byte) []float32 {
	samples := make([]float32, len(data)/4)
	for i := 0; i < len(samples); i++ {
		bits := binary.LittleEndian.Uint32(data[i*4 : i*4+4])
		samples[i] = math.Float32frombits(bits)
	}
	return samples
}

// float32SamplesToBytes конвертирует 32-битные float семплы в байты
func float32SamplesToBytes(samples []float32) []byte {
	data := make([]byte, len(samples)*4)
	for i, sample := range samples {
		bits := math.Float32bits(sample)
		binary.LittleEndian.PutUint32(data[i*4:i*4+4], bits)
	}
	return data
}

// int16SamplesToFloat32 конвертирует 16-битные семплы в float32
func int16SamplesToFloat32(samples []int16) []float32 {
	floatSamples := make([]float32, len(samples))
	for i, sample := range samples {
		floatSamples[i] = float32(sample) / 32768.0
	}
	return floatSamples
}

// float32SamplesToInt16 конвертирует float32 семплы в 16-битные
func float32SamplesToInt16(samples []float32) []int16 {
	int16Samples := make([]int16, len(samples))
	for i, sample := range samples {
		// Ограничиваем значение в диапазоне [-1.0, 1.0]
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		int16Samples[i] = int16(sample * 32767.0)
	}
	return int16Samples
}

// stereoToMonoFloat32 конвертирует стерео в моно (float32)
func stereoToMonoFloat32(stereoSamples []float32) []float32 {
	monoSamples := make([]float32, len(stereoSamples)/2)
	for i := 0; i < len(monoSamples); i++ {
		// Берем среднее значение левого и правого каналов
		left := stereoSamples[i*2]
		right := stereoSamples[i*2+1]
		monoSamples[i] = (left + right) / 2.0
	}
	return monoSamples
}

// monoToStereoFloat32 конвертирует моно в стерео (float32)
func monoToStereoFloat32(monoSamples []float32) []float32 {
	stereoSamples := make([]float32, len(monoSamples)*2)
	for i, sample := range monoSamples {
		stereoSamples[i*2] = sample   // левый канал
		stereoSamples[i*2+1] = sample // правый канал
	}
	return stereoSamples
}

// resampleAudioFloat32 выполняет простой ресэмплинг (float32)
func resampleAudioFloat32(samples []float32, sourceRate, targetRate uint32, channels uint16) []float32 {
	if sourceRate == targetRate {
		return samples
	}

	ratio := float64(sourceRate) / float64(targetRate)
	sourceSamplesPerChannel := len(samples) / int(channels)
	targetSamplesPerChannel := int(float64(sourceSamplesPerChannel) / ratio)
	targetSamples := make([]float32, targetSamplesPerChannel*int(channels))

	for i := 0; i < targetSamplesPerChannel; i++ {
		sourceIndex := int(float64(i) * ratio)
		if sourceIndex >= sourceSamplesPerChannel {
			sourceIndex = sourceSamplesPerChannel - 1
		}

		for ch := 0; ch < int(channels); ch++ {
			targetSamples[i*int(channels)+ch] = samples[sourceIndex*int(channels)+ch]
		}
	}

	return targetSamples
}

// bytesToInt16Samples конвертирует байты в 16-битные семплы
func bytesToInt16Samples(data []byte) []int16 {
	samples := make([]int16, len(data)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}
	return samples
}

// int16SamplesToBytes конвертирует 16-битные семплы в байты
func int16SamplesToBytes(samples []int16) []byte {
	data := make([]byte, len(samples)*2)
	for i, sample := range samples {
		binary.LittleEndian.PutUint16(data[i*2:i*2+2], uint16(sample))
	}
	return data
}
