package main

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"net/url"
)

var stopUpLoad = make(chan bool)

func main() {
	a := app.New()
	w := a.NewWindow("模型上传工具")
	w.SetContent(makeUI(w))
	w.Resize(fyne.NewSize(700, 600))
	w.ShowAndRun()
}

func makeUI(w fyne.Window) fyne.CanvasObject {
	header := canvas.NewText("模型上传", theme.PrimaryColor())
	header.TextSize = 42
	header.Alignment = fyne.TextAlignCenter

	u, _ := url.Parse("")
	footer := widget.NewHyperlinkWithStyle("github.com/taweizhong/model-upload.git", u, fyne.TextAlignCenter, fyne.TextStyle{})

	modelPathEntry := widget.NewEntry()
	modelPathEntry.MultiLine = true
	modelPathEntry.Wrapping = fyne.TextWrapBreak
	modelPathEntry.SetPlaceHolder("输入模型文件路径(已经下载的模型文件地址)\n示例：/data/models/glm-4b")

	uploadPathEntry := widget.NewEntry()
	uploadPathEntry.MultiLine = true
	uploadPathEntry.Wrapping = fyne.TextWrapBreak
	uploadPathEntry.SetPlaceHolder("输入上传路径(仓库地址)")

	outputInfoEntry := widget.NewEntry()
	outputInfoEntry.MultiLine = true
	outputInfoEntry.Wrapping = fyne.TextWrapBreak
	outputInfoEntry.SetPlaceHolder("执行结果显示")

	ProgressBar := widget.NewProgressBar()
	ProgressInfo := widget.NewLabel("执行进度")

	ProgressVBox := container.NewVBox(ProgressInfo, ProgressBar)

	infoSplit := container.NewVSplit(ProgressVBox, outputInfoEntry)
	infoSplit.Offset = 0.3

	templateButton := widget.NewButtonWithIcon("模版文件", theme.DocumentIcon(), func() {
		mkdirTemplateFile(modelPathEntry, outputInfoEntry, w)
	})
	templateButton.Importance = widget.HighImportance

	editTemplateButton := widget.NewButtonWithIcon("模版文件编辑", theme.DocumentCreateIcon(), func() {
		editTemplateFile(modelPathEntry, outputInfoEntry, w)
	})

	stopButton := widget.NewButtonWithIcon("停止上传", theme.ContentClearIcon(), func() {
		stopUpLoad <- false
	})
	stopButton.Importance = widget.MediumImportance

	uploadButton := widget.NewButtonWithIcon("上传", theme.MediaSkipNextIcon(), func() {
		m += "--------------------------- \n"
		ProgressBar.SetValue(0)
		go UpLoad(modelPathEntry, ProgressBar, outputInfoEntry, uploadPathEntry, w, context.Background())
	})

	uploadButton.Importance = widget.HighImportance

	content := container.NewBorder(container.NewVBox(modelPathEntry, uploadPathEntry), container.NewGridWithColumns(4, templateButton, editTemplateButton, stopButton, uploadButton), nil, nil, infoSplit)

	return container.NewBorder(header, footer, nil, nil, content)
}
