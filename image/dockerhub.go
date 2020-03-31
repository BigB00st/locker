package image

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeclysm/extract"
	"github.com/pkg/errors"
)

func toJson(resp *http.Response) map[string]interface{} {
	ret := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&ret)
	return ret
}

func setHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func PullImage(imageName string) error {
	imageDir := filepath.Join(imagesDir, imageName)
	if _, err := os.Stat(imageDir); !os.IsNotExist(err) {
		return fmt.Errorf("Image %s exists", imageName)
	}

	repository := "library/" + imageName
	client := &http.Client{}
	authUrl := "https://auth.docker.io/token"
	regService := "registry.docker.io"
	regResp, err := http.Get(registry)
	if err != nil {
		return errors.Wrap(err, "error getting registry")
	}

	if regResp.StatusCode == 401 {
		authHeader := strings.Split(regResp.Header["Www-Authenticate"][0], `"`)
		authUrl = authHeader[authUrlIndex]
		if len(authHeader) > authHeaderIndex {
			regService = authHeader[authHeaderIndex]
		} else {
			regService = ""
		}
	}

	authResp, err := http.Get(fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", authUrl, regService, repository))
	if err != nil {
		errors.Wrapf(err, "error getting repository %s", repository)
	}

	accessToken := toJson(authResp)["token"].(string)
	authHead := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/vnd.docker.distribution.manifest.v2+json",
	}

	manifestReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/manifests/%s", registry, repository, "latest"), nil)
	if err != nil {
		return errors.Wrap(err, "error creating manifest request")
	}
	setHeaders(manifestReq, authHead)
	manifestResp, err := client.Do(manifestReq)
	if err != nil {
		return errors.Wrap(err, "error sending manifest request")
	}

	manifestBody := toJson(manifestResp)
	layers, ok := manifestBody["layers"]
	if !ok {
		return fmt.Errorf("Repository %s request invalid", repository)
	}
	config := manifestBody["config"].(map[string]interface{})["digest"].(string)
	confReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s/blobs/%s", registry, repository, config), nil)
	if err != nil {
		return errors.Wrap(err, "error creating config request")
	}
	setHeaders(confReq, authHead)
	confResp, err := client.Do(confReq)
	if err != nil {
		return errors.Wrap(err, "error receiving config request")
	}

	if err := os.Mkdir(imageDir, 0744); err != nil {
		return err
	}

	confFile, err := os.Create(filepath.Join(imageDir, ConfigFile))
	if err != nil {
		return errors.Wrap(err, "error creating config file")
	}
	defer confFile.Close()
	if _, err := io.Copy(confFile, confResp.Body); err != nil {
		return errors.Wrap(err, "couldn't write to config file")
	}

	var layerList []string
	parentId := ""
	for _, layer := range layers.([]interface{}) {
		layer := layer.(map[string]interface{})
		ublob := layer["digest"].(string)
		hash := sha256.New()
		hash.Write([]byte(parentId + ublob))
		fakeLayerId := hex.EncodeToString(hash.Sum(nil))
		layerDir := filepath.Join(imageDir, fakeLayerId)
		err := os.Mkdir(layerDir, 0744)
		if err != nil {
			return errors.Wrapf(err, "error creating layer %s directory", fakeLayerId)
		}

		fmt.Println("Pulling fs layer", fakeLayerId[:idPrintLen])
		req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/blobs/%s", registry, repository, ublob), nil)
		if err != nil {
			return errors.Wrap(err, "error creating blob request")
		}
		setHeaders(req, authHead)
		blobResp, err := client.Do(req)
		if err != nil {
			return errors.Wrap(err, "error sending blob request")
		}

		if err := extract.Gz(context.Background(), blobResp.Body, layerDir, nil); err != nil {
			return errors.Wrapf(err, "error extracting layer %s", fakeLayerId)
		}

		layerList = append(layerList, layerDir)

	}

	imagesMap, err := getImagesMap()
	if err != nil {
		return err
	}
	imagesMap[imageName] = layerList
	if err := updateImagesJson(imagesMap); err != nil {
		return err
	}

	return nil
}
