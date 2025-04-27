package processor

import (
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/internal/sbi/consumer"
	"github.com/free5gc/smf/pkg/app"
)

const (
	CONTEXT_NOT_FOUND = "CONTEXT_NOT_FOUND"
)

type ProcessorSmf interface {
	app.App

	Consumer() *consumer.Consumer
}

type Processor struct {
	ProcessorSmf
	UsageReports map[string]UsageReportPerUE `json:"usageReports"` // key: SUPI
}

// record the usage reports for each QoS flow of a UE
// key = QFI
// value = []smf_context.UsageReport
type UsageReportPerUE map[uint8][]smf_context.UsageReport 

func NewProcessor(smf ProcessorSmf) (*Processor, error) {
	p := &Processor{
		ProcessorSmf: smf,
		UsageReports: make(map[string]UsageReportPerUE), //use SUPI as key
	}
	return p, nil
}
