package processor

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/smf/internal/logger"
	smf_context "github.com/free5gc/smf/internal/context"
)

type PDUSessionInfo struct {
	Supi         string
	PDUSessionID string
	Dnn          string
	Sst          string
	Sd           string
	AnType       models.AccessType
	PDUAddress   string
	SessionRule  models.SessionRule
	UpCnxState   models.UpCnxState
	Tunnel       smf_context.UPTunnel
}

func (p *Processor) HandleOAMGetUEPDUSessionInfo(c *gin.Context, smContextRef string) {
	smContext := smf_context.GetSMContextByRef(smContextRef)
	if smContext == nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	pduSessionInfo := &PDUSessionInfo{
		Supi:         smContext.Supi,
		PDUSessionID: strconv.Itoa(int(smContext.PDUSessionID)),
		Dnn:          smContext.Dnn,
		Sst:          strconv.Itoa(int(smContext.SNssai.Sst)),
		Sd:           smContext.SNssai.Sd,
		AnType:       smContext.AnType,
		PDUAddress:   smContext.PDUAddress.String(),
		UpCnxState:   smContext.UpCnxState,
		// Tunnel: context.UPTunnel{
		// 	//UpfRoot:  smContext.Tunnel.UpfRoot,
		// 	ULCLRoot: smContext.Tunnel.UpfRoot,
		// },
	}
	c.JSON(http.StatusOK, pduSessionInfo)
}

func (p *Processor) HandleGetSMFDebugInfo(c *gin.Context) {
	ue1Supi := "imsi-208930000000001"
	PduSessionId := int32(1)

	type debugInfo struct {
		PCCRulesMap     map[string]smf_context.PCCRule 	`json:"pccRulesMap,omitempty"`
		ChargingDataMap map[string]models.ChargingData 	`json:"chargingDataMap,omitempty"`
	}

	ue1SmCtx := smf_context.GetSMContextById(ue1Supi, PduSessionId)
	if ue1SmCtx == nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	pccRules := make(map[string]smf_context.PCCRule)
	for key, value := range ue1SmCtx.PCCRules {
		if value != nil {
			pccRules[key] = *value
		}
	}

	chargingData := make(map[string]models.ChargingData)
	for key, value := range ue1SmCtx.ChargingData {
		if value != nil {
			chargingData[key] = *value
		}
	}

	rsp := debugInfo{
		PCCRulesMap:     pccRules,
		ChargingDataMap: chargingData,
	}

	logger.PduSessLog.Infof("Debug Info: %+v", rsp)

	c.JSON(http.StatusOK, rsp)
}

func (p *Processor) HandleGetUserUsageInfo(c *gin.Context) {

	c.JSON(http.StatusOK, p.UsageReports)

	// clean up the reports
	p.UsageReports = make(map[string]UsageReportPerUE, 0)
}
