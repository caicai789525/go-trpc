package ui

import (
	"fmt"
	"github.com/fatih/color"
)

func ShowLogo() {
	fmt.Println(color.HiCyanString("╔══════════════════════════════════════╗"))
	fmt.Println(color.HiCyanString("║        tRPC File Transfer CLI        ║"))
	fmt.Println(color.HiCyanString("╚══════════════════════════════════════╝"))
	fmt.Println()
}

func PrintInfo(msg ...interface{}) {
	color.New(color.FgGreen).Println("[INFO]", fmt.Sprint(msg...))
}

func PrintError(msg ...interface{}) {
	color.New(color.FgRed).Println("[ERROR]", fmt.Sprint(msg...))
}
