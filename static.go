package PaimengBot

import "embed"

//go:embed static
var staticFiles embed.FS

// GetStaticFS 获取静态资源文件对象
func GetStaticFS() embed.FS {
	return staticFiles
}
