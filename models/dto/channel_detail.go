package dto

import (
	"fmt"
	"strconv"
	"strings"
	"tsEngine/tsMicro"
	"tsEngine/tsString"
)

type ReqChannelDetail struct {
	tsMicro.MicroSign
	ChannelId     int64  `valid:"required;"`
	Platform      string `valid:"required;enum[alipay,wechat,bank,yunshanfu]"`
	Reduce        int    `valid:""`
	Type          int    `valid:"enum[1,2,3]"`
	VipWeights    int    `valid:"range[1-100]"`
	VipRange      string `valid:""`
	NormalWeights int    `valid:"range[1-100]"`
	NormalRange   string `valid:""`
}

type ReqChannelDetailEdit struct {
	tsMicro.MicroSign
	Id            int64  `valid:"required;"`
	Name          string `valid:""`
	Platform      string `valid:"required;enum[alipay,wechat,bank,yunshanfu]"`
	Reduce        int    `valid:""`
	Type          int    `valid:"enum[1,2,3]"`
	VipWeights    int    `valid:"range[1-100]"`
	VipRange      string `valid:""`
	NormalWeights int    `valid:"range[1-100]"`
	NormalRange   string `valid:""`
}

type ReqChannelDetailEnable struct {
	Id      int64
	Enabled bool
}

type ReqChannelDetailType struct {
	Id   int64
	Type int `valid:"required;enum[1,2,3]"`
}

type ReqChannelDetailWeights struct {
	tsMicro.MicroSign
	Id      int64
	Type    int `valid:"required;enum[1,2]"`
	Weights int `valid:"required;range[1:100]"`
}

// 确定该通道是否符合指定金额
func (this *ReqChannelDetail) CheckAmountRange(vip bool) bool {
	// 获取金额区间
	amountRange := this.NormalRange
	if vip {
		amountRange = this.VipRange
	}

	// 判断是否为指定金额
	if strings.Contains(amountRange, "-") {
		// 金额区间
		amountSlice := strings.Split(amountRange, "-")
		min := tsString.ToInt(amountSlice[0])
		max := tsString.ToInt(amountSlice[1])

		return min < max && 0 < min
	} else {
		// 指定金额
		amountRange = fmt.Sprintf(",%s,", amountRange)
		rangeSlice := strings.Split(amountRange, ",")

		for _, a := range rangeSlice {
			// 空
			if len(a) == 0 {
				continue
			}

			// 非数字
			if amount, err := strconv.ParseInt(a, 10, 64); err != nil || amount < 1 {
				return false
			}

			// 重复的金额设置 1000,5000,5000
			if strings.Count(amountRange, fmt.Sprintf(",%s,", a)) == 2 {
				return false
			}
		}
		return true
	}
}

// 确定该通道是否符合指定金额
func (this *ReqChannelDetailType) CheckAmountRange(amountRange string) bool {
	// 判断是否为指定金额
	if strings.Contains(amountRange, "-") {
		// 金额区间
		amountSlice := strings.Split(amountRange, "-")
		min := tsString.ToInt(amountSlice[0])
		max := tsString.ToInt(amountSlice[1])

		return min < max && 0 < min
	} else {
		// 指定金额
		amountRange = fmt.Sprintf(",%s,", amountRange)
		rangeSlice := strings.Split(amountRange, ",")

		for _, a := range rangeSlice {
			// 空
			if len(a) == 0 {
				continue
			}

			// 非数字
			if amount, err := strconv.ParseInt(a, 10, 64); err != nil || amount < 1 {
				return false
			}

			// 重复的金额设置 1000,5000,5000
			if strings.Count(amountRange, fmt.Sprintf(",%s,", a)) == 2 {
				return false
			}
		}
		return true
	}
}

// 确定该通道是否符合指定金额
func (this *ReqChannelDetailEdit) CheckAmountRange(vip bool) bool {
	// 获取金额区间
	amountRange := this.NormalRange
	if vip {
		amountRange = this.VipRange
	}

	// 判断是否为指定金额
	if strings.Contains(amountRange, "-") {
		// 金额区间
		amountSlice := strings.Split(amountRange, "-")
		min := tsString.ToInt(amountSlice[0])
		max := tsString.ToInt(amountSlice[1])

		return min < max && 0 < min
	} else {
		// 指定金额
		amountRange = fmt.Sprintf(",%s,", amountRange)
		rangeSlice := strings.Split(amountRange, ",")

		for _, a := range rangeSlice {
			// 空
			if len(a) == 0 {
				continue
			}

			// 非数字
			if amount, err := strconv.ParseInt(a, 10, 64); err != nil || amount < 1 {
				return false
			}

			// 重复的金额设置 1000,5000,5000
			if strings.Count(amountRange, fmt.Sprintf(",%s,", a)) == 2 {
				return false
			}
		}
		return true
	}
}
