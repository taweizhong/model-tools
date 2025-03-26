package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
)

func MakeIndexUI(w fyne.Window, preferences fyne.Preferences) fyne.CanvasObject {
	background := canvas.NewImageFromResource(resourceIndexJpg)
	background.FillMode = canvas.ImageFillStretch // 拉伸填充整个容器
	background.SetMinSize(fyne.NewSize(550, 600)) // 设置最小尺寸，确保图片可见

	homeContent := container.NewVBox(
	//homeLabel,
	)

	// 使用 container.NewStack 将背景图片和内容叠加
	homeTabContent := container.NewStack(
		background,                       // 背景图片在最底层
		container.NewCenter(homeContent), // 内容居中显示
	)
	return container.NewBorder(homeTabContent, nil, nil, nil)
}

// loadImage 从文件加载图片
func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file) // 假设图片是 JPEG 格式
	if err != nil {
		return nil, err
	}
	return img, nil
}

// loadImageWithOpacity 加载图片并设置透明度
func loadImageWithOpacity(path string, opacity float32) (image.Image, error) {
	// 加载原始图片
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return nil, err
	}

	// 获取图片的边界
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// 创建一个新的 RGBA 图片
	newImg := image.NewRGBA(bounds)

	// 遍历图片的每个像素，调整 Alpha 通道
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// 将 Alpha 通道值调整为原始值的 opacity 倍
			newAlpha := uint8(float32(a>>8) * opacity)
			newImg.SetRGBA(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: newAlpha,
			})
		}
	}

	return newImg, nil
}
