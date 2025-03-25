package pkg

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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
		dialog.ShowError(fmt.Errorf("image 文件夹不存在: %s", imageDir), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：image 文件夹不存在: %s", imageDir))
		return false
	}
	if _, err := os.Stat(labelDir); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("label 文件夹不存在: %s", labelDir), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：label 文件夹不存在: %s", labelDir))
		return false
	}

	// 从 preferences 中读取划分比例
	trainRatio, err := strconv.ParseFloat(preferences.StringWithFallback("train", "70"), 64)
	if err != nil {
		dialog.ShowError(fmt.Errorf("训练集比例无效: %v", err), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：训练集比例无效: %v", err))
		return false
	}
	valRatio, err := strconv.ParseFloat(preferences.StringWithFallback("val", "20"), 64)
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

	// 确保比例总和为 100（仅训练集和验证集需要总和为 100，测试集独立）
	if trainRatio+valRatio != 100 {
		dialog.ShowError(fmt.Errorf("训练集和验证集比例总和必须为 100，当前为 %.0f", trainRatio+valRatio), w)
		outputInfoEntry.SetText(fmt.Sprintf("错误：训练集和验证集比例总和必须为 100，当前为 %.0f", trainRatio+valRatio))
		return false
	}

	// 创建导出目录结构
	sets := []string{"train", "valid", "test"}
	for _, set := range sets {
		// 创建 image 和 label 子目录
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

	// 过滤出有效的图片文件（假设支持 jpg、png、jpeg）
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
		dialog.ShowError(fmt.Errorf("image 文件夹中没有有效的图片文件"), w)
		outputInfoEntry.SetText("错误：image 文件夹中没有有效的图片文件")
		return false
	}

	// 随机打乱图片文件列表
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(validImageFiles), func(i, j int) {
		validImageFiles[i], validImageFiles[j] = validImageFiles[j], validImageFiles[i]
	})

	// 计算训练集、验证集、测试集的图片数量
	totalImages := len(validImageFiles)
	trainCount := int(float64(totalImages) * trainRatio / 100)
	valCount := int(float64(totalImages) * valRatio / 100)
	testCount := int(float64(totalImages) * testRatio / 100)

	// 调整数量以确保总数一致
	if trainCount+valCount+testCount != totalImages {
		testCount = totalImages - trainCount - valCount
	}

	// 初始化进度条
	ProgressBar.Min = 0
	ProgressBar.Max = float64(totalImages)
	ProgressBar.SetValue(0)

	// 划分图片和对应的 label 文件
	processed := 0
	for i, imageFile := range validImageFiles {
		var targetSet string
		if i < trainCount {
			targetSet = "train"
		} else if i < trainCount+valCount {
			targetSet = "valid"
		} else {
			targetSet = "test"
		}

		// 拷贝图片文件
		targetImagePath := filepath.Join(exportPath, targetSet, "images", filepath.Base(imageFile))
		if err := copyFile(imageFile, targetImagePath); err != nil {
			dialog.ShowError(fmt.Errorf("拷贝图片文件失败: %s", err), w)
			outputInfoEntry.SetText(fmt.Sprintf("错误：拷贝图片文件失败: %s", err))
			return false
		}

		// 拷贝对应的 label 文件（假设 label 文件名与图片文件名相同，仅扩展名不同）
		labelFile := filepath.Join(labelDir, strings.TrimSuffix(filepath.Base(imageFile), filepath.Ext(imageFile))+".txt")
		if _, err := os.Stat(labelFile); err == nil {
			targetLabelPath := filepath.Join(exportPath, targetSet, "labels", filepath.Base(labelFile))
			if err := copyFile(labelFile, targetLabelPath); err != nil {
				dialog.ShowError(fmt.Errorf("拷贝 label 文件失败: %s", err), w)
				outputInfoEntry.SetText(fmt.Sprintf("错误：拷贝 label 文件失败: %s", err))
				return false
			}
		} else {
			outputInfoEntry.SetText(fmt.Sprintf("警告：未找到对应的 label 文件: %s", labelFile))
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
