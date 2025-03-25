package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"math"
	"model-upload/pkg"
	"net/url"
	"strconv"
)

func MakeDataSplitUI(w fyne.Window, preferences fyne.Preferences) fyne.CanvasObject {
	// 头部
	header := canvas.NewText("数据集划分", theme.PrimaryColor())
	header.TextSize = 32
	header.Alignment = fyne.TextAlignCenter

	// 底部
	u, _ := url.Parse("")
	footer := widget.NewHyperlinkWithStyle("github.com/taweizhong/model-upload.git", u, fyne.TextAlignCenter, fyne.TextStyle{})

	// 创建一个标签用于显示选择的文件夹路径
	DataSetPathLabel := widget.NewLabel("尚未选择文件夹")
	// 选择文件按钮
	DataSetPathSelectFileButton := widget.NewButton("选择文件夹", func() {
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
				DataSetPathLabel.SetText(folder.Path())
			},
			w,
		)
		// 显示对话框
		folderDialog.Show()
	})
	ExportPathLabel := widget.NewLabel("尚未选择文件夹")
	exportPath := ""
	ExportPathSelectFileButton := widget.NewButton("选择文件夹", func() {
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
				ExportPathLabel.SetText(folder.Path())
				exportPath = folder.Path()
			},
			w,
		)
		// 显示对话框
		folderDialog.Show()
	})

	ExportFileLabel := widget.NewLabel("文件名称:")
	ExportFileEntry := widget.NewEntry()

	ExportFileEntry.OnChanged = func(s string) {
		ExportPathLabel.SetText(exportPath + "/" + s)
	}
	fileContent := container.New(&customSplitLayout{offset: 0.3}, ExportFileLabel, ExportFileEntry)
	exportContent := container.New(&customSplitLayout{offset: 0.5}, ExportPathSelectFileButton, fileContent)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "数据目录", Widget: DataSetPathSelectFileButton},
			{Text: "输出目录", Widget: exportContent},
		},
	}

	// 我们也可以追加项目
	form.Append("文件地址", DataSetPathLabel)
	form.Append("输出地址", ExportPathLabel)

	outputInfoEntry := widget.NewEntry()
	outputInfoEntry.MultiLine = true
	outputInfoEntry.Wrapping = fyne.TextWrapBreak
	outputInfoEntry.SetPlaceHolder("执行结果显示")

	ProgressBar := widget.NewProgressBar()
	ProgressInfo := widget.NewLabel("执行进度")

	ProgressVBox := container.NewVBox(ProgressInfo, ProgressBar)

	infoSplit := container.NewVSplit(ProgressVBox, outputInfoEntry)
	infoSplit.Offset = 0.3

	settingButton := widget.NewButtonWithIcon("划分设置", theme.ContentClearIcon(), func() {
		MakeSplitSettingUI(outputInfoEntry, preferences, w)

	})
	settingButton.Importance = widget.MediumImportance

	splitButton := widget.NewButtonWithIcon("执行", theme.MediaSkipNextIcon(), func() {
		if !pkg.Split(DataSetPathLabel, ProgressBar, outputInfoEntry, ExportPathLabel, preferences, w) {
			ProgressBar.SetValue(0)
			outputInfoEntry.SetText("本次划分失败")
		}
	})

	splitButton.Importance = widget.HighImportance

	showTrainFileButton := widget.NewButtonWithIcon("训练文件", theme.DocumentIcon(), func() {
		pkg.MkdirTrainFile(ExportPathLabel, outputInfoEntry, w)
	})
	showTrainFileButton.Importance = widget.HighImportance

	editTrainFileButton := widget.NewButtonWithIcon("训练文件编辑", theme.DocumentCreateIcon(), func() {
		pkg.EditTrainFile(ExportPathLabel, outputInfoEntry, w)
	})

	content := container.NewBorder(container.NewVBox(form), container.NewGridWithColumns(4, settingButton, splitButton, showTrainFileButton, editTrainFileButton), nil, nil, infoSplit)

	return container.NewBorder(header, footer, nil, nil, content)
}

// 自定义水平布局
type customSplitLayout struct {
	offset float32 // 比例，例如 0.3 表示左侧占 30%
}

func (l *customSplitLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		return
	}
	leftWidth := size.Width * l.offset
	rightWidth := size.Width * (1 - l.offset)

	// 左侧控件
	objects[0].Resize(fyne.NewSize(leftWidth, size.Height))
	objects[0].Move(fyne.NewPos(0, 0))

	// 右侧控件
	objects[1].Resize(fyne.NewSize(rightWidth, size.Height))
	objects[1].Move(fyne.NewPos(leftWidth, 0))
}

func (l *customSplitLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 2 {
		return fyne.NewSize(0, 0)
	}
	leftMin := objects[0].MinSize()
	rightMin := objects[1].MinSize()
	return fyne.NewSize(leftMin.Width+rightMin.Width, float32(math.Max(float64(leftMin.Height), float64(rightMin.Height))))
}

func MakeSplitSettingUI(outputInfoEntry *widget.Entry, preferences fyne.Preferences, w fyne.Window) {
	trainEntry := widget.NewEntry()
	trainEntry.SetPlaceHolder("70")
	valEntry := widget.NewEntry()
	valEntry.SetPlaceHolder("30")
	testEntry := widget.NewEntry()
	testEntry.SetPlaceHolder("20")

	slider1 := widget.NewSlider(0, 100) // 范围：0 到 100
	slider1.SetValue(70)                // 初始值：50
	slider1.Step = 1                    // 步长：1

	slider2 := widget.NewSlider(0, 100) // 范围：0 到 100
	slider2.SetValue(30)                // 初始值：50
	slider2.Step = 1                    // 步长：1

	slider3 := widget.NewSlider(0, 100) // 范围：0 到 100
	slider3.SetValue(20)                // 初始值：50
	slider3.Step = 1                    // 步长：1

	// 标志变量，用于避免循环调用
	var updating bool

	// 滑块 1 的变化逻辑：调整 slider2，slider3 不变
	slider1.OnChanged = func(value float64) {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		// 调整 slider2 使 slider1 + slider2 = 100
		slider2Value := 100 - value
		// 边界检查
		if slider2Value < 0 {
			slider2Value = 0
			value = 100
			slider1.SetValue(value)
		} else if slider2Value > 100 {
			slider2Value = 100
			value = 0
			slider1.SetValue(value)
		}
		slider2.SetValue(slider2Value)

		// 同步输入框
		trainEntry.SetText(fmt.Sprintf("%.0f", value))
		valEntry.SetText(fmt.Sprintf("%.0f", slider2.Value))
		testEntry.SetText(fmt.Sprintf("%.0f", slider3.Value))
	}

	// 滑块 2 的变化逻辑：调整 slider1，slider3 不变
	slider2.OnChanged = func(value float64) {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		// 调整 slider1 使 slider1 + slider2 = 100
		slider1Value := 100 - value
		// 边界检查
		if slider1Value < 0 {
			slider1Value = 0
			value = 100
			slider2.SetValue(value)
		} else if slider1Value > 100 {
			slider1Value = 100
			value = 0
			slider2.SetValue(value)
		}
		slider1.SetValue(slider1Value)

		// 同步输入框
		trainEntry.SetText(fmt.Sprintf("%.0f", slider1.Value))
		valEntry.SetText(fmt.Sprintf("%.0f", value))
		testEntry.SetText(fmt.Sprintf("%.0f", slider3.Value))
	}

	// 滑块 3 的变化逻辑：独立调整，不影响 slider1 和 slider2
	slider3.OnChanged = func(value float64) {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		// 边界检查
		if value < 0 {
			value = 0
			slider3.SetValue(value)
		} else if value > 100 {
			value = 100
			slider3.SetValue(value)
		}

		// 同步输入框
		trainEntry.SetText(fmt.Sprintf("%.0f", slider1.Value))
		valEntry.SetText(fmt.Sprintf("%.0f", slider2.Value))
		testEntry.SetText(fmt.Sprintf("%.0f", value))
	}

	// 输入框的变化逻辑
	trainEntry.OnChanged = func(value string) {
		if updating {
			return
		}
		if value == "" {
			return
		}
		if val, err := strconv.ParseFloat(value, 64); err == nil && val >= 0 && val <= 100 {
			slider1.SetValue(val)
		} else {
			trainEntry.SetText(fmt.Sprintf("%.0f", slider1.Value))
		}
	}

	valEntry.OnChanged = func(value string) {
		if updating {
			return
		}
		if value == "" {
			return
		}
		if val, err := strconv.ParseFloat(value, 64); err == nil && val >= 0 && val <= 100 {
			slider2.SetValue(val)
		} else {
			valEntry.SetText(fmt.Sprintf("%.0f", slider2.Value))
		}
	}

	testEntry.OnChanged = func(value string) {
		if updating {
			return
		}
		if value == "" {
			return
		}
		if val, err := strconv.ParseFloat(value, 64); err == nil && val >= 0 && val <= 100 {
			slider3.SetValue(val)
		} else {
			testEntry.SetText(fmt.Sprintf("%.0f", slider3.Value))
		}
	}

	// 使用自定义布局替代 HSplit
	content1 := container.New(&customSplitLayout{offset: 0.11}, trainEntry, slider1)
	content2 := container.New(&customSplitLayout{offset: 0.11}, valEntry, slider2)
	content3 := container.New(&customSplitLayout{offset: 0.11}, testEntry, slider3)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "训练集", Widget: content1},
			{Text: "验证集", Widget: content2},
			{Text: "测试集", Widget: content3},
		},
		OnSubmit: func() {
			dialog.ShowInformation("保存", "划分内容保存成功", w)
		},
		OnCancel: func() {
			dialog.ShowInformation("取消", "划分内容取消", w)
		},
	}
	train := ""
	val := ""
	test := ""
	formDialog := dialog.NewForm("设置", "保存", "取消", form.Items, func(b bool) {
		if b {
			form.OnSubmit()
			train = trainEntry.Text
			val = valEntry.Text
			test = testEntry.Text
			preferences.SetString("train", train)
			preferences.SetString("val", val)
			preferences.SetString("test", test)
			pkg.InfoPrint(outputInfoEntry, "划分配置设置成功")
		} else {
			form.OnCancel()
		}
	}, w)
	formDialog.Resize(fyne.NewSize(500, 300))
	formDialog.Show()
}
