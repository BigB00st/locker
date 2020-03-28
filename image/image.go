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
	"gitlab.com/amit-yuval/locker/utils"
)

// define error for missing image
type ImageMissingError struct {
	msg string // description of error
}

func (e *ImageMissingError) Error() string { return e.msg }

func RemoveImage(imageName string) error {
	imagesMap, err := getImagesMap()
	if err != nil {
		return err
	}
	if _, ok := imagesMap[imageName]; !ok {
		return fmt.Errorf("image %s not found", imageName)
	}
	delete(imagesMap, imageName)
	if err := updateImagesJson(imagesMap); err != nil {
		return err
	}
	imageDir := filepath.Join(ImagesDir, imageName)
	return os.RemoveAll(imageDir)
}

func createOverlayDirs(basedir string) error {
	for _, d := range []string{work, upper, Merged} {
		if err := os.Mkdir(filepath.Join(basedir, d), 0744); err != nil {
			return errors.Wrapf(err, "failed to create directory %s", d)
		}
	}
	return nil
}

func Cleanup(imageName string) {
	imageDir := filepath.Join(ImagesDir, imageName)
	MergedDir := filepath.Join(imageDir, Merged)

	syscall.Unmount(MergedDir, 0)

	os.RemoveAll(MergedDir)
	os.RemoveAll(filepath.Join(imageDir, work))
	os.RemoveAll(filepath.Join(imageDir, upper))
}

func MountImage(imageName string) error {
	layerList, err := getLayerList(imageName)
	if err != nil {
		if _, ok := err.(*ImageMissingError); ok { // image not found locally
			fmt.Printf("Unable to find image %s locally\n", imageName)
			if err := PullImage(imageName); err != nil {
				return err
			}
			layerList, _ = getLayerList(imageName)
		}
	}
	if err := createOverlayDirs(filepath.Join(ImagesDir, imageName)); err != nil {
		return err
	}
	if err := mountLayers(layerList); err != nil {
		return err
	}
	return nil
}

func mountLayers(layerList []string) error {
	imageDir := filepath.Dir(layerList[0])
	opts := fmt.Sprintf("index=off,lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(layerList, ":"), filepath.Join(imageDir, upper), filepath.Join(imageDir, work))
	if err := syscall.Mount("overlay", filepath.Join(imageDir, Merged), "overlay", 0, opts); err != nil {
		return errors.Wrap(err, "unable to mount image")
	}
	return nil
}

func getLayerList(imageName string) ([]string, error) {
	imagesMap, err := getImagesMap()
	if err != nil {
		return nil, err
	}
	layerList, ok := imagesMap[imageName]
	if !ok {
		return nil, &ImageMissingError{}
	}

	return layerList, nil
}

func getImagesMap() (map[string][]string, error) {
	jsonFile, err := ioutil.ReadFile(imagesJsonFile)
	if err != nil {
		return nil, err
	}
	imagesMap := make(map[string][]string)
	if err := json.Unmarshal(jsonFile, &imagesMap); err != nil {
		return nil, errors.Wrap(err, "couldn't load images map from json file")
	}
	return imagesMap, nil
}

func updateImagesJson(data map[string][]string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "couldn't marshal json data")
	}
	if err := ioutil.WriteFile(imagesJsonFile, jsonData, 0644); err != nil {
		return errors.Wrap(err, "couldn't write to images json file")
	}
	return nil
}

func getImageConfig(imageName string) (map[string]interface{}, error) {
	jsonFile, err := ioutil.ReadFile(filepath.Join(ImagesDir, imageName, ConfigFile))
	if err != nil {
		return nil, err
	}
	imageConfig := make(map[string]interface{})
	if err := json.Unmarshal(jsonFile, &imageConfig); err != nil {
		return nil, errors.Wrap(err, "couldn't load config from json file")
	}
	return imageConfig["config"].(map[string]interface{}), nil
}

// function returns env, cmdList from config file
func ReadConfigFile(imageName string) ([]string, []string, error) {
	imageConfig, err := getImageConfig(imageName)
	if err != nil {
		return nil, nil, err
	}
	env := utils.InterfaceArrToStrArr(imageConfig["Env"].([]interface{}))
	cmd := utils.InterfaceArrToStrArr(imageConfig["Cmd"].([]interface{}))
	return cmd, env, nil
}
