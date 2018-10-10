package nsqd

import (
	"net/http"
	"net/url"
	"nsq-learn/internal/http_api"
	"nsq-learn/internal/protocol"
	"nsq-learn/internal/version"
	"os"

	"github.com/julienschmidt/httprouter"
)

type httpServer struct {
	ctx         *context
	router      http.Handler
	tlsEnabled  bool
	tlsRequired bool
}

func NewHttpServer(ctx *context, tlsEnabled bool, tlsRequired bool) *httpServer {
	log := http_api.Log(ctx.nsqd.logf)
	router := httprouter.New()
	// 如果没有对用的路由 返回405
	router.HandleMethodNotAllowed = true
	router.PanicHandler = http_api.LogPanicHandler(ctx.nsqd.logf)
	router.NotFound = http_api.LogNotFoundHandler(ctx.nsqd.logf)
	router.MethodNotAllowed = http_api.LogMethodNotAllowedHandler(ctx.nsqd.logf)
	s := &httpServer{
		ctx:         ctx,
		router:      router,
		tlsEnabled:  tlsEnabled,
		tlsRequired: tlsRequired,
	}
	router.Handle("GET", "/ping", http_api.Decorate(s.pingHandler, log, http_api.PlainText))
	router.Handle("GET", "/info", http_api.Decorate(s.doInfo, log, http_api.V1))
	// 创建topic
	router.Handle("POST", "/topic/create", http_api.Decorate(s.doCreateTopic, log, http_api.V1))
	return s
}

func (s *httpServer) pingHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	health := s.ctx.nsqd.getHealth()
	if !s.ctx.nsqd.isHealth() {
		return nil, http_api.Err{Code: 500, Text: health}
	}
	return health, nil
}

func (s *httpServer) doInfo(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, http_api.Err{Code: 500, Text: err.Error()}
	}
	return struct {
		Version  string `json:"version"`
		Hostname string `json:"hostname"`
	}{
		Version:  version.Binary,
		Hostname: hostname,
	}, nil
}

// 去除了https支持
func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

func (s *httpServer) doCreateTopic(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	_, _, err := s.getTopicFromQuery(req)
	return nil, err
}

func (s *httpServer) getTopicFromQuery(req *http.Request) (url.Values, *Topic, error) {
	reqParams, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		s.ctx.nsqd.logf(LOG_ERROR, "failed to parse request params - %s", err)
		return nil, nil, http_api.Err{400, "INVALID_REQUEST"}
	}

	topicNames, ok := reqParams["topic"]
	if !ok {
		return nil, nil, http_api.Err{400, "MISSING_ARG_TOPIC"}
	}
	topicName := topicNames[0]
	if !protocol.IsValidTopicName(topicName) {
		return nil, nil, http_api.Err{400, "INVALID_TOPIC"}
	}

	// if !protocol.IsValidTopicName(topicName) {
	// 	return nil, nil, http_api.Err{400, "INVALID_TOPIC"}
	// }

	return reqParams, s.ctx.nsqd.GetTopic(topicName), nil
}
