package client

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// DownloadToFile 下载文件，并返回其绝对路径
func DownloadToFile(filename, url string, tryTime int) error {
	c := NewHttpClient(&HttpOptions{TryTime: tryTime})
	return c.DownloadToFile(filename, url)
}

// DownloadToFile 下载文件，并返回其绝对路径
func (c *HttpClient) DownloadToFile(filename, url string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	reader, err := c.GetReader(url)
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(f, reader)
	return err
}

// DownloadToFile 下载文件，并返回其绝对路径
func (c *HttpClient) DownloadToFileHash(filename, url string) (string, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer f.Close()
	reader, err := c.GetReader(url)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	hasher := md5.New()
	var dst = io.MultiWriter(f, hasher)
	_, err = io.Copy(dst, reader)
	return hex.EncodeToString(hasher.Sum(nil)), err
}
