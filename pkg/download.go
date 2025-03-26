package pkg

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"model-upload/common"
	"os/exec"
)

func DownloadSetting(outputInfoEntry *widget.Entry, preferences fyne.Preferences, w fyne.Window) {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("模型平台用户名")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("模型平台密码或token")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "用户名", Widget: usernameEntry},
			{Text: "密码", Widget: passwordEntry},
		},
		OnSubmit: func() {
			dialog.ShowInformation("保存", "下载设置保存成功", w)
		},
		OnCancel: func() {
			dialog.ShowInformation("取消", "下载设置取消", w)
		},
	}
	username := ""
	password := ""
	formDialog := dialog.NewForm("设置", "保存", "取消", form.Items, func(b bool) {
		if b {
			form.OnSubmit()
			username = usernameEntry.Text
			password = passwordEntry.Text
			preferences.SetString("modelusername", username)
			preferences.SetString("modelpassword", password)
			InfoPrint(outputInfoEntry, "下载配置设置成功")
		} else {
			form.OnCancel()
		}
	}, w)
	formDialog.Resize(fyne.NewSize(400, 300))
	formDialog.Show()
}
func Download(pathLabel *widget.Label, ProgressBar *widget.ProgressBarInfinite, outputInfoEntry *widget.Entry,
	gitUrlEntry *widget.Entry, preferences fyne.Preferences, w fyne.Window) bool {
	if pathLabel.Text == "" {
		ErrorPrint("请输入正确的地址", w)
		return false
	}

	filePath := pathLabel.Text
	if !common.FileExist(filePath) {
		ErrorPrint("存储地址不存在", w)
		return false
	}
	// 下载地址
	repoURL := gitUrlEntry.Text
	// 仓库名称
	repoName := getRepoNameFromURL(repoURL)
	repoPath := filePath + "/" + repoName

	//username := preferences.String("modelusername")
	//password := preferences.String("modelpassword")
	//gitUrl := ""
	//if repoURL[0:5] == "https" {
	//	gitUrl = fmt.Sprintf("https://%s:%s@", username, password) + repoURL[8:]
	//} else {
	//	gitUrl = fmt.Sprintf("http://%s:%s@", username, password) + repoURL[7:]
	//}
	ProgressBar.Start()
	go func() {
		if !common.FileExist(repoPath) {
			cmd := exec.Command("git", "clone", repoURL, filePath+"/"+repoName)
			output, err := cmd.CombinedOutput()
			if err != nil {
				ErrorPrint("克隆镜像失败: "+string(output)+err.Error(), w)
			}
			InfoPrint(outputInfoEntry, string(output))
			InfoPrint(outputInfoEntry, "下载成功")
		} else {
			ErrorPrint("模型目录已经存在:", w)
		}
		ProgressBar.Stop()
	}()
	return true
}
