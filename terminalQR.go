package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/skip2/go-qrcode"
	"math"
	"math/rand"
	"os"
	"time"
)

var (
	rainbowMode   bool
	trueColorMode bool
	content       string
)

const (
	spread     = 3.0
	freq       = 0.1
	blackBlock = "\033[40m  \033[0m"
	whiteBlock = "\033[47m  \033[0m"
)

func parseFlags() {
	flag.BoolVar(&rainbowMode, "r", false, "Rainbow mode")
	flag.BoolVar(&trueColorMode, "t", false, "True color mode")
	flag.Parse()
	content = flag.Arg(0)
}

func detectTrueColorMode() bool {
	return os.Getenv("COLORTERM") == "truecolor"
}

func rainbow(freq, i float64) (int, int, int) {
	red := int(math.Sin(freq*i+0)*128 + 128)
	green := int(math.Sin(freq*i+2*math.Pi/3)*127 + 128)
	blue := int(math.Sin(freq*i+4*math.Pi/3)*127 + 128)
	return red, green, blue
}

func rgbTo256(red, green, blue int, content string) string {
	var gray bool
	grayPossible := true
	sep := 42.5
	for grayPossible {
		if float64(red) < sep || float64(green) < sep || float64(blue) < sep {
			gray = float64(red) < sep && float64(green) < sep && float64(blue) < sep
			grayPossible = false
		}
		sep += 42.5
	}
	if gray {
		return fmt.Sprintf("\033[48;5;%dm%s\033[0m", 232+int(math.Round((float64(red)+float64(green)+float64(blue))/33)), content)
	} else {
		return fmt.Sprintf("\033[48;5;%dm%s\033[0m", 16+(int(6*float64(red)/256)*36+int(6*float64(green)/256)*6+int(6*float64(blue)/256)*1), content)
	}
}

func rgbToTrueColor(red, green, blue int, content string) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm%s\033[0m", red, green, blue, content)
}

func main() {
	parseFlags()
	if content == "" {
		fmt.Fprintln(os.Stderr, "Content is empty")
		os.Exit(1)
	}
	fmt.Println("content:", content)
	if !trueColorMode {
		trueColorMode = detectTrueColorMode()
	}
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to generate QR code:", err.Error())
		os.Exit(1)
	}
	var buf bytes.Buffer
	qrBitmap := qr.Bitmap()
	rainbowRandomSeed := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(256)
	rainbowOffset := float64(0)
	for y := range qrBitmap {
		for x := range qrBitmap[y] {
			if qrBitmap[y][x] {
				// Foreground
				if rainbowMode {
					red, green, blue := rainbow(freq, float64(rainbowRandomSeed)+(rainbowOffset/spread))
					if trueColorMode {
						buf.WriteString(rgbToTrueColor(red, green, blue, " "))
					} else {
						buf.WriteString(rgbTo256(red, green, blue, " "))
					}
					rainbowOffset++
					red, green, blue = rainbow(freq, float64(rainbowRandomSeed)+(rainbowOffset/spread))
					if trueColorMode {
						buf.WriteString(rgbToTrueColor(red, green, blue, " "))
					} else {
						buf.WriteString(rgbTo256(red, green, blue, " "))
					}
				} else {
					buf.WriteString(blackBlock)
				}
			} else {
				// Background
				if trueColorMode {
					buf.WriteString(rgbToTrueColor(255, 255, 255, "  "))
				} else {
					buf.WriteString(whiteBlock)
				}
				if rainbowMode {
					rainbowOffset++
				}
			}
			if rainbowMode {
				rainbowOffset++
			}
		}
		if rainbowMode {
			rainbowOffset = 0
			rainbowRandomSeed++
		}
		buf.WriteString("\n")
	}
	fmt.Println(buf.String())
}
