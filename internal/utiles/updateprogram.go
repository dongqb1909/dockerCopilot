package utiles

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/onlyLTY/dockerCopilot/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func UpdateProgram(ctx *svc.ServiceContext) error {
	githubProxy := os.Getenv("githubProxy")
	if githubProxy != "" {
		githubProxy = strings.TrimRight(githubProxy, "/") + "/"
	}
	versionURL := githubProxy + "https://raw.githubusercontent.com/onlyLTY/dockerCopilot/UGREEN/version"
	releaseBaseURL := githubProxy + "https://github.com/onlyLTY/dockerCopilot/releases/download"
	logx.Infof("versionURL: %s", versionURL)
	resp, err := http.Get(versionURL)
	if err != nil {
		logx.Info("没有获取到最新版本信息:", err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logx.Error("关闭resp.Body失败:", err)
		}
	}(resp.Body)

	versionData, err := io.ReadAll(resp.Body)
	logx.Infof("versionData: %s", versionData)
	if err != nil {
		logx.Info("没有获取到最新版本信息:", err)
		return nil
	}

	version := strings.TrimSpace(string(versionData))
	logx.Info("获取到最新版本：", version)
	// 2. 构造下载链接
	downloadURL := fmt.Sprintf("%s/%s/dockerCopilot-%s.tar.gz", releaseBaseURL, version, runtime.GOARCH)
	logx.Info("下载链接：", downloadURL)
	dest := "dockerCopilot.tar.gz"

	if err := downloadFile(downloadURL, dest); err != nil {
		logx.Error("下载失败:", err)
		return err
	}
	logx.Info("下载成功")

	if err := decompressTarGz(dest, "."); err != nil {
		logx.Info("解压缩失败:", err)
		return err
	}
	logx.Info("解压缩成功")

	return nil
}

func downloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logx.Error("关闭resp.Body失败:", err)
		}
	}(resp.Body)

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			logx.Error("关闭out失败:", err)
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}

func decompressTarGz(gzFilePath string, dest string) error {
	file, err := os.Open(gzFilePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logx.Error("关闭file失败:", err)
		}
	}(file)

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func(gzr *gzip.Reader) {
		err := gzr.Close()
		if err != nil {
			logx.Error("关闭gzr失败:", err)
		}
	}(gzr)

	tarReader := tar.NewReader(gzr)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := dest + "/" + header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				err := outFile.Close()
				if err != nil {
					return err
				}
				return err
			}
			err = outFile.Close()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("未知类型: %v in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}
