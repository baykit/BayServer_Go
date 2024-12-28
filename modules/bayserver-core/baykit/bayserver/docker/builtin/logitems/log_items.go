package logitems

import (
	"bayserver-core/baykit/bayserver/tour"
	"strconv"
	"strings"
	"time"
)

/****************************************/
/* TextItem                             */
/****************************************/

type TextItem struct {
	/** text to print */
	text string
}

func NewTextItem(text string) LogItem {
	item := TextItem{
		text: text,
	}
	return &item
}

func (t TextItem) Init(param string) {
}

func (t TextItem) GetItem(tour tour.Tour) string {
	return t.text
}

/****************************************/
/* NullItem                             */
/****************************************/

/**
 * Return null result
 */

type NullItem struct {
}

func NewNullItem() LogItem {
	return &NullItem{}
}

func (n NullItem) Init(param string) {
}

func (n NullItem) GetItem(tour tour.Tour) string {
	return ""
}

/****************************************/
/* RemoteIpItem                         */
/****************************************/

/**
 * Return remote IP address (%a)
 */

type RemoteIpItem struct {
}

func NewRemoteIpItem() LogItem {
	return &RemoteIpItem{}
}

func (r RemoteIpItem) Init(param string) {

}

func (r RemoteIpItem) GetItem(tour tour.Tour) string {
	return tour.Req().RemoteAddress()
}

/****************************************/
/* ServerIpItem                         */
/****************************************/

/**
 * Return local IP address (%A)
 */

type ServerIpItem struct {
}

func NewServerIpItem() LogItem {
	return &ServerIpItem{}
}

func (s ServerIpItem) Init(param string) {

}

func (s ServerIpItem) GetItem(tour tour.Tour) string {
	return tour.Req().ServerAddress()
}

/****************************************/
/* RequestBytesItem1                    */
/****************************************/

/**
 * Return number of bytes that is sent from clients (Except HTTP headers)
 * (%B)
 */

type RequestBytesItem1 struct {
}

func NewRequestBytesItem1() LogItem {
	return &RequestBytesItem1{}
}

func (r RequestBytesItem1) Init(param string) {

}

func (r RequestBytesItem1) GetItem(tour tour.Tour) string {
	bytes := tour.Req().Headers().ContentLength()
	if bytes < 0 {
		bytes = 0
	}

	return strconv.Itoa(bytes)
}

/****************************************/
/* RequestBytesItem2                    */
/****************************************/

/**
 * Return number of bytes that is sent from clients in CLF format (Except
 * HTTP headers) (%b)
 */

type RequestBytesItem2 struct {
}

func NewRequestBytesItem2() LogItem {
	return &RequestBytesItem2{}
}

func (r RequestBytesItem2) Init(param string) {

}

func (r RequestBytesItem2) GetItem(tour tour.Tour) string {
	bytes := tour.Req().Headers().ContentLength()
	if bytes < 0 {
		return "-"
	}

	return strconv.Itoa(bytes)
}

/****************************************/
/* ConnectionStatusItem                 */
/****************************************/

/**
 * Return connection status (%c)
 */

type ConnectionStatusItem struct {
}

func NewConnectionStatusItem() LogItem {
	return &ConnectionStatusItem{}
}

func (c ConnectionStatusItem) Init(param string) {

}

func (c ConnectionStatusItem) GetItem(tour tour.Tour) string {
	if tour.IsAborted() {
		return "X"
	} else {
		return "-"
	}
}

/****************************************/
/* FileNameItem                         */
/****************************************/

type FileNameItem struct {
}

func NewFileNameItem() LogItem {
	return &FileNameItem{}
}

func (f FileNameItem) Init(param string) {

}

func (f FileNameItem) GetItem(tour tour.Tour) string {
	return tour.Req().ScriptName()
}

/****************************************/
/* RemoteHostItem                       */
/****************************************/

type RemoteHostItem struct {
}

func NewRemoteHostItem() LogItem {
	return &RemoteHostItem{}
}

func (r RemoteHostItem) Init(param string) {

}

func (r RemoteHostItem) GetItem(tour tour.Tour) string {
	return tour.Req().ReqHost()
}

/****************************************/
/* RemoteLogItem                        */
/****************************************/

/**
 * Return remote log name (%l)
 */

type RemoteLogItem struct {
}

func NewRemoteLogItem() LogItem {
	return &RemoteLogItem{}
}

func (r RemoteLogItem) Init(param string) {

}

func (r RemoteLogItem) GetItem(tour tour.Tour) string {
	return ""
}

/****************************************/
/* ProtocolItem                         */
/****************************************/

type ProtocolItem struct {
}

func NewProtocolItem() LogItem {
	return &ProtocolItem{}
}

func (p ProtocolItem) Init(param string) {

}

func (p ProtocolItem) GetItem(tour tour.Tour) string {
	return tour.Req().Protocol()
}

/****************************************/
/* RequestHeaderItem                    */
/****************************************/

/**
 * Return requested header (%{Foobar}i)
 */

type RequestHeaderItem struct {
	/** Header name */
	name string
}

func NewRequestHeaderItem() LogItem {
	return &RequestHeaderItem{}
}

func (r RequestHeaderItem) Init(param string) {
	r.name = param
}

func (r RequestHeaderItem) GetItem(tour tour.Tour) string {
	return tour.Req().Headers().Get(r.name)
}

/****************************************/
/* MethodItem                           */
/****************************************/

/**
 * Return request method (%m)
 */

type MethodItem struct {
}

func NewMethodItem() LogItem {
	return &MethodItem{}
}
func (m MethodItem) Init(param string) {

}

func (m MethodItem) GetItem(tour tour.Tour) string {
	return tour.Req().Method()
}

/****************************************/
/* ResponseHeaderItem                   */
/****************************************/

/**
 * Return responde header (%{Foobar}o)
 */

type ResponseHeaderItem struct {
	name string
}

func NewResponseHeaderItem() LogItem {
	return &ResponseHeaderItem{}
}

func (r ResponseHeaderItem) Init(param string) {
	r.name = param
}

func (r ResponseHeaderItem) GetItem(tour tour.Tour) string {
	return tour.Res().Headers().Get(r.name)
}

/****************************************/
/* PortItem                             */
/****************************************/

/**
 * The server port (%p)
 */

type PortItem struct {
}

func NewPortItem() LogItem {
	return &PortItem{}
}

func (p PortItem) Init(param string) {

}

func (p PortItem) GetItem(tour tour.Tour) string {
	return strconv.Itoa(tour.Req().ServerPort())
}

/****************************************/
/* QueryStringItem                      */
/****************************************/

type QueryStringItem struct {
}

func NewQueryStringItem() LogItem {
	return &QueryStringItem{}
}

func (q QueryStringItem) Init(param string) {

}

func (q QueryStringItem) GetItem(tour tour.Tour) string {
	qStr := tour.Req().QueryString()
	if qStr != "" {
		return "?" + qStr
	} else {
		return ""
	}
}

/****************************************/
/* StartLineItem                        */
/****************************************/

/**
 * The start line (%r)
 */

type StartLineItem struct {
}

func (s StartLineItem) Init(param string) {

}

func (s StartLineItem) GetItem(tour tour.Tour) string {
	return tour.Req().Method() + " " + tour.Req().Uri() + " " + tour.Req().Protocol()
}

func NewStartLineItem() LogItem {
	return &StartLineItem{}
}

/****************************************/
/* StatusItem                           */
/****************************************/

/**
 * Return status (%s)
 */

type StatusItem struct {
}

func NewStatusItem() LogItem {
	return &StatusItem{}
}

func (s StatusItem) Init(param string) {

}

func (s StatusItem) GetItem(tour tour.Tour) string {
	return strconv.Itoa(tour.Res().Headers().Status())
}

/****************************************/
/* TimeItem                             */
/****************************************/

type TimeItem struct {
	layout string
}

func NewTimeItem() LogItem {
	return &TimeItem{
		layout: "01/Jan/2000 23:59:59 -0900",
	}
}

func (t TimeItem) Init(param string) {
	t.layout = param
}

func (t TimeItem) GetItem(tour tour.Tour) string {
	return time.Now().Format(t.layout)
}

/****************************************/
/* IntervalItem                         */
/****************************************/

/**
 * Return how long request took (%T)
 */

type IntervalItem struct {
}

func NewIntervalItem() LogItem {
	return &IntervalItem{}
}

func (i IntervalItem) Init(param string) {

}

func (i IntervalItem) GetItem(tour tour.Tour) string {
	return strconv.Itoa(tour.Interval() / 1000)
}

/****************************************/
/* RemoteUserItem                       */
/****************************************/

/**
 * Return remote user (%u)
 */

type RemoteUserItem struct {
}

func NewRemoteUserItem() LogItem {
	return &RemoteUserItem{}
}

func (r RemoteUserItem) Init(param string) {

}

func (r RemoteUserItem) GetItem(tour tour.Tour) string {
	return tour.Req().RemoteUser()
}

/****************************************/
/* RequestUrlItem                       */
/****************************************/

/**
 * Return requested URL(not content query string) (%U)
 */

type RequestUrlItem struct {
}

func NewRequestUrlItem() LogItem {
	return &RequestUrlItem{}
}

func (r RequestUrlItem) Init(param string) {

}

func (r RequestUrlItem) GetItem(tour tour.Tour) string {
	url := tour.Req().Uri()
	pos := strings.Index(url, "?")
	if pos >= 0 {
		url = url[:pos]
	}
	return url
}

/****************************************/
/* ServerNameItem                       */
/****************************************/

type ServerNameItem struct {
}

func NewServerNameItem() LogItem {
	return &ServerNameItem{}
}

func (s ServerNameItem) Init(param string) {

}

func (s ServerNameItem) GetItem(tour tour.Tour) string {
	return tour.Req().ServerName()
}
