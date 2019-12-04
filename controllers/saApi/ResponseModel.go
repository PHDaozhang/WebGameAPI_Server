package saApi

import "encoding/xml"

type LoginRequestResponse struct {
	XMLName xml.Name `xml:"LoginRequestResponse"`
	Token    string      `xml:"Token""`
	DisplayName    string   `xml:"DisplayName"`
	GameURL    string   `xml:"GameURL"`
	ErrorMsgId    int   `xml:"ErrorMsgId"`
	ErrorMsg    string   `xml:"ErrorMsg"`
}
