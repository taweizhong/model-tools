package pkg

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"io"
	"io/ioutil"
	"math/rand"
	"model-upload/common"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var dataTemplate string = `
train: ../train/images
val: ../valid/images
test: ../test/images
`

func Split(DataPath *widget.Label, ProgressBar *widget.ProgressBar, outputInfoEntry *widget.Entry,
	ExportPath *widget.Label, preferences fyne.Preferences, w fyne.Window) bool {
	// 获取路径
	dataPath := DataPath.Text
	exportPath := ExportPath.Text

	// 检查路径是否有效
	if dataPath == "" || exportPath == "" {
		dialog.ShowError(fmt.Errorf("数据集路径或导出路径为空"), w)
		outputInfoEntry.SetText("错误：数据集路径或导出路径为空")
		return false
	}

	// 检查 DataPath 是否存在
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("数据集路径不存在: %s", dataPath), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：数据集路径不存在: %s", dataPath))
		return false
	}

	// 检查 image 和 label 文件夹是否存在
	imageDir := filepath.Join(dataPath, "images")
	labelDir := filepath.Join(dataPath, "labels")
	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("images 文件夹不存在: %s", imageDir), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：images 文件夹不存在: %s", imageDir))
		return false
	}
	if _, err := os.Stat(labelDir); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("labels 文件夹不存在: %s", labelDir), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：labels 文件夹不存在: %s", labelDir))
		return false
	}

	// 从 preferences 中读取划分比例
	value := preferences.StringWithFallback("train", "70")
	trainRatio, err := strconv.ParseFloat(value, 64)
	if err != nil {
		dialog.ShowError(fmt.Errorf("训练集比例无效: %v", err), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：训练集比例无效: %v", err))
		return false
	}
	valRatio, err := strconv.ParseFloat(preferences.StringWithFallback("val", "30"), 64)
	if err != nil {
		dialog.ShowError(fmt.Errorf("验证集比例无效: %v", err), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：验证集比例无效: %v", err))
		return false
	}
	testRatio, err := strconv.ParseFloat(preferences.StringWithFallback("test", "10"), 64)
	if err != nil {
		dialog.ShowError(fmt.Errorf("测试集比例无效: %v", err), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：测试集比例无效: %v", err))
		return false
	}

	// 确保 train 和 val 比例总和为 100
	if trainRatio+valRatio != 100 {
		dialog.ShowError(fmt.Errorf("训练集和验证集比例总和必须为 100，当前为 %.0f", trainRatio+valRatio), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：训练集和验证集比例总和必须为 100，当前为 %.0f", trainRatio+valRatio))
		return false
	}

	// 创建导出目录结构
	sets := []string{"train", "valid", "test"}
	for _, set := range sets {
		if err := os.MkdirAll(filepath.Join(exportPath, set, "images"), 0755); err != nil {
			dialog.ShowError(fmt.Errorf("创建目录失败: %s", err), w)
			outputInfoEntry.SetText(fmt.Sprintf("错误：创建目录失败: %s", err))
			return false
		}
		if err := os.MkdirAll(filepath.Join(exportPath, set, "labels"), 0755); err != nil {
			dialog.ShowError(fmt.Errorf("创建目录失败: %s", err), w)
			outputInfoEntry.SetText(fmt.Sprintf("错误：创建目录失败: %s", err))
			return false
		}
	}

	// 读取 image 文件夹中的所有图片文件
	imageFiles, err := filepath.Glob(filepath.Join(imageDir, "*"))
	if err != nil {
		dialog.ShowError(fmt.Errorf("读取图片文件失败: %s", err), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：读取图片文件失败: %s", err))
		return false
	}

	// 过滤出有效的图片文件（支持 jpg、png、jpeg）
	var validImageFiles []string
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	for _, file := range imageFiles {
		ext := strings.ToLower(filepath.Ext(file))
		if imageExtensions[ext] {
			validImageFiles = append(validImageFiles, file)
		}
	}

	if len(validImageFiles) == 0 {
		dialog.ShowError(fmt.Errorf("images 文件夹中没有有效的图片文件"), w)
		outputInfoEntry.SetText("错误：images 文件夹中没有有效的图片文件")
		return false
	}

	// 随机打乱图片文件列表
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(validImageFiles), func(i, j int) {
		validImageFiles[i], validImageFiles[j] = validImageFiles[j], validImageFiles[i]
	})

	// 计算测试集、训练集和验证集的图片数量
	totalImages := len(validImageFiles)
	testCount := int(float64(totalImages) * testRatio / 100)
	remainingImages := totalImages - testCount // 剩余图片用于 train 和 val
	trainCount := int(float64(remainingImages) * trainRatio / 100)
	valCount := remainingImages - trainCount // 确保 train + val = remainingImages

	// 调整数量以确保总数一致
	if trainCount+valCount != remainingImages {
		valCount = remainingImages - trainCount
	}

	// 随机抽取测试集
	testImages := validImageFiles[:testCount]

	// 从剩余图片中随机划分 train 和 val
	rand.Shuffle(len(validImageFiles), func(i, j int) {
		validImageFiles[i], validImageFiles[j] = validImageFiles[j], validImageFiles[i]
	})
	trainImages := validImageFiles[:trainCount]
	valImages := validImageFiles[trainCount:]

	// 构造图片到目标集的映射
	imageToSet := make(map[string]string)
	for _, imageFile := range testImages {
		imageToSet[imageFile] = "test"
	}
	for _, imageFile := range trainImages {
		imageToSet[imageFile] = "train"
	}
	for _, imageFile := range valImages {
		imageToSet[imageFile] = "valid"
	}

	// 初始化进度条
	ProgressBar.Min = 0
	ProgressBar.Max = float64(totalImages)
	ProgressBar.SetValue(0)

	// 划分图片和对应的 label 文件
	processed := 0
	for imageFile, targetSet := range imageToSet {
		// 拷贝图片文件
		targetImagePath := filepath.Join(exportPath, targetSet, "images", filepath.Base(imageFile))
		if err := copyFile(imageFile, targetImagePath); err != nil {
			dialog.ShowError(fmt.Errorf("拷贝图片文件失败: %s", err), w)
			outputInfoEntry.SetText(fmt.Sprintf("错误：拷贝图片文件失败: %s", err))
			return false
		}

		// 拷贝对应的 label 文件
		labelFile := filepath.Join(labelDir, strings.TrimSuffix(filepath.Base(imageFile), filepath.Ext(imageFile))+".txt")
		if _, err := os.Stat(labelFile); err == nil {
			targetLabelPath := filepath.Join(exportPath, targetSet, "labels", filepath.Base(labelFile))
			if err := copyFile(labelFile, targetLabelPath); err != nil {
				dialog.ShowError(fmt.Errorf("拷贝 label 文件失败: %s", err), w)
				outputInfoEntry.SetText(fmt.Sprintf("错误：拷贝 labels 文件失败: %s", err))
				return false
			}
		} else {
			outputInfoEntry.SetText(fmt.Sprintf("警告：未找到对应的 labels 文件: %s", labelFile))
		}

		// 更新进度条
		processed++
		ProgressBar.SetValue(float64(processed))
		outputInfoEntry.SetText(fmt.Sprintf("已处理 %d/%d 个文件", processed, totalImages))
	}

	// 拷贝 DataPath 中除 image 和 label 外的其他文件到 ExportPath 根目录
	dirEntries, err := os.ReadDir(dataPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("读取 DataPath 失败: %s", err), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：读取 DataPath 失败: %s", err))
		return false
	}

	for _, entry := range dirEntries {
		if entry.Name() != "images" && entry.Name() != "labels" {
			srcPath := filepath.Join(dataPath, entry.Name())
			dstPath := filepath.Join(exportPath, entry.Name())
			if entry.IsDir() {
				if err := copyDir(srcPath, dstPath); err != nil {
					dialog.ShowError(fmt.Errorf("拷贝目录失败: %s", err), w)
					outputInfoEntry.SetText(fmt.Sprintf("错误：拷贝目录失败: %s", err))
					return false
				}
			} else {
				if err := copyFile(srcPath, dstPath); err != nil {
					dialog.ShowError(fmt.Errorf("拷贝文件失败: %s", err), w)
					outputInfoEntry.SetText(fmt.Sprintf("错误：拷贝文件失败: %s", err))
					return false
				}
			}
		}
	}

	// 完成
	outputInfoEntry.SetText("数据集划分完成")
	dialog.ShowInformation("成功", "数据集划分完成", w)
	return true
}

// copyFile 拷贝单个文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyDir 递归拷贝目录
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func MkdirTrainFile(ExportPathLabel *widget.Label, outputInfoEntry *widget.Entry, w fyne.Window) {
	exportPath := ExportPathLabel.Text
	if !common.FileExist(exportPath) {
		ErrorPrint("文件不存在", w)
		return
	}
	if !common.FileExist(exportPath + "/data.yaml") {
		file, err := os.Create(exportPath + "/data.yaml")
		if err != nil {
			ErrorPrint("文件创建失败", w)
			return
		}
		defer file.Close()
	}
	if !common.FileExist(exportPath + "/classes.txt") {
		ErrorPrint("类文件不存在", w)
	}
	nc, names := setClasses(exportPath, outputInfoEntry, w)
	// 构造 data.yaml 内容
	yamlContent := fmt.Sprintf(`
train: ../train/images
val: ../valid/images
test: ../test/images
nc: %d
names: %s`, nc, names)

	// 写入 data.yaml 文件
	err := ioutil.WriteFile(exportPath+"/data.yaml", []byte(yamlContent), 0644)
	if err != nil {
		ErrorPrint("写入 data.yaml 失败: %s", w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：写入 data.yaml 失败: %s", err))
	}

	outputInfoEntry.SetText("成功生成 data.yaml 文件")
	dialog.ShowInformation("成功", "data.yaml 文件已生成", w)

	InfoPrint(outputInfoEntry, "训练文件配置成功")
	return
}
func EditTrainFile(ExportPath *widget.Label, outputInfoEntry *widget.Entry, w fyne.Window) {
	editorWindow := fyne.CurrentApp().NewWindow("训练文件编辑")
	file := ExportPath.Text + "/data.yaml"
	if !common.FileExist(file) {
		ErrorPrint("文件不存在", editorWindow)
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		ErrorPrint("文件读取失败", editorWindow)
		return
	}
	textEditor := widget.NewMultiLineEntry()
	textEditor.SetText(string(content))
	saveButton := widget.NewButton("Save File", func() {
		err := ioutil.WriteFile(file, []byte(textEditor.Text), 0644)
		if err != nil {
			ErrorPrint("训练文件保存失败", editorWindow)
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

func setClasses(exportPath string, outputInfoEntry *widget.Entry, w fyne.Window) (int, string) {
	// 读取 classes.txt 文件
	data, err := ioutil.ReadFile(exportPath + "/classes.txt")
	if err != nil {
		ErrorPrint("读取 classes.txt 失败: %s", w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：读取 classes.txt 失败: %s", err))
		return 0, ""
	}
	// 按行分割，获取类别名称
	classLines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var classes []string
	for _, line := range classLines {
		line = strings.TrimSpace(line)
		if line != "" {
			classes = append(classes, line)
		}
	}
	if len(classes) == 0 {
		dialog.ShowError(fmt.Errorf("classes.txt 文件为空或没有有效类别"), w)
		outputInfoEntry.SetText("错误：classes.txt 文件为空或没有有效类别")
		return 0, ""
	}

	// 计算类别数量
	nc := len(classes)
	// 构造 names 字段（格式为 ['class1', 'class2', ...]）
	names := "["
	for i, class := range classes {
		names += fmt.Sprintf("'%s'", class)
		if i < len(classes)-1 {
			names += ", "
		}
	}
	names += "]"
	return nc, names
}
