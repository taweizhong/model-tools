#! /bin/bash

 read -p "输入已下载好的模型文件地址：" model_file
 read -p "输入上传的gitlab地址：" repoURL
 read -p "输入上传gitlab仓库的用户名：" repo_user
 read -p "输入上传gitlab仓库的密码：" repo_passwors
#export http_proxy=http://127.0.0.1:7890
#export https_proxy=$http_proxy

function is_exit() {
    if (( ${2} != 0 ))
    then
        echo "[ERROR]${1}"
        exit 1
    fi
}


if [[ -d $model_file ]]
then
    echo "1.模型文件存在"
else
    is_exit "模型文件不存在"
fi

if [[ -d ${model_file}/.git ]]
then
    if ! rm -rf ${model_file}/.git
    then
        is_exit "原模型.git删除失败"
    fi
    echo "2.原模型.git删除"
fi

function is_https() {
    if [[ ${repoURL: 0: 5} == "https" ]]
    then
      return 1
    else
      return 0
    fi
}

function format_url() {
    if (( $1 == 1 ))
    then
        echo https://${repo_user}:${repo_passwors}@${repoURL#*//}
    else
        echo http://${repo_user}:${repo_passwors}@${repoURL#*//}
    fi
}

function clone() {
    if ! git clone $RepoUrl
    then
        is_exit "仓库下载失败" ${PIPESTATUS[0]}
    fi
    echo "3.上传仓库下载完成"
}

is_https

RepoUrl=$(format_url $?)


clone

repo=${repoURL##*/}
repoFile=${repo%.git*}



if ! cp -r $(pwd)/${repoFile}/.git ${model_file}
then
    is_exit ".git复制失败" ${PIPESTATUS[0]}
fi
echo "4.上传仓库.git复制完成"

cd ${model_file}

if ! git add .
then
    is_exit "添加失败" ${PIPESTATUS[0]}
fi
echo "5.文件add完成"


if ! git commit -m "first commit"
then
   is_exit "commit 失败" ${PIPESTATUS[0]}
fi
echo "6.commit提交完成"

if ! git push $RepoUrl main
then
    is_exit "push失败" ${PIPESTATUS[0]}
fi
echo "7.推送完成"
