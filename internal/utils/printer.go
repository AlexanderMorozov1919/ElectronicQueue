package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func PrintFile(printerName, filePath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("mspaint", "/pt", filePath, printerName)
	case "linux", "darwin":
		cmd = exec.Command("lp", "-d", printerName, filePath)
	default:
		return fmt.Errorf("неподдерживаемая операционная система: %s", runtime.GOOS)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка при печати файла: %v\nВывод команды: %s", err, string(output))
	}
	return nil
}
