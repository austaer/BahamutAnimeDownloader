package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sqweek/dialog"
	"gopkg.in/ini.v1"
)

func envCheck() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	isErr("Path not found", err)
	os.Setenv("PATH", os.Getenv("PATH")+";"+dir)

	err = exec.Command("ffmpeg").Run()
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			isErr("ffmpeg not found -", errors.New("please download from official website first"))
		}
	}

	if _, err := os.Stat("./conf.ini"); os.IsNotExist(err) {
		createDefaultConfig()
	}
}

func (conf *config) loadConfig() {
	cfg, err := ini.Load("./conf.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	conf.targetPath = cfg.Section("paths").Key("target_dir").String()
	conf.tmpPath = cfg.Section("paths").Key("tmp_dir").String()

	if conf.targetPath == "" {
		directory, err := dialog.Directory().Title("影片存放路徑").Browse()

		if err != nil {
			isErr("Cannot found path - ", err)
		} else {
			cfg.Section("paths").Key("target_dir").SetValue(directory)
			conf.targetPath = directory
			cfg.SaveTo("./conf.ini")
		}
	}

	if conf.tmpPath == "" {
		directory, err := dialog.Directory().Title("m3u8 暫存檔路徑").Browse()

		if err != nil {
			isErr("Cannot found path - ", err)
		} else {
			cfg.Section("paths").Key("tmp_dir").SetValue(directory)
			conf.tmpPath = directory
			cfg.SaveTo("./conf.ini")
		}
	}
}

func createDefaultConfig() {
	conf := ini.Empty()
	paths, _ := conf.NewSection("paths")
	paths.NewKey("target_dir", "")
	paths.NewKey("tmp_dir", "")
	conf.SaveTo("./conf.ini")
}

func isErr(msg string, err error) {
	if err != nil {
		f, e := os.OpenFile("error.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if e != nil {
			log.Fatal(err.Error())
		}

		log.SetOutput(f)
		msg := msg + " " + err.Error()
		fmt.Println(msg)
		log.Fatal(msg)
	}
}

func (h *bahamut) getQuality() (string, string) {
	return h.res, h.quality
}

func (h *bahamut) request(action, url string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		isErr("Create "+action+" request failed - ", err)
	}

	req.Header.Add("cookie", h.cookie)
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	req.Header.Add("referer", "https://ani.gamer.com.tw/animeVideo.php?sn="+h.sn)
	req.Header.Add("origin", "https://ani.gamer.com.tw")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		isErr("Create "+action+" request failed - ", err)
	}

	return resp
}

func (h *bahamut) mergeChunk() {
	fmt.Println("Merging chunk file...please wait a moment...")
	output := h.conf.targetPath + "/"
	os.Mkdir(output, 0755)
	exec.Command("ffmpeg", "-allowed_extensions", "ALL", "-y", "-i", h.tmp+"/"+h.plName, "-c", "copy", output+h.title+".mp4").Run()
	fmt.Printf("File location: %s%s.mp4\n", output, h.sn)
}

func (h *bahamut) cleanUp() {
	// Delete a temporary directory
	os.RemoveAll(h.tmp)

	fmt.Println("Cleaned up.")
	fmt.Println(fmt.Sprintf("Total time: %ds", time.Now().Unix()-h.startTime))
}
