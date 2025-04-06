package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"model-upload/pkg"
	"net/url"
)

func MakeUpLoadUI(w fyne.Window, preferences fyne.Preferences) fyne.CanvasObject {
	// 头部
	header := canvas.NewText("模型上传", theme.PrimaryColor())
	header.TextSize = 32
	header.Alignment = fyne.TextAlignCenter

	// 底部
	u, _ := url.Parse("")
	footer := widget.NewHyperlinkWithStyle("github.com/taweizhong/model-tools.git", u, fyne.TextAlignCenter, fyne.TextStyle{})

	// 创建一个标签用于显示选择的文件夹路径
	pathLabel := widget.NewLabel("尚未选择文件夹")
	// 选择文件按钮
	selectFileButton := widget.NewButton("选择文件夹", func() {
		// 创建文件夹选择对话框
		folderDialog := dialog.NewFolderOpen(
			func(folder fyne.ListableURI, err error) {
				if err != nil {
					log.Println("选择文件夹时出错:", err)
					return
				}
				if folder == nil {
					log.Println("用户取消了选择")
					return
				}
				// 更新标签显示选择的文件夹路径
				pathLabel.SetText(folder.Path())
			},
			w,
		)
		// 显示对话框
		folderDialog.Show()
	})

	gitUrlEntry := widget.NewEntry()
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "模型文件", Widget: selectFileButton},
			{Text: "仓库地址", Widget: gitUrlEntry},
		},
		//OnSubmit: func() { // 可选，处理表单提交
		//	log.Println("模型文件：", pathLabel.Text)
		//	log.Println("仓库地址：", gitUrlEntry.Text)
		//},
	}

	// 我们也可以追加项目
	form.Append("文件地址", pathLabel)

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
		pkg.MkdirTemplateFile(pathLabel, outputInfoEntry, w)
	})
	templateButton.Importance = widget.HighImportance

	editTemplateButton := widget.NewButtonWithIcon("模版文件编辑", theme.DocumentCreateIcon(), func() {
		pkg.EditTemplateFile(pathLabel, outputInfoEntry, w)
	})

	stopButton := widget.NewButtonWithIcon("上传设置", theme.ContentClearIcon(), func() {
		pkg.UpLoadSetting(outputInfoEntry, preferences, w)

	})
	stopButton.Importance = widget.MediumImportance

	uploadButton := widget.NewButtonWithIcon("上传", theme.MediaSkipNextIcon(), func() {
		if !pkg.UpLoad(pathLabel, ProgressBar, outputInfoEntry, gitUrlEntry, preferences, w) {
			ProgressBar.SetValue(0)
			outputInfoEntry.SetText("本次上传失败")
		}
	})

	uploadButton.Importance = widget.HighImportance

	content := container.NewBorder(container.NewVBox(form), container.NewGridWithColumns(4, templateButton, editTemplateButton, stopButton, uploadButton), nil, nil, infoSplit)

	return container.NewBorder(header, footer, nil, nil, content)
}
