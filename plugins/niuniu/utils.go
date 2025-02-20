// Package niu 牛牛大作战
package niuniu

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/RicheyJang/PaimengBot/basic/sc"
)

func randomChoice(options []string) string {
	return options[rand.Intn(len(options))]
}

func profit(niuniu float64) (money float64, t bool, message string) {
	switch {
	case 0 < niuniu && niuniu <= 15:
		message = randomChoice([]string{
			"你的牛牛太小啦",
			"这么小的牛牛就要肩负起这么大的责任吗？快去打胶吧！",
		})
	case niuniu > 15:
		money = (niuniu * 10)
		message = randomChoice([]string{
			fmt.Sprintf("你的牛牛已经离你而去了,你赚取了%d个%s", money, sc.Unit()),
			fmt.Sprintf("啊！你的牛☞已经没啦🤣,为了这点钱就出卖你的牛牛可真不值,你赚取了%d个%s", money, sc.Unit()),
		})
		t = true
	case niuniu <= 0 && niuniu >= -15:
		message = randomChoice([]string{
			"你的牛牛太小啦",
			"这么小的牛牛就要肩负起这么大的责任吗？快去找别人玩吧！",
		})
	case niuniu < -15:
		money = (math.Abs(niuniu * 10))
		message = randomChoice([]string{
			fmt.Sprintf("此世做了女孩子来世来当男孩子(bushi),你赚取了%d个%s", money, sc.Unit()),
			fmt.Sprintf("呜呜呜,不哭不哭当女孩子不委屈的,你赚取了%d个%s", money, sc.Unit()),
		})
		t = true
	}
	return
}

func hitGlueNiuNiu(niuniu float64) (string, float64) {
	probability := rand.Intn(100 + 1)
	reduce := math.Abs(hitGlue(niuniu))
	switch {
	case probability <= 40:
		niuniu += reduce
		return randomChoice([]string{
			fmt.Sprintf("你嘿咻嘿咻一下，促进了牛牛发育，牛牛增加%.2fcm了呢！", reduce),
			fmt.Sprintf("你打了个舒服痛快的🦶呐，牛牛增加了%.2fcm呢！", reduce),
		}), niuniu
	case probability <= 60:
		return randomChoice([]string{
			"你打了个🦶，但是什么变化也没有，好奇怪捏~",
			"你的牛牛刚开始变长了，可过了一会又回来了，什么变化也没有，好奇怪捏~",
		}), niuniu
	default:
		niuniu -= reduce
		if niuniu < 0 {
			return randomChoice([]string{
				fmt.Sprintf("哦吼！？看来你的牛牛凹进去了%.2fcm呢！", reduce),
				fmt.Sprintf("你突发恶疾！你的牛牛凹进去了%.2fcm！", reduce),
				fmt.Sprintf("笑死，你因为打🦶过度导致牛牛凹进去了%.2fcm！🤣🤣🤣", reduce),
			}), niuniu
		}
		return randomChoice([]string{
			fmt.Sprintf("阿哦，你过度打🦶，牛牛缩短%.2fcm了呢！", reduce),
			fmt.Sprintf("你的牛牛变长了很多，你很激动地继续打🦶，然后牛牛缩短了%.2fcm呢！", reduce),
			fmt.Sprintf("小打怡情，大打伤身，强打灰飞烟灭！你过度打🦶，牛牛缩短了%.2fcm捏！", reduce),
		}), niuniu
	}
}

func generateRandomString(niuniu float64) string {
	switch {
	case niuniu <= -100:
		return "wtf？你已经进化成魅魔了！魅魔在击剑时有20%的几率消耗自身长度吞噬对方牛牛呢。"
	case niuniu <= -50:
		return "嗯....好像已经穿过了身体吧..从另一面来看也可以算是凸出来的吧?"
	case niuniu <= -25:
		return randomChoice([]string{
			"这名女生，你的身体很健康哦！",
			"WOW,真的凹进去了好多呢！",
			"你已经是我们女孩子的一员啦！",
		})
	case niuniu <= -10:
		return randomChoice([]string{
			"你已经是一名女生了呢，",
			"从女生的角度来说，你发育良好(,",
			"你醒啦？你已经是一名女孩子啦！",
			"唔...可以放进去一根手指了都...",
		})
	case niuniu <= 0:
		return randomChoice([]string{
			"安了安了，不要伤心嘛，做女生有什么不好的啊。",
			"不哭不哭，摸摸头，虽然很难再长出来，但是请不要伤心啦啊！",
			"加油加油！我看好你哦！",
			"你醒啦？你现在已经是一名女孩子啦！",
		})
	case niuniu <= 10:
		return randomChoice([]string{
			"你行不行啊？细狗！",
			"虽然短，但是小小的也很可爱呢。",
			"像一只蚕宝宝。",
			"长大了。",
		})
	case niuniu <= 25:
		return randomChoice([]string{
			"唔...没话说",
			"已经很长了呢！",
		})
	case niuniu <= 50:
		return randomChoice([]string{
			"话说这种真的有可能吗？",
			"厚礼谢！",
		})
	case niuniu <= 100:
		return randomChoice([]string{
			"已经突破天际了嘛...",
			"唔...这玩意应该不会变得比我高吧？",
			"你这个长度会死人的...！",
			"你马上要进化成牛头人了！！",
			"你是什么怪物，不要过来啊！！",
		})
	default:
		return "惊世骇俗！你已经进化成牛头人了！牛头人在击剑时有20%的几率消耗自身长度吞噬对方牛牛呢。"
	}
}

// fencing 击剑对决逻辑，返回对决结果和myLength的变化值
func fencing(myLength, oppoLength float64) (string, float64, float64) {
	devourLimit := 0.27

	probability := rand.Intn(100) + 1

	switch {
	case oppoLength <= -100 && myLength > 0 && 10 < probability && probability <= 20:
		change := hitGlue(oppoLength) + rand.Float64()*math.Log2(math.Abs(0.5*(myLength+oppoLength)))
		myLength += change
		myLength *= 0.85
		return fmt.Sprintf("对方身为魅魔诱惑了你，你同化成魅魔！当前长度%.2fcm！", -myLength), -myLength, oppoLength

	case oppoLength >= 100 && myLength > 0 && 10 < probability && probability <= 20:
		change := math.Min(math.Abs(devourLimit*myLength), math.Abs(1.5*myLength))
		myLength += change
		myLength *= 0.85
		return fmt.Sprintf("对方以牛头人的荣誉摧毁了你的牛牛！当前长度%.2fcm！", myLength), myLength, oppoLength

	case myLength <= -100 && oppoLength > 0 && 10 < probability && probability <= 20:
		change := hitGlue(myLength+oppoLength) + rand.Float64()*math.Log2(math.Abs(0.5*(myLength+oppoLength)))
		oppoLength -= change
		myLength -= change
		myLength *= 0.85
		return fmt.Sprintf("你身为魅魔诱惑了对方，吞噬了对方部分长度！当前长度%.2fcm！", myLength), myLength, oppoLength

	case myLength >= 100 && oppoLength > 0 && 10 < probability && probability <= 20:
		myLength -= oppoLength
		myLength *= 0.85
		oppoLength = 0.01
		return fmt.Sprintf("你以牛头人的荣誉摧毁了对方的牛牛！当前长度%.2fcm！", myLength), myLength, oppoLength

	default:
		return determineResultBySkill(myLength, oppoLength)
	}
}

// determineResultBySkill 根据击剑技巧决定结果
func determineResultBySkill(myLength, oppoLength float64) (string, float64, float64) {
	probability := rand.Intn(100) + 1
	winProbability := calculateWinProbability(myLength, oppoLength) * 100
	return applySkill(myLength, oppoLength,
		float64(probability) <= winProbability)
}

// calculateWinProbability 计算胜率
func calculateWinProbability(heightA, heightB float64) float64 {
	pA := 0.9
	heightRatio := math.Max(heightA, heightB) / math.Min(heightA, heightB)
	reductionRate := 0.1 * (heightRatio - 1)
	reduction := pA * reductionRate

	adjustedPA := pA - reduction
	return math.Max(adjustedPA, 0.01)
}

// applySkill 应用击剑技巧并生成结果
func applySkill(myLength, oppoLength float64, increaseLength1 bool) (string, float64, float64) {
	reduce := fence(oppoLength)
	// 兜底操作
	if reduce == 0 {
		reduce = rand.Float64() + float64(rand.Intn(3))
	}
	if increaseLength1 {
		myLength += reduce
		oppoLength -= 0.8 * reduce
		if myLength < 0 {
			return fmt.Sprintf("哦吼！？你的牛牛在长大欸！长大了%.2fcm！", reduce), myLength, oppoLength
		}
		return fmt.Sprintf("你以绝对的长度让对方屈服了呢！你的长度增加%.2fcm，当前长度%.2fcm！", reduce, myLength), myLength, oppoLength
	}
	myLength -= reduce
	oppoLength += 0.8 * reduce
	if myLength < 0 {
		return fmt.Sprintf("哦吼！？看来你的牛牛因为击剑而凹进去了呢🤣🤣🤣！凹进去了%.2fcm！", reduce), myLength, oppoLength
	}
	return fmt.Sprintf("对方以绝对的长度让你屈服了呢！你的长度减少%.2fcm，当前长度%.2fcm！", reduce, myLength), myLength, oppoLength
}

// fence 根据长度计算减少的长度
func fence(rd float64) float64 {
	rd = math.Abs(rd)
	if rd == 0 {
		rd = 1
	}
	r := hitGlue(rd)*2 + rand.Float64()*math.Log2(rd)

	return float64(int(r * rand.Float64()))
}

func hitGlue(l float64) float64 {
	if l == 0 {
		l = 0.1
	}
	l = math.Abs(l)
	switch {
	case l > 1 && l <= 10:
		return rand.Float64() * math.Log2(l*2)
	case 10 < l && l <= 100:
		return rand.Float64() * math.Log2(l*1.5)
	case 100 < l && l <= 1000:
		return rand.Float64() * (math.Log10(l*1.5) * 2)
	case l > 1000:
		return rand.Float64() * (math.Log10(l) * 2)
	default:
		return rand.Float64()
	}
}
