package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/amit-yuval/locker/internal/utils"

	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

type ImageConfig struct {
	Dir string
}

// define error for missing image
type ImageMissingError struct {
	msg string // description of error
}

func (e *ImageMissingError) Error() string { return e.msg }

// MountImage mounts requested image, pulls image if not found locally
func MountImage(imageName string) (*ImageConfig, error) {
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
	baseDir, err := ioutil.TempDir(filepath.Join(imagesDir, imageName), "cntr-")
	if err != nil {
		return nil, errors.Wrap(err, "error creating base directory for container")
	}
	imageConfig := &ImageConfig{
		Dir: baseDir,
	}
	if err := createOverlayDirs(baseDir); err != nil {
		return nil, err
	}
	if err := mountLayers(baseDir, layerList); err != nil {
		return nil, err
	}
	return imageConfig, nil
}

// RemoveImage deletes content of image, updates images data file
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
	imageDir := filepath.Join(imagesDir, imageName)
	return os.RemoveAll(imageDir)
}

// createOverlayDirs creates neccesary directories for overlay2 mount
func createOverlayDirs(baseDir string) error {
	for _, d := range []string{work, upper, Merged} {
		if err := os.Mkdir(filepath.Join(baseDir, d), 0744); err != nil {
			return errors.Wrapf(err, "failed to create directory %s", d)
		}
	}
	return nil
}

// Cleanup unmounts image, removes changes
func (c *ImageConfig) Cleanup() {
	unix.Unmount(filepath.Join(c.Dir, Merged), 0)
	os.RemoveAll(c.Dir)
}

// ListImages returns a string containing list of local images, and data about them
func ListImages() (string, error) {
	imagesMap, err := getImagesMap()
	if err != nil {
		return "", err
	}
	ret := utils.Pad(lsPrintPad, " ", "NAME", "SIZE") + "\n"
	for k, _ := range imagesMap {
		du, err := utils.DirSize(filepath.Join(imagesDir, k))
		if err != nil {
			return "", errors.Wrap(err, "couldn't get disk usage of directory")
		}
		ret += utils.Pad(lsPrintPad, " ", k, bytefmt.ByteSize(uint64(du))) + "\n"
	}
	return ret, nil
}

// mountLayers mounts given layers of image
func mountLayers(baseDir string, layerList []string) error {
	opts := fmt.Sprintf("index=off,lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(layerList, ":"), filepath.Join(baseDir, upper), filepath.Join(baseDir, work))
	if err := unix.Mount("overlay", filepath.Join(baseDir, Merged), "overlay", 0, opts); err != nil {
		return errors.Wrap(err, "unable to mount image")
	}
	return nil
}

// getLayerList returns list of layers of image
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

// getImagesMap returns a map of images to layers
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

// updateImagesJson updates images data file with given map
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

// getImageConfig gets config for requested image
func getImageConfig(imageName string) (map[string]interface{}, error) {
	jsonFile, err := ioutil.ReadFile(filepath.Join(imagesDir, imageName, ConfigFile))
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
