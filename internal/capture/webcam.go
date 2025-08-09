package capture

import (
	"errors"

	"gocv.io/x/gocv"
)

type WebCam struct {
	cam *gocv.VideoCapture
}

// OpenWebcam initializes the webcam at index 0
func OpenWebCam() (*WebCam, error) {
	cam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		return nil, err
	}

	return &WebCam{cam: cam}, nil
}

// ReadFrame returns a new frame from the webcam
func (w *WebCam) ReadFrame() (gocv.Mat, error) {
	img := gocv.NewMat()
	if ok := w.cam.Read(&img); !ok || img.Empty() {
		return gocv.Mat{}, errors.New("failed to read frame")
	}

	return img, nil
}

func (w *WebCam) Close() {
	w.cam.Close()
}

// SetProperty sets a webcam property
func (w *WebCam) SetProperty(prop gocv.VideoCaptureProperties, value float64) {
	w.cam.Set(prop, value)
}
