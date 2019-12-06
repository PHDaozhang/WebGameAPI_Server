package dto

type AgentInfo struct {
	ChannelId			string 		`bson:"ChannelId"`
	AgentId				int64		`bson:"AgentId"`
	AgentName			string 		`bson:"AgentName"`
}