package main

import (
	"context"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var jinja string = `
# 创建template.jinja文件,不同模型的文件内容不一样的
{{ (messages|selectattr('role', 'equalto', 'system')|list|last).content|trim if (messages|selectattr('role', 'equalto', 'system')|list) else '' }}

{%- for message in messages -%}
    {%- if message['role'] == 'user' -%}
        {{- '<reserved_106>' + message['content'] -}}
    {%- elif message['role'] == 'assistant' -%}
        {{- '<reserved_107>' + message['content'] -}}
    {%- endif -%}
{%- endfor -%}

{%- if add_generation_prompt and messages[-1]['role'] != 'assistant' -%}
    {{- '<reserved_107>' -}}
{% endif %}
`

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

func mkdirTemplateFile(modelPathEntry *widget.Entry, outputInfoEntry *widget.Entry, w fyne.Window) {
	modelPath := modelPathEntry.Text
	if !fileExist(modelPath) {
		ErrorPrint("模型文件不存在", w)
		return
	}
	if !fileExist(modelPath + "/template") {
		err := os.Mkdir(modelPath+"/template", 0755)
		if err != nil {
			ErrorPrint("模版文件目录创建失败", w)
			return
		}
		file, err := os.Create(modelPath + "/template/template.jinja")
		if err != nil {
			ErrorPrint("模版文件创建失败", w)
			return
		}
		defer file.Close()
		_, err = file.WriteString(jinja)
		if err != nil {
			ErrorPrint("模版文件写入失败", w)
			return
		}
	}
	InfoPrint(outputInfoEntry, "模型文件配置成功")
	return
}

func editTemplateFile(modelPathEntry *widget.Entry, outputInfoEntry *widget.Entry, w fyne.Window) {
	editorWindow := fyne.CurrentApp().NewWindow("模型模版文件编辑")
	file := modelPathEntry.Text + "/template/template.jinja"
	if !fileExist(file) {
		ErrorPrint("模板文件不存在", editorWindow)
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		ErrorPrint("模板文读取失败", editorWindow)
		return
	}

	// 创建一个文本框以显示和编辑文件内容
	textEditor := widget.NewMultiLineEntry()
	textEditor.SetText(string(content))

	// 保存文件按钮
	saveButton := widget.NewButton("Save File", func() {
		err := ioutil.WriteFile(file, []byte(textEditor.Text), 0644)
		if err != nil {
			ErrorPrint("模板文保存失败", editorWindow)
			return
		}
		dialog.ShowInformation("Saved", "文件保存成功!", editorWindow)
		editorWindow.Close()
		InfoPrint(outputInfoEntry, "修改成功")
	})
	editorWindow.SetContent(container.NewBorder(nil, saveButton, nil, nil, textEditor))
	editorWindow.Resize(fyne.NewSize(600, 400))
	editorWindow.Show()
}

var m string

func InfoPrint(outputInfoEntry *widget.Entry, massage string) {
	m += massage + "\n"
	outputInfoEntry.SetText(m)
	outputInfoEntry.Refresh()
}
func ErrorPrint(error string, w fyne.Window) {
	err := errors.New(error)
	dialog.ShowError(err, w)
}

func UpLoad(modelPathEntry *widget.Entry, ProgressBar *widget.ProgressBar, outputInfoEntry *widget.Entry,
	uploadPathEntry *widget.Entry, w fyne.Window, c context.Context) {

	if modelPathEntry.Text == "" {
		ErrorPrint("请输入正确的地址", w)
		return
	}

	modelPath := modelPathEntry.Text
	if !fileExist(modelPath) {
		ErrorPrint("模型文件不存在", w)
		return
	}
	ProgressBar.SetValue(0.1)
	InfoPrint(outputInfoEntry, "1.模型读取成功")

	if fileExist(modelPath + "/.git") {
		err := os.RemoveAll(modelPath + "/.git")
		if err != nil {
			ErrorPrint("删除.git文件错误", w)
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
			ErrorPrint("克隆镜像失败: "+string(output)+err.Error(), w)
			return
		}

	}
	ProgressBar.SetValue(0.33)
	InfoPrint(outputInfoEntry, "3.仓库下载成功")

	cmd := exec.Command("cp", "-r", repoPath+"/.git", modelPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		ErrorPrint("拷贝.git文件失败: "+string(output)+err.Error(), w)
		dialog.ShowError(err, w)
		return
	}
	InfoPrint(outputInfoEntry, "4.git文件拷贝成功")

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = modelPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		ErrorPrint("add失败: "+string(output)+err.Error(), w)
		return
	}
	ProgressBar.SetValue(0.5)
	InfoPrint(outputInfoEntry, "5.add")

	cmd = exec.Command("git", "commit", "-m", "first commit")
	cmd.Dir = modelPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		ErrorPrint("commit失败: "+string(output)+err.Error(), w)
		return
	}
	ProgressBar.SetValue(0.6)
	InfoPrint(outputInfoEntry, "6.commit")
	// 父协程通过此channel停止子进程的执行。
	ctx, Cancel := context.WithCancel(c)
	cmd = exec.CommandContext(ctx, "git", "push", "origin", "main")
	cmd.Dir = modelPath
	err = cmd.Start()
	if err != nil {
		ErrorPrint("push开始: "+string(output)+err.Error(), w)
		return
	}
	if !<-stopUpLoad {
		Cancel()
		if err = cmd.Process.Signal(os.Interrupt); err != nil {
			ErrorPrint("push停止失败: "+string(output)+err.Error(), w)
		}
		InfoPrint(outputInfoEntry, "push被停止成功")
		return
	}
	go func() {
		err = cmd.Wait()
		if err != nil {
			ErrorPrint("push失败: "+string(output)+err.Error(), w)
		} else {
			ErrorPrint("push成功: "+string(output)+err.Error(), w)
		}
	}()

	ProgressBar.SetValue(0.8)
	InfoPrint(outputInfoEntry, "7.push")
	InfoPrint(outputInfoEntry, "上传成功")
	ProgressBar.SetValue(1)
}
