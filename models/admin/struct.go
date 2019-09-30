package admin

type DataChannelPolicy struct {
	Id         int64
	Name       string `orm:"null"`
	Enabled    bool   `orm:"description(是否启用)"`
	Deleted    int    `orm:"default(0)"`
	Operator   int64  `orm:"null;description(操作人);"`
	CreateTime int64  `orm:"null"`
	UpdateTime int64  `orm:"null"`
}

func (DataChannelPolicy) TableName() string {
	return "admin_data_channel_policy"
}
