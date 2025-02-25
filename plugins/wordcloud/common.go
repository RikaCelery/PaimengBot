package wordcloud

import (
	"bytes"
	"image/color"
	"image/jpeg"
	"sort"

	"github.com/RicheyJang/PaimengBot/utils/consts"
	"github.com/go-ego/gse"
	"github.com/psykhi/wordclouds"
)

var (
	seg gse.Segmenter
)

func GetWordCloud(text string) ([]byte, error) {
	splites := seg.Cut(text, true)
	splites = seg.Stop(splites)
	wordCounts := make(map[string]int)
	for _, word := range splites {
		wordCounts[word]++
	}
	// wordCounts = top100(wordCounts)
	min := 0
	max := 1
	for _, v := range wordCounts {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	// for i := range wordCounts {
	// v := wordCounts[i]
	// factor := float64(v-min) / float64(max-min)
	// wordCounts[i] = int(factor * 100)
	// }
	// os.WriteFile(fmt.Sprintf("wordcloud_%d.txt", time.Now().UnixMilli()), []byte(utils.JsonString(wordCounts)), os.ModePerm)
	w := wordclouds.NewWordcloud(
		wordCounts,
		wordclouds.FontFile(consts.DefaultTTFPath),
		// wordclouds.FontMaxSize(1000),
		// wordclouds.FontMinSize(10),
		// wordclouds.RandomPlacement(true),
		// wordclouds.Height(500),
		// wordclouds.Width(500),
		wordclouds.Colors(
			[]color.Color{
				color.RGBA{255, 99, 71, 255},   // 番茄红
				color.RGBA{255, 153, 51, 255},  // 橙色
				color.RGBA{255, 215, 0, 255},   // 金色
				color.RGBA{50, 205, 50, 255},   // 酸橙绿
				color.RGBA{0, 128, 128, 255},   // 海蓝
				color.RGBA{0, 255, 255, 255},   // 青色
				color.RGBA{65, 105, 225, 255},  // 道奇蓝
				color.RGBA{135, 206, 250, 255}, // 天蓝
				color.RGBA{255, 105, 180, 255}, // 热情粉
				color.RGBA{255, 20, 147, 255},  // 深粉色
				color.RGBA{255, 182, 193, 255}, // 刺玫瑰白
				color.RGBA{221, 160, 221, 255}, // 淡紫色
				color.RGBA{147, 112, 219, 255}, // 中兰花紫
				color.RGBA{138, 43, 226, 255},  // 蓝 violet
				color.RGBA{176, 196, 222, 255}, // 淡蓝色
				color.RGBA{100, 149, 237, 255}, // 玉石绿
				color.RGBA{173, 255, 47, 255},  // 绿黄色
				color.RGBA{255, 255, 0, 255},   // 黄色
				color.RGBA{255, 228, 181, 255}, // 桃色
				color.RGBA{160, 82, 45, 255},   // 赤褐色
			}),
	)

	img := w.Draw()
	var buf = bytes.NewBuffer(nil)
	jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes(), nil
}

func top100(wordCounts map[string]int) map[string]int {
	type wordCount struct {
		word  string
		count int
	}
	var words []wordCount
	for word, count := range wordCounts {
		words = append(words, wordCount{word, count})
	}
	sort.Slice(words, func(i, j int) bool {
		return words[i].count > words[j].count
	})
	if len(words) > 100 {
		words = words[:100]
	}
	result := make(map[string]int)
	for _, w := range words {
		result[w.word] = w.count
	}
	return result
}
