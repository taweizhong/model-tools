#!/bin/bash

# 脚本名称：compile_fyne.sh
# 功能：编译 Fyne 项目为 Windows (.exe) 和 macOS (.app)，输出到 app 目录

# 定义变量
APP_DIR="app"
ICON="icon.png"
APP_NAME="model-tools"
MAIN_FILE="main.go"

# 检查必要的工具是否安装
check_requirements() {
    echo "检查必要的工具..."

    # 检查 Go 是否安装
    if ! command -v go &> /dev/null; then
        echo "错误：未找到 Go 环境。请安装 Go 1.17 或更高版本。"
        exit 1
    fi

    # 检查 fyne 工具是否安装
    if ! command -v fyne &> /dev/null; then
        echo "错误：未找到 fyne 工具。请运行 'go install fyne.io/fyne/v2/cmd/fyne@latest' 安装。"
        exit 1
    fi

    # 检查主文件是否存在
    if [ ! -f "$MAIN_FILE" ]; then
        echo "错误：未找到主文件 $MAIN_FILE。"
        exit 1
    fi

    # 检查图标文件是否存在
    if [ ! -f "$ICON" ]; then
        echo "错误：未找到图标文件 $ICON。"
        exit 1
    fi
}

# 创建输出目录
setup_output_dir() {
    echo "创建输出目录 $APP_DIR..."
    if [ -d "$APP_DIR" ]; then
        rm -rf "$APP_DIR"
    fi
    mkdir -p "$APP_DIR"
}

# 编译为 Windows 可执行文件
compile_windows() {
    echo "正在编译 Windows 可执行文件..."

    # 检查 MinGW 编译器（假设使用 mingw-w64-x86_64-gcc）
    if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        echo "错误：未找到 MinGW 编译器。请安装 mingw-w64-x86_64-gcc（例如通过 MSYS2）。"
        exit 1
    fi

    # 在子 shell 中设置环境变量，避免影响其他函数
    bash -c "
        export GOOS=windows
        export GOARCH=amd64
        export CGO_ENABLED=1
        export CC=x86_64-w64-mingw32-gcc

        # 编译并打包
        fyne package -os windows -icon \"$ICON\" -name \"$APP_NAME\" -executable \"$APP_NAME.exe\"
        if [ \$? -ne 0 ]; then
            echo \"错误：Windows 编译失败。\"
            exit 1
        fi

        # 移动生成的文件到 app 目录
        mv \"$APP_NAME.exe\" \"$APP_DIR/$APP_NAME.exe\"
    "

    if [ $? -ne 0 ]; then
        exit 1
    fi
    echo "Windows 可执行文件已生成：$APP_DIR/$APP_NAME.exe"
}

# 编译为 macOS 应用
compile_macos() {
    echo "正在编译 macOS 应用..."

    # 检查 Xcode 命令行工具（macOS 编译需要）
    if ! command -v xcodebuild &> /dev/null; then
        echo "错误：未找到 Xcode 命令行工具。请运行 'xcode-select --install' 安装。"
        exit 1
    fi

    # 检查 clang 编译器
    if ! command -v clang &> /dev/null; then
        echo "错误：未找到 clang 编译器。请确保 Xcode 命令行工具已正确安装。"
        exit 1
    fi

    # 在子 shell 中设置环境变量，避免影响其他函数
    bash -c "
        export GOOS=darwin
        export GOARCH=arm64  # 使用 arm64 架构，适配 ARM Mac
        export CGO_ENABLED=1
        export CC=clang      # 显式使用 clang 编译器

        # 编译并打包
        fyne package -os darwin -icon \"$ICON\" -name \"$APP_NAME\"
        if [ \$? -ne 0 ]; then
            echo \"错误：macOS 编译失败。\"
            exit 1
        fi

        # 移动生成的文件到 app 目录
        mv \"$APP_NAME.app\" \"$APP_DIR/$APP_NAME.app\"
    "

    if [ $? -ne 0 ]; then
        exit 1
    fi
    echo "macOS 应用已生成：$APP_DIR/$APP_NAME.app"
}

# 主函数
main() {
    check_requirements
    setup_output_dir
    compile_windows
    compile_macos
    echo "编译完成！所有文件已输出到 $APP_DIR 目录。"
}

# 执行主函数
main