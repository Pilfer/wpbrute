package models

import "encoding/xml"

type XMLRPCLoginRequestPayload struct {
	XMLName    xml.Name                        `xml:"methodCall"`
	Text       string                          `xml:",chardata"`
	MethodName string                          `xml:"methodName"`
	Params     XMLRPCLoginRequestPayloadParams `xml:"params"`
}

type XMLRPCLoginRequestPayloadParams struct {
	Param []XMLRPCLoginRequestPayloadParam `xml:"param"`
}

type XMLRPCLoginRequestPayloadParam struct {
	Text  string `xml:",chardata"`
	Value string `xml:"value"`
}

type XMLRPCCheckEnabledPayload struct {
	XMLName    xml.Name `xml:"methodCall"`
	Text       string   `xml:",chardata"`
	MethodName string   `xml:"methodName"`
	Params     string   `xml:"params"`
}

type XMLRPCLoginResponseObject struct {
	IsInvalid        bool
	IsAdmin          bool
	IsAuthor         bool
	IsRateLimited    bool
	IsXMLRPCDisabled bool
}

// the respond of XMLRPC test (listMethods)
type XMLRPCListMethodsResponse struct {
	XMLName xml.Name `xml:"methodResponse"`
	Text    string   `xml:",chardata"`
	Params  struct {
		Text  string `xml:",chardata"`
		Param struct {
			Text  string `xml:",chardata"`
			Value struct {
				Text  string `xml:",chardata"`
				Array struct {
					Text string `xml:",chardata"`
					Data struct {
						Text  string `xml:",chardata"`
						Value []struct {
							Text   string `xml:",chardata"`
							String string `xml:"string"`
						} `xml:"value"`
					} `xml:"data"`
				} `xml:"array"`
			} `xml:"value"`
		} `xml:"param"`
	} `xml:"params"`
}

func NewXMLRPCLoginRequestPayload(username, password string) *XMLRPCLoginRequestPayload {
	return &XMLRPCLoginRequestPayload{
		MethodName: "wp.getUsersBlogs",
		Params: XMLRPCLoginRequestPayloadParams{
			Param: []XMLRPCLoginRequestPayloadParam{
				{Value: username},
				{Value: password},
			},
		},
	}
}

func NewXMLRPCCheckEnabledPayload() *XMLRPCCheckEnabledPayload {
	return &XMLRPCCheckEnabledPayload{
		MethodName: "system.listMethods",
	}
}
