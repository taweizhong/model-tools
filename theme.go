package main

import (
	"fyne.io/fyne/v2"
)

type CustomTheme struct {
	fyne.Theme
}

func (t *CustomTheme) Padding() fyne.Size {
	return fyne.NewSize(50, 30)
}

//func (t *CustomTheme) SeparatorColor() color.Color {
//	return color.Transparent // 设置分割线颜色为透明
//}
