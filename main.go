package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"model-upload/ui"
)

func main() {
	a := app.New()
	a.Settings().SetTheme(&CustomTheme{Theme: theme.DefaultTheme()})
	preferences := a.Preferences()
	w := a.NewWindow("模型上传工具")

	tabs := container.NewAppTabs(
		container.NewTabItem("首页", ui.MakeIndexUI(w, preferences)),
		container.NewTabItem("模型下载", ui.MakeDownloadUI(w, preferences)),
		container.NewTabItem("模型上传", ui.MakeUpLoadUI(w, preferences)),
		container.NewTabItem("数据集划分", ui.MakeDataSplitUI(w, preferences)),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	w.SetContent(tabs)
	w.Resize(fyne.NewSize(550, 600))
	w.SetFixedSize(true)
	w.ShowAndRun()
}
