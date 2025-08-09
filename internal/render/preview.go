package render

import (
	"fmt"
	"log"
	"time"

	"github.com/saswatsam786/snapshell/internal/capture"

	"gocv.io/x/gocv"
)

func StartLocalPreview() {
	fmt.Println("Starting video ASCII preview...")

	// Check if webcam is available
	webcam, err := capture.OpenWebCam()
	if err != nil {
		log.Fatal("Error opening webcam:", err)
	}
	defer webcam.Close()

	fmt.Println("Webcam opened successfully!")
	fmt.Println("Press Ctrl+C to exit...")

	img := gocv.NewMat()
	defer img.Close()

	gray := gocv.NewMat()
	defer gray.Close()

	frameRate := time.Duration(100) * time.Millisecond

	for {
		img, err := webcam.ReadFrame()
		if err != nil {
			log.Println("Cannot read from webcam:", err)
			continue
		}

		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
		img.Close()

		asciiArt := ConvertFrameToASCII(gray)
		ClearTerminal()
		fmt.Print(asciiArt)

		time.Sleep(frameRate)
	}
}
