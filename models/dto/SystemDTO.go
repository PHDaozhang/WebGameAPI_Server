package dto

import "github.com/astaxie/beego"

// 通用的请求处理接口
type CommReqInterface interface {
	CheckRequest() bool
	SetDefault()
	GetHash() (hash string)
}

type CommonReq struct {
	Hash string
}

// 登陆验证的 struct
type ReqLogin struct {
	Username string
	Password string
	Captcha  string
}

// 通用的search struct
type ReqSearch struct {
	Keyword     string
	Page        int64
	PageSize    int64
	Sort        string
	BeginTime   int64
	EndTime     int64
	FilterKey   string // ChannelId// 筛选指定条件的数据
	FilterValue string // ss1001,ss1002
	Filter      string //[{"key":"xxx","value":"xxx"},...]
}

func (this *ReqSearch) SetDefault() {
	if this.PageSize < 0 || this.PageSize > 100 {
		this.PageSize = 20
	}

	if this.Page < 1 {
		this.Page = 0
	}

	if len(this.Sort)==0 {
		this.Sort = "CreateTime"
	}
}

// 通用的search struct
type ReqChannelSearch struct {
	Keyword   string
	ChannelId string
	Page      int64
	PageSize  int64
	Sort      string
	BeginTime uint64
	EndTime   uint64
}

// admin update
type ReqAddAdmin struct {
	CommonReq
	Username        string
	Password        string
	ConfirmPassword string
	Status          int64
	Role            string
	CreateRole      string
	Name            string
	Mobile          string
	ContactInf      string
	Note            string
	AdminType       int64
	Scores          int64
	CloudCoin       float64
	BuyScoreRate    float64
	ReturnRate      float64
	SettleType      int
	//SettleRate      float64
	SettlePrice float64
}

// admin update
type ReqEditAdmin struct {
	Id           int64
	AdminId      int64
	Username     string
	Password     string
	Status       int64
	Role         string
	CreateRole   string
	Name         string
	ContactInf   string
	Mobile       string
	Note         string
	Scores       int64
	BuyScoreRate float64
	ReturnRate   float64
	SettleType   int
	//SettleRate   int
	SettlePrice float64
}

// model update
type ReqAddModel struct {
	CommonReq
	Name        string
	Type        int64
	Key         string
	ParentId    int64
	Logs        int64
	Description string
	Sort        int64
	NodeId      int64
}

// node update
type ReqAddNode struct {
	CommonReq
	Name        string
	Url         string
	Icon        string
	Sort        int64
	LangCn      string
	LangTw      string
	LangUs      string
	Description string
	ParentId    int64
}

// Template add/update
type ReqAddTemplate struct {
	CommonReq
	ChannelId    string
	PlatformName string
	GameVersion  int
	Icon         int
	Background   int64
	Games        string
	IpLimit      string
	Description  string
}

type List struct {
	ReqSearch
	PlayerId int64
	Status   int
}

func (this *ReqAddTemplate) GetHash() (hash string) {
	return this.Hash
}

func (this *ReqAddTemplate) CheckRequest() bool {
	return len(this.ChannelId) > 0 && len(this.PlatformName) > 0
}

func (this *ReqAddTemplate) SetDefault() {
	// nothing
}

type ReqAddWithdraw struct {
	CommonReq
	ChannelId   string
	Account     string
	Name        string
	AccountType string
	RealMoney   int
	Money       int
	ChargeFee   int
	LeftGold    int
	Status      int
}

func (this *ReqAddWithdraw) GetHash() (hash string) {
	return this.Hash
}

func (this *ReqAddWithdraw) CheckRequest() bool {
	return len(this.Account) > 0 && len(this.Name) > 0 && len(this.AccountType) > 0
}

func (this *ReqAddWithdraw) SetDefault() {
	// nothing
}

type ReqAddPayment struct {
	CommonReq
	PayUrl      string
	PayType     int
	PayName     string
	CallBackUrl string
	Status      int
	Sort        int
}

func (this *ReqAddPayment) GetHash() (hash string) {
	return this.Hash
}

func (this *ReqAddPayment) CheckRequest() bool {
	payTypes := beego.AppConfig.Strings("PayType::PayType")
	return len(this.PayName) > 0 && this.PayType >= 0 && len(this.PayUrl) > 0 && len(this.CallBackUrl) > 0 && len(payTypes) > this.PayType
}

func (this *ReqAddPayment) SetDefault() {
	if this.Sort == 0 {
		this.Sort = 1
	}
	if this.Status == 0 {
		this.Status = 1
	}
}

type ReqAddPaymentConfig struct {
	CommonReq
	VendorName  string
	VendorType  int
	Quality     int
	IsOpen      int
	Discount    float64
	IsHot       int
	RandTimes   int64
	CalledTimes int64
	MinPay      int64
	MaxPay      int64
	Weigh       int
	Sort        int
}

func (this *ReqAddPaymentConfig) GetHash() (hash string) {
	return this.Hash
}

func (this *ReqAddPaymentConfig) CheckRequest() bool {
	paySwitchTypes := beego.AppConfig.Strings("PaySwitchTypes::PaySwitchTypes")
	return len(this.VendorName) > 0 && this.VendorType >= 0 && len(paySwitchTypes) > this.VendorType && (this.Discount >= 0 && this.Discount <= 1)
}

func (this *ReqAddPaymentConfig) SetDefault() {
	if this.Sort == 0 {
		this.Sort = 1
	}
	if this.IsOpen == 0 {
		this.IsOpen = 1
	}
}

type ReqAddPaymentRequest struct {
	CommonReq
	ChannelId string
	PayType   int
	DestId    string
	Amount    int64
	Status    int
	Sort      int
}

func (this *ReqAddPaymentRequest) GetHash() (hash string) {
	return this.Hash
}

func (this *ReqAddPaymentRequest) CheckRequest() bool {
	return len(this.ChannelId) > 0 && this.Amount > 0 && this.PayType > 0
}

func (this *ReqAddPaymentRequest) SetDefault() {
	if this.Sort == 0 {
		this.Sort = 1
	}
	if this.Status == 0 {
		this.Status = 1
	}
}
