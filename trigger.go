package ping

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

// DefaultPort is the default port for Ping service
const DefaultPort = "9090"

// log is the default package logger
var log = logger.GetLogger("trigger-mashling-ping")

// Trigger is the ping trigger
type Trigger struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	response string
	*http.Server
}

// NewFactory create a new Trigger factory
func NewFactory(metadata *trigger.Metadata) trigger.Factory {
	return &Factory{metadata: metadata}
}

// Factory Ping Trigger factory
type Factory struct {
	metadata *trigger.Metadata
}

// New Creates a new trigger instance for a given id
func (f *Factory) New(config *trigger.Config) trigger.Trigger {
	type PingResponse struct {
		Version        string
		Appversion     string
		Appdescription string
	}

	response := PingResponse{
		Version:        config.GetSetting("version"),
		Appversion:     config.GetSetting("appversion"),
		Appdescription: config.GetSetting("appdescription"),
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Error("Ping service data formation error")
	}

	port := config.GetSetting("port")
	if len(port) == 0 {
		port = DefaultPort
	}

	mux := http.NewServeMux()
	trigger := &Trigger{
		metadata: f.metadata,
		config:   config,
		response: string(data),
		Server: &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		},
	}

	mux.HandleFunc("/ping", trigger.PingResponseHandlerShort)
	mux.HandleFunc("/ping/details", trigger.PingResponseHandlerDetail)
	return trigger
}

// Metadata implements trigger.Trigger.Metadata
func (t *Trigger) Metadata() *trigger.Metadata {
	return t.metadata
}

// Initialize start the Ping service
func (t *Trigger) Initialize(context trigger.InitContext) error {
	return nil
}

func (t *Trigger) PingResponseHandlerShort(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "{\"response\":\"Ping successful\"}\n")
}

//PingResponseHandlerDetail handles simple response
func (t *Trigger) PingResponseHandlerDetail(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, t.response+"\n")
}

// Start implements util.Managed.Start
func (t *Trigger) Start() error {
	log.Info("Ping service starting...")

	go func() {
		if err := t.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Ping service err:", err)
		}
	}()
	log.Info("Ping service started")
	return nil
}

// Stop implements util.Managed.Stop
func (t *Trigger) Stop() error {
	if err := t.Shutdown(nil); err != nil {
		log.Errorf("[mashling-ping-service] Ping service error when stopping:", err)
		return err
	}
	log.Info("[mashling-ping-service] Ping service stopped")
	return nil
}
