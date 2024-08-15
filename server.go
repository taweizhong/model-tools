package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func fileExist(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}
func getRepoNameFromURL(repoURL string) string {
	// 去掉协议部分 (例如 "http://", "https://", "git://")
	repoURL = strings.TrimPrefix(repoURL, "http://")
	repoURL = strings.TrimPrefix(repoURL, "https://")
	repoURL = strings.TrimPrefix(repoURL, "git://")

	// 去掉任何 URL 参数和查询部分
	repoURL = strings.Split(repoURL, "?")[0]
	repoURL = strings.Split(repoURL, "#")[0]

	// 提取路径部分
	path := strings.TrimSuffix(filepath.Base(repoURL), ".git")

	return path
}

var m string

func InfoPrint(outputInfoEntry *widget.Entry, massage string) {
	m += massage + "\n"
	outputInfoEntry.SetText(m)
	outputInfoEntry.Refresh()
}

func UpLoad(modelPathEntry *widget.Entry, ProgressBar *widget.ProgressBar, outputInfoEntry *widget.Entry, uploadPathEntry *widget.Entry, w fyne.Window) {
	if modelPathEntry.Text == "" {
		err := errors.New("请输入正确的地址")
		dialog.ShowError(err, w)
		return
	}

	modelPath := modelPathEntry.Text
	if !fileExist(modelPath) {
		err := errors.New("模型文件不存在")
		dialog.ShowError(err, w)
		return
	}
	ProgressBar.SetValue(0.1)
	InfoPrint(outputInfoEntry, "1.模型读取成功")

	if fileExist(modelPath + "/.git") {
		err := os.RemoveAll(modelPath + "/.git")
		if err != nil {
			err := errors.New("删除.git文件错误")
			dialog.ShowError(err, w)
			return
		}

	}
	ProgressBar.SetValue(0.2)
	InfoPrint(outputInfoEntry, "2.模型初始化成功")

	dir := filepath.Dir(modelPath)
	repoURL := uploadPathEntry.Text
	repoName := getRepoNameFromURL(repoURL)
	repoPath := dir + "/" + repoName
	if !fileExist(repoPath) {
		cmd := exec.Command("git", "clone", repoURL, dir+"/"+repoName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			err = errors.New("克隆镜像失败: " + string(output) + err.Error())
			dialog.ShowError(err, w)
			return
		}

	}
	ProgressBar.SetValue(0.33)
	InfoPrint(outputInfoEntry, "3.仓库下载成功")

	cmd := exec.Command("cp", "-r", repoPath+"/.git", modelPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.New("拷贝.git文件失败: " + string(output) + err.Error())
		dialog.ShowError(err, w)
		return
	}
	InfoPrint(outputInfoEntry, "4.git文件拷贝成功")

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = modelPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		err = errors.New("add失败: " + string(output) + err.Error())
		dialog.ShowError(err, w)
		return
	}
	ProgressBar.SetValue(0.5)
	InfoPrint(outputInfoEntry, "5.add")

	cmd = exec.Command("git", "commit", "-m", "first commit")
	cmd.Dir = modelPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		err = errors.New("commit失败: " + string(output) + err.Error())
		dialog.ShowError(err, w)
		return
	}
	ProgressBar.SetValue(0.6)
	InfoPrint(outputInfoEntry, "6.commit")

	cmd = exec.Command("git", "push", "origin", "main")
	cmd.Dir = modelPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		err = errors.New("push失败: " + string(output) + err.Error())
		dialog.ShowError(err, w)
		return
	}
	ProgressBar.SetValue(0.8)
	InfoPrint(outputInfoEntry, "7.push")
	InfoPrint(outputInfoEntry, "上传成功")
	ProgressBar.SetValue(1)
}
