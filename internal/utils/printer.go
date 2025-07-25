package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func PrintFile(printerName, filePath string) error {
	var cmd *exec.Cmd

	// Проверка доступности принтера
	var checkCmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		checkCmd = exec.Command("powershell", "-Command", fmt.Sprintf("Get-Printer -Name '%s'", printerName))
	case "linux", "darwin":
		checkCmd = exec.Command("lpstat", "-p", printerName)
	default:
		return fmt.Errorf("неподдерживаемая операционная система: %s", runtime.GOOS)
	}

	checkOutput, checkErr := checkCmd.CombinedOutput()
	if checkErr != nil {
		return fmt.Errorf("принтер '%s' недоступен или не найден: %v\nВывод команды: %s", printerName, checkErr, string(checkOutput))
	}

	// Печать файла
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("mspaint", "/pt", filePath, printerName)
	case "linux", "darwin":
		cmd = exec.Command("lp", "-d", printerName, filePath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка при печати файла: %v\nВывод команды: %s", err, string(output))
	}
	return nil
}
