# SMF OAM API Documentation

This document describes the Operations, Administration, and Maintenance (OAM) API endpoints available in the SMF (Session Management Function).

## Base URL
The OAM API is available at: `http://<smf-host>:<oam-port>/nsmf-oam/v1`

For the SMF running at 127.0.0.2:8000, the base URL is: `http://127.0.0.2:8000/nsmf-oam/v1`

## API Endpoints

### 1. Health Check

**GET /nsmf-oam/v1/** 

Check if the SMF service is available.

#### Response
- **200 OK**: Service is available
  ```json
  {
    "status": "Service Available"
  }
  ```

---

### 2. Get All UE PDU Session Information

**GET /nsmf-oam/v1/ue-pdu-session-info/**

Retrieve detailed information about all UE PDU sessions currently managed by the SMF, including QoS flow information.

#### Parameters
None required.

#### Response
- **200 OK**: All PDU Session information retrieved successfully
  ```json
  {
    "poolSize": 1,
    "smContexts": {
      "urn:uuid:ed4ff6a5-1c94-47b2-94b3-6ffffea7a4a3": {
        "supi": "imsi-208930000000001",
        "pduSessionId": "1",
        "pduAddress": "10.60.0.1",
        "qosFlows": {
          "1": {
            "5qi": 9,
            "state": "Default",
            "isGbrFlow": false,
            "maxbrUl": "1000 Mbps",
            "maxbrDl": "1000 Mbps",
            "sdfFilter": "1.1.1.1/32"
          },
          "2": {
            "5qi": 8,
            "state": "Set",
            "isGbrFlow": false,
            "maxbrUl": "208 Mbps",
            "maxbrDl": "208 Mbps",
            "sdfFilter": "1.1.1.1/32"
          }
        }
      }
    },
    "totalContexts": 1
  }
  ```
- **404 Not Found**: No SM contexts found
  ```json
  {
    "message": "No SM contexts found"
  }
  ```

#### Response Fields Description
- **poolSize**: Total number of SM contexts in the pool
- **totalContexts**: Number of contexts returned in this response
- **smContexts**: Map of SM context references to PDU session information
  - **Key**: SM Context Reference (UUID format)
  - **Value**: PDU Session Information object

#### PDU Session Information Fields
- **supi**: Subscription Permanent Identifier
- **pduSessionId**: PDU Session Identifier (as string)
- **pduAddress**: Assigned UE IP address (empty string if not allocated)
- **qosFlows**: Map of QoS Flow Identifier to QoS Flow Information
  - **Key**: QFI (QoS Flow Identifier) as string
  - **Value**: QoS Flow Information object

#### QoS Flow Information Fields
- **5qi**: 5G QoS Identifier (integer)
- **state**: QoS Flow state ("Default", "Set", "Unset", "ToBeModify", "Unknown")
- **isGbrFlow**: Boolean indicating if this is a Guaranteed Bit Rate flow
- **maxbrUl**: Maximum Bit Rate Uplink (optional, e.g., "208 Mbps")
- **maxbrDl**: Maximum Bit Rate Downlink (optional, e.g., "208 Mbps")
- **gbrUl**: Guaranteed Bit Rate Uplink (optional, only for GBR flows, e.g., "108 Mbps")
- **gbrDl**: Guaranteed Bit Rate Downlink (optional, only for GBR flows, e.g., "108 Mbps")
- **sdfFilter**: Service Data Flow filter (converted format, e.g., "1.1.1.1/32" or "any")

#### SDF Filter Format
The SDF filter field is automatically converted from the internal PFCP format:
- `permit out ip from 1.1.1.1/32 to assigned` → `1.1.1.1/32`
- `permit out ip from any to assigned` → `any`
- Other formats are returned as-is

#### Notes
- This endpoint returns all active SM contexts with their associated QoS flows
- QoS flows include both additional flows and default flows from DNN configuration
- SDF filters are extracted from PDR (Packet Detection Rules) in the data path
- GBR fields are only present for Guaranteed Bit Rate flows
- If no active sessions exist, a 404 response with a message is returned

---

### 3. Get SMF User Plane Information (Debug)

**GET /nsmf-oam/v1/user-plane-info-debug/**

Retrieve debugging information about SMF user plane for a specific UE (development/debug use only).

#### Response
- **200 OK**: Debug information retrieved successfully
  ```json
  {
    "pccRulesMap": {
      "rule_id_1": {
        "PccRuleId": "rule_id_1",
        "FlowInfos": [...],
        "Precedence": 10,
        // Other PCC rule fields
      }
    },
    "chargingDataMap": {
      "charging_id_1": {
        "ChgId": "charging_id_1",
        "ChgMethod": "ONLINE",
        // Other charging data fields  
      }
    }
  }
  ```
- **404 Not Found**: UE context not found

#### Fields Description
- **pccRulesMap**: Policy and Charging Control rules mapping
- **chargingDataMap**: Charging data information mapping

#### Note
This endpoint is hardcoded to query UE with SUPI "imsi-208930000000001" and PDU Session ID 1 for debugging purposes.

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "Invalid request parameters"
}
```

### 404 Not Found  
```json
{
  "error": "Resource not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

---

## Usage Examples

### Check Service Status
```bash
curl -X GET http://127.0.0.2:8000/nsmf-oam/v1/
```

### Get All UE PDU Session Info
```bash
curl -X GET http://127.0.0.2:8000/nsmf-oam/v1/ue-pdu-session-info/
```

#### Example Response
```bash
# When there are active sessions with QoS flows
curl -X GET http://127.0.0.2:8000/nsmf-oam/v1/ue-pdu-session-info/ | jq
{
  "poolSize": 1,
  "smContexts": {
    "urn:uuid:ed4ff6a5-1c94-47b2-94b3-6ffffea7a4a3": {
      "supi": "imsi-208930000000001",
      "pduSessionId": "1",
      "pduAddress": "10.60.0.1",
      "qosFlows": {
        "1": {
          "5qi": 9,
          "state": "Default",
          "isGbrFlow": false,
          "maxbrUl": "1000 Mbps",
          "maxbrDl": "1000 Mbps",
          "sdfFilter": "1.1.1.1/32"
        },
        "2": {
          "5qi": 8,
          "state": "Set",
          "isGbrFlow": false,
          "maxbrUl": "208 Mbps",
          "maxbrDl": "208 Mbps",
          "sdfFilter": "1.1.1.1/32"
        }
      }
    }
  },
  "totalContexts": 1
}

# When there are no active sessions
curl -X GET http://127.0.0.2:8000/nsmf-oam/v1/ue-pdu-session-info/
{
  "message": "No SM contexts found"
}
```

### Get Debug Information
```bash
curl -X GET http://127.0.0.2:8000/nsmf-oam/v1/user-plane-info-debug/
```

## 重要配置說明

### API Prefix 配置
SMF OAM API 使用固定的 prefix: `/nsmf-oam/v1`

### 服務啟用檢查
要使 OAM API 正常工作，需要確認以下配置：

1. **ServiceNameList 配置**: 在 SMF 配置文件中，`serviceNameList` 必須包含 `nsmf-oam`
   ```yaml
   configuration:
     serviceNameList:
       - nsmf-pdusession
       - nsmf-event-exposure
       - nsmf-oam  # 必須包含此項
   ```

2. **基本測試**: 
   ```bash
   # 健康檢查（應該返回 {"status":"Service Available"}）
   curl -X GET http://127.0.0.2:8000/nsmf-oam/v1/
   
   # 如果返回 404 page not found，檢查：
   # - 服務是否正在運行
   # - serviceNameList 是否包含 nsmf-oam
   # - 是否使用了正確的 URL prefix
   ```

3. **常見問題排查**:
   - `404 page not found`: 檢查是否使用了正確的 prefix `/nsmf-oam/v1`
   - `null` 回應: API 正常工作，但沒有相關數據（例如沒有活躍的 SM Context）
   - `{}` 空物件: API 正常工作，但集合為空（例如沒有使用報告）

---

## Notes

1. The OAM API is intended for operational monitoring and troubleshooting purposes.
2. The debug endpoint (`/user-plane-info-debug/`) should only be used in development/testing environments.
3. The `/ue-pdu-session-info/` endpoint provides comprehensive QoS flow information extracted from SMF context and PFCP data paths.
4. SDF filters are automatically converted from PFCP format to a simplified format for better readability.
5. QoS flow information includes both additional flows (from `AdditonalQosFlows`) and default flows (from DNN configuration).
6. Bit rate values are formatted as human-readable strings (e.g., "208 Mbps").
7. GBR (Guaranteed Bit Rate) fields are only included for flows where `isGbrFlow` is true.
8. All QFI (QoS Flow Identifier) keys in the `qosFlows` map are strings for JSON compatibility.