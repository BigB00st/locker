package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

const (
	imagesJsonFile = "images.json"
)

// define error for missing image
type ImageMissingError struct {
	msg string // description of error
}

func (e *ImageMissingError) Error() string { return e.msg }

type OverlayDirs struct {
	upper  string
	work   string
	merged string
	lower  []string
}

func NewOverlayDirs(lowerLayers []string) *OverlayDirs {
	s := new(OverlayDirs)
	s.Init(lowerLayers)
	return s
}

func (s *OverlayDirs) Init(lowerLayers []string) {
	s.upper = os.TempDir()
	s.work = os.TempDir()
	s.merged = os.TempDir()
	s.lower = lowerLayers
}

func (s *OverlayDirs) Destroy() {
	os.RemoveAll(s.upper)
	os.RemoveAll(s.work)
	os.RemoveAll(s.merged)
}

func RemoveImage(imageName string) error {
	imageDir := filepath.Join(imagesDir, imageName)
	return os.RemoveAll(imageDir)
}

func MountImage(imageName string) (*OverlayDirs, error) {
	layerList, err := getLayerList(imageName)
	if err != nil {
		if _, ok := err.(*ImageMissingError); ok { // image not found locally
			fmt.Printf("Unable to find image %s locally\n", imageName)
			if err := PullImage(imageName); err != nil {
				return nil, err
			}
			layerList, _ = getLayerList(imageName)
		}
	}
	s := NewOverlayDirs(layerList)
	if err := mountLayers(s); err != nil {
		return nil, err
	}
	return s, nil
}

func mountLayers(s *OverlayDirs) error {
	opts := fmt.Sprintf("index=off,lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(s.lower, ":"), s.upper, s.work)
	if err := syscall.Mount("overlay", s.merged, "overlay", 0, opts); err != nil {
		return errors.Wrap(err, "unable to mount image")
	}
	return nil
}

func getLayerList(imageName string) ([]string, error) {
	jsonFile, err := ioutil.ReadFile(imagesJsonFile)
	if err != nil {
		return nil, err
	}
	imagesMap := make(map[string][]string)

	if err := json.Unmarshal(jsonFile, &imagesMap); err != nil {
		return nil, err
	}
	layerList, ok := imagesMap[imageName]
	if !ok {
		return nil, &ImageMissingError{}
	}

	return layerList, nil
}
