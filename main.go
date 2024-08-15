package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"net/url"
)

func main() {
	a := app.New()
	w := a.NewWindow("模型上传")
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
	modelPathEntry.SetPlaceHolder("输入模型文件路径")

	uploadPathEntry := widget.NewEntry()
	uploadPathEntry.MultiLine = true
	uploadPathEntry.Wrapping = fyne.TextWrapBreak
	uploadPathEntry.SetPlaceHolder("输入上传路径")

	outputInfoEntry := widget.NewEntry()
	outputInfoEntry.MultiLine = true
	outputInfoEntry.Wrapping = fyne.TextWrapBreak
	outputInfoEntry.SetPlaceHolder("执行结果显示")

	ProgressBar := widget.NewProgressBar()
	ProgressInfo := widget.NewLabel("执行进度")

	ProgressVBox := container.NewVBox(ProgressInfo, ProgressBar)

	infoSplit := container.NewVSplit(ProgressVBox, outputInfoEntry)
	infoSplit.Offset = 0.3

	templateButton := widget.NewButtonWithIcon("模版文件", theme.MediaSkipNextIcon(), func() {

	})
	templateButton.Importance = widget.HighImportance

	stopButton := widget.NewButtonWithIcon("停止上传", theme.ContentClearIcon(), func() {

	})
	stopButton.Importance = widget.MediumImportance

	uploadButton := widget.NewButtonWithIcon("上传", theme.MediaSkipNextIcon(), func() {
		UpLoad(modelPathEntry, ProgressBar, outputInfoEntry, uploadPathEntry, w)
	})

	uploadButton.Importance = widget.HighImportance

	content := container.NewBorder(container.NewVBox(modelPathEntry, uploadPathEntry), container.NewGridWithColumns(3, templateButton, stopButton, uploadButton), nil, nil, infoSplit)

	return container.NewBorder(header, footer, nil, nil, content)
}
