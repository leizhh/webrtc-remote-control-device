package client

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func controlHandler(msg []byte) string {
	resq := make(map[string]string)
	err := json.Unmarshal(msg, &resq)
	if err != nil {
		return "{\"result\":\"failure\",\"msg\":\"" + err.Error() + "\",\"timestamp\":" + strconv.FormatInt(time.Now().Unix(), 10) + "}"
	}

	if !isExist("./control/" + resq["device_id"]) {
		return "{\"result\":\"failure\",\"msg\":\"device_id not exist\",\"timestamp\":" + strconv.FormatInt(time.Now().Unix(), 10) + "}"
	}

	if !isExist("./control/" + resq["device_id"] + "/" + resq["operate"]) {
		return "{\"result\":\"failure\",\"msg\":\"operate not exist\",\"timestamp\":" + strconv.FormatInt(time.Now().Unix(), 10) + "}"
	}

	cmd := exec.Command("sh", "-c", "./control/"+resq["device_id"]+"/"+resq["operate"])
	out, err := cmd.Output()
	if err != nil {
		return "{\"result\":\"failure\",\"msg\":\"" + err.Error() + "\",\"timestamp\":" + strconv.FormatInt(time.Now().Unix(), 10) + "}"
	}
	return "{\"result\":\"success\",\"msg\":\"" + string(out) + "\",\"timestamp\":" + strconv.FormatInt(time.Now().Unix(), 10) + "}"
}

func isExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}
