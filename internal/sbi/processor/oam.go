package processor

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/models"
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/internal/logger"
)

type PDUSessionInfo struct {
	Supi         string                 `json:"supi"`
	PDUSessionID string                 `json:"pduSessionId"`
	PDUAddress   string                 `json:"pduAddress"`
	QosFlows     map[string]QosFlowInfo `json:"qosFlows,omitempty"`
}

type QosFlowInfo struct {
	Var5qi    int32  `json:"5qi"`
	State     string `json:"state"`
	IsGBRFlow bool   `json:"isGbrFlow"`
	MaxbrUl   string `json:"maxbrUl,omitempty"`
	MaxbrDl   string `json:"maxbrDl,omitempty"`
	GbrUl     string `json:"gbrUl,omitempty"`
	GbrDl     string `json:"gbrDl,omitempty"`
	SdfFilter string `json:"sdfFilter,omitempty"`
}

// Helper function to format bit rate
func formatBitRate(bitsPerSec uint64) string {
	mbps := bitsPerSec / 1000000
	return strconv.FormatUint(mbps, 10) + " Mbps"
}

// Helper function to convert SDF filter format
func convertSdfFilter(sdfFilter string) string {
	if sdfFilter == "" {
		return ""
	}

	// Handle "permit out ip from 1.1.1.1/32 to assigned" -> "1.1.1.1/32"
	if strings.Contains(sdfFilter, "permit out ip from") && strings.Contains(sdfFilter, "to assigned") {
		// Extract the source IP/subnet between "from" and "to"
		parts := strings.Split(sdfFilter, " ")
		for i, part := range parts {
			if part == "from" && i+1 < len(parts) {
				sourceIP := parts[i+1]
				if sourceIP != "any" {
					return sourceIP
				}
				return "any"
			}
		}
	}

	// Return original if no conversion pattern matches
	return sdfFilter
}

// Helper function to get QoS flow state string
func getQosFlowStateString(state smf_context.QoSFlowState) string {
	switch state {
	case smf_context.QoSFlowUnset:
		return "Unset"
	case smf_context.QoSFlowSet:
		return "Set"
	case smf_context.QoSFlowToBeModify:
		return "ToBeModify"
	default:
		return "Unknown"
	}
}

// Helper function to extract SDF filters from data paths
func extractSdfFiltersFromDataPaths(smContext *smf_context.SMContext) map[uint8]string {
	sdfFilters := make(map[uint8]string)

	if smContext.Tunnel == nil || smContext.Tunnel.DataPathPool == nil {
		return sdfFilters
	}

	for _, dataPath := range smContext.Tunnel.DataPathPool {
		for node := dataPath.FirstDPNode; node != nil; node = node.Next() {
			if node.UPF == nil {
				continue
			}

			// Check UpLink PDR
			if node.UpLinkTunnel != nil && node.UpLinkTunnel.PDR != nil {
				pdr := node.UpLinkTunnel.PDR
				if pdr.PDI.SDFFilter != nil {
					// Extract QFI from QERs associated with this PDR
					for _, qer := range pdr.QER {
						if qer != nil {
							sdfFilters[qer.QFI.QFI] = convertSdfFilter(string(pdr.PDI.SDFFilter.FlowDescription))
						}
					}
				}
			}

			// Check DownLink PDR
			if node.DownLinkTunnel != nil && node.DownLinkTunnel.PDR != nil {
				pdr := node.DownLinkTunnel.PDR
				if pdr.PDI.SDFFilter != nil {
					// Extract QFI from QERs associated with this PDR
					for _, qer := range pdr.QER {
						if qer != nil {
							sdfFilters[qer.QFI.QFI] = convertSdfFilter(string(pdr.PDI.SDFFilter.FlowDescription))
						}
					}
				}
			}
		}
	}

	return sdfFilters
}

// Helper function to create QoS flow info from additional QoS flows
func createQosFlowFromAdditional(qosFlow *smf_context.QoSFlow, sdfFilter string) QosFlowInfo {
	qosFlowInfo := QosFlowInfo{
		State:     getQosFlowStateString(qosFlow.State),
		IsGBRFlow: qosFlow.IsGBRFlow(),
		SdfFilter: sdfFilter,
	}

	if qosFlow.QoSProfile != nil {
		qosFlowInfo.Var5qi = qosFlow.QoSProfile.Var5qi

		// Set MBR values
		if qosFlow.QoSProfile.MaxbrUl != "" {
			qosFlowInfo.MaxbrUl = qosFlow.QoSProfile.MaxbrUl
		}
		if qosFlow.QoSProfile.MaxbrDl != "" {
			qosFlowInfo.MaxbrDl = qosFlow.QoSProfile.MaxbrDl
		}

		// Set GBR values if it's a GBR flow
		if qosFlow.IsGBRFlow() {
			if qosFlow.QoSProfile.GbrUl != "" {
				qosFlowInfo.GbrUl = qosFlow.QoSProfile.GbrUl
			}
			if qosFlow.QoSProfile.GbrDl != "" {
				qosFlowInfo.GbrDl = qosFlow.QoSProfile.GbrDl
			}
		}
	}

	return qosFlowInfo
}

// Helper function to create default QoS flow info
func createDefaultQosFlow(smContext *smf_context.SMContext, sdfFilter string) QosFlowInfo {
	qosFlowInfo := QosFlowInfo{
		State:     "Default",
		IsGBRFlow: false,
		SdfFilter: sdfFilter,
	}

	if smContext.DnnConfiguration.Var5gQosProfile != nil {
		qosFlowInfo.Var5qi = smContext.DnnConfiguration.Var5gQosProfile.Var5qi

		// Set AMBR values from DnnConfiguration
		if smContext.DnnConfiguration.SessionAmbr != nil {
			qosFlowInfo.MaxbrUl = smContext.DnnConfiguration.SessionAmbr.Uplink
			qosFlowInfo.MaxbrDl = smContext.DnnConfiguration.SessionAmbr.Downlink
		}
	}

	return qosFlowInfo
}

func (p *Processor) HandleOAMGetUEPDUSessionInfo(c *gin.Context) {
	// Get all SM contexts from the pool
	allSMContexts := smf_context.GetAllSMContexts()

	if len(allSMContexts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No SM contexts found"})
		return
	}

	// Convert all SM contexts to PDU session info
	allPduSessionInfos := make(map[string]*PDUSessionInfo)

	for ref, smContext := range allSMContexts {
		pduSessionInfo := &PDUSessionInfo{
			Supi:         smContext.Supi,
			PDUSessionID: strconv.Itoa(int(smContext.PDUSessionID)),
			QosFlows:     make(map[string]QosFlowInfo),
		}

		// Handle PDUAddress safely
		if smContext.PDUAddress != nil {
			pduSessionInfo.PDUAddress = smContext.PDUAddress.String()
		} else {
			pduSessionInfo.PDUAddress = ""
		}

		// Extract SDF filters from data paths
		sdfFilters := extractSdfFiltersFromDataPaths(smContext)

		// Extract QoS flow information from AdditonalQosFlows
		if smContext.AdditonalQosFlows != nil {
			for qfi, qosFlow := range smContext.AdditonalQosFlows {
				if qosFlow != nil {
					sdfFilter := sdfFilters[qfi]
					qosFlowInfo := createQosFlowFromAdditional(qosFlow, sdfFilter)
					pduSessionInfo.QosFlows[strconv.Itoa(int(qfi))] = qosFlowInfo
				}
			}
		}

		// Extract default QoS flow information from DnnConfiguration if available
		if smContext.DnnConfiguration.Var5gQosProfile != nil {
			sdfFilter := sdfFilters[1] // Default QFI is typically 1
			defaultQosFlowInfo := createDefaultQosFlow(smContext, sdfFilter)
			pduSessionInfo.QosFlows["1"] = defaultQosFlowInfo
		}

		allPduSessionInfos[ref] = pduSessionInfo
	}

	// Return response with pool size and all contexts
	response := gin.H{
		"poolSize":      smf_context.GetSMContextPoolSize(),
		"smContexts":    allPduSessionInfos,
		"totalContexts": len(allSMContexts),
	}

	c.JSON(http.StatusOK, response)
}

func (p *Processor) HandleGetSMFDebugInfo(c *gin.Context) {
	ue1Supi := "imsi-208930000000001"
	PduSessionId := int32(1)

	type debugInfo struct {
		PCCRulesMap     map[string]smf_context.PCCRule `json:"pccRulesMap,omitempty"`
		ChargingDataMap map[string]models.ChargingData `json:"chargingDataMap,omitempty"`
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
