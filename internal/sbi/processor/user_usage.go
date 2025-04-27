package processor

import (
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/internal/logger"
)

func (p *Processor) StoreUserUsageInfo(seid uint64) {

	smContext := smf_context.GetSMContextBySEID(seid)
	supi := smContext.Supi
	if supi == "" {
		logger.ProcessorLog.Errorf("cannot find SUPI for SEID[%d]\n", seid)
		return
	}
	logger.ProcessorLog.Infof("StoreUserUsageInfo: SEID[%d] SUPI[%s]\n", seid, supi)

	if smContext == nil {
		logger.ProcessorLog.Errorf("StoreUserUsageInfo: SMContext not found for SEID[%d]\n", seid)
		return
	}
	if smContext.UrrReports == nil {
		logger.ProcessorLog.Errorf("StoreUserUsageInfo: SMContext UrrReports is nil for SEID[%d]\n", seid)
		return
	}

	// DEBUG ---------------
	logger.ProcessorLog.Errorf("StoreUserUsageInfo: SEID[%d] SUPI[%s]\n", seid, supi)
	for id, pccRule := range smContext.PCCRules {
		logger.ProcessorLog.Errorf("PCCRule[%s] = %+v \n", id, pccRule) // 使用 %+v 顯示屬性名稱
	}
	for id, sessionRule := range smContext.SessionRules {
		logger.ProcessorLog.Errorf("SessionRule[%s] : DefQosQFI = %v ", id, sessionRule.DefQosQFI) // 使用 %+v 顯示屬性名稱
		logger.ProcessorLog.Errorf("AuthSessAmbr = %+v\n", sessionRule.AuthSessAmbr) // 使用 %+v 顯示屬性名稱
		logger.ProcessorLog.Errorf("AuthDefQos = %+v\n", sessionRule.AuthDefQos) // 使用 %+v 顯示屬性名稱
	}
	for id, chargingData := range smContext.ChargingData {
		logger.ProcessorLog.Errorf("ChargingData[%s] = %+v \n", id, chargingData) // 使用 %+v 顯示屬性名稱
	}
	for id, qosData := range smContext.QosDatas {
		logger.ProcessorLog.Errorf("QosData[%s] = %+v \n", id, qosData) // 使用 %+v 顯示屬性名稱
	}

	logger.ProcessorLog.Errorf("UrrIdMap = %+v \n", smContext.UrrIdMap) // 使用 %+v 顯示屬性名稱

	// ---------------------

	usageReports, exists := p.UsageReports[supi]
	if !exists {
		// Initialize the usage reports for this SUPI
		usageReports = make(UsageReportPerUE, 0)
		for _, sessionRule := range smContext.SessionRules {
			usageReports[sessionRule.DefQosQFI] = make([]smf_context.UsageReport, 0) 
		}
		for _, pccRule := range smContext.PCCRules {
			usageReports[pccRule.QFI] = make([]smf_context.UsageReport, 0)
		}
		p.UsageReports[supi] = usageReports
	}

	// Directly modify p.UsageReports[supi] to ensure changes are applied
	for _, report := range smContext.UrrReports {
		switch {
		case report.UrrId == 7:
			p.UsageReports[supi][0] = append(p.UsageReports[supi][0], report) //TODO: (just hardcoding now) default QFI
		case report.UrrId == 8:
			p.UsageReports[supi][2] = append(p.UsageReports[supi][2], report) //TODO: (just hardcoding now) default QFI
		default:
			// Handle other cases if needed
		}
	}

	return
}
