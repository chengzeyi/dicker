package util

import (
	"io/ioutil"
	"math/rand"
	"path"
	"strings"
	"time"

	"github.com/chengzeyi/dicker/container"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenRandStrBytes(n int) string {
	const letterBytes = "1234567890"
	bSlice := make([]byte, n)
	for i := range bSlice {
		bSlice[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(bSlice)
}

func GetContainerPidByName(containerName string) (string, error) {
	containerInfoPath := path.Join(container.DEFAULT_INFO_DIR_PATH, containerName)
	configPath := path.Join(containerInfoPath, container.CONFIG_FILE_NAME)
	_, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	// TODO: json.Unmarshal
	return "", nil
}

// How to standardize this?
func GetEnvsByPid(pid string) []string {
	environPath :=	path.Join("/proc", pid, "environ")
	contentBytes, err := ioutil.ReadFile(environPath)
	if err != nil {
		log.Errorf("ReadFile() %s error %v", environPath, err)
		return nil
	}
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}
