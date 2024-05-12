package bench

type harPageTimings struct {
	OnContentLoad float64 `json:"onContentLoad"`
	OnLoad        float64 `json:"onLoad"`
}

type harPage struct {
	StartedDateTime string         `json:"startedDateTime"`
	Id              string         `json:"id"`
	Title           string         `json:"title"`
	PageTimings     harPageTimings `json:"pageTimings"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type harInitiaor struct {
	Type string `json:"type"`
}

type harCache struct{}

type harNameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Expires  string `json:"expires"`
	HttpOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

type harRequest struct {
	Method      string `json:"method"`
	Url         string `json:"url"`
	HttpVersion string `json:"httpVersion"`

	Headers     []harNameValue `json:"headers"`
	QueryString []harNameValue `json:"queryString"`
	Cookies     []harCookie    `json:"cookies"`
	HeadersSize int            `json:"headersSize"`
	BodySize    int64          `json:"bodySize"`
}

type harContent struct {
	Size         int64  `json:"size"`
	MimeType     string `json:"mimeType"`
	Text         string `json:"text"`
	RedirectURL  string `json:"redirectURL"`
	HeadersSize  int    `json:"headersSize"`
	BodySize     int64  `json:"bodySize"`
	TransferSize int64  `json:"transferSize"`
	Error        string `json:"_error"`
}

type harResponse struct {
	Status      int16       `json:"status"`
	StatusText  string      `json:"statusText"`
	HttpVersion string      `json:"httpVersion"`
	Cookies     []harCookie `json:"cookies"`
	Content     harContent  `json:"content"`
}

type harTimings struct {
	Blocked         float64 `json:"blocked"`
	Dns             float64 `json:"dns"`
	Ssl             float64 `json:"ssl"`
	Connect         float64 `json:"connect"`
	Send            float64 `json:"send"`
	Wait            float64 `json:"wait"`
	Receive         float64 `json:"receive"`
	BlockedQueueing float64 `json:"_blocked_queueing"`
}

type harEntry struct {
	Initiator       harInitiaor `json:"_initiator"`
	Priority        string      `json:"_priority"`
	ResourceType    string      `json:"_resourceType"`
	Cache           harCache    `json:"cache"`
	Connection      string      `json:"connection"`
	PageRef         string      `json:"pageref"`
	Request         harRequest  `json:"request"`
	Response        harResponse `json:"response"`
	ServerIPAddress string      `json:"serverIPAddress"`
	StartedDateTime string      `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Timings         harTimings  `json:"timings"`
}

type harLog struct {
	Version string     `json:"version"`
	Creator harCreator `json:"creator"`
	Pages   []harPage  `json:"pages"`
	Entries []harEntry `json:"entries"`
}

type har struct {
	Log harLog `json:"log"`
}
