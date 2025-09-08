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

Retrieve detailed information about all UE PDU sessions currently managed by the SMF.

#### Parameters
None required.

#### Response
- **200 OK**: All PDU Session information retrieved successfully
  ```json
  {
    "poolSize": 3,
    "totalContexts": 3,
    "smContexts": {
      "urn:uuid:12345678-1234-1234-1234-123456789abc": {
        "Supi": "imsi-208930000000001",
        "PDUSessionID": "1",
        "Dnn": "internet",
        "Sst": "1",
        "Sd": "010203",
        "AnType": "3GPP_ACCESS",
        "PDUAddress": "10.60.0.1",
        "UpCnxState": "ACTIVATED"
      },
      "urn:uuid:87654321-4321-4321-4321-cba987654321": {
        "Supi": "imsi-208930000000002",
        "PDUSessionID": "1",
        "Dnn": "ims",
        "Sst": "2",
        "Sd": "020304",
        "AnType": "3GPP_ACCESS",
        "PDUAddress": "10.60.0.2",
        "UpCnxState": "ACTIVATED"
      }
    }
  }
  ```
- **404 Not Found**: No SM contexts found
  ```json
  {
    "message": "No SM contexts found"
  }
  ```

#### Response Fields Description
- **poolSize**: Total number of SM contexts in the pool (from sync.Map.Range count)
- **totalContexts**: Number of contexts returned in this response
- **smContexts**: Map of SM context references to PDU session information
  - **Key**: SM Context Reference (UUID format)
  - **Value**: PDU Session Information object

#### PDU Session Information Fields
- **Supi**: Subscription Permanent Identifier
- **PDUSessionID**: PDU Session Identifier (as string)
- **Dnn**: Data Network Name
- **Sst**: Slice/Service Type (as string)
- **Sd**: Slice Differentiator
- **AnType**: Access Network Type (3GPP_ACCESS, NON_3GPP_ACCESS)
- **PDUAddress**: Assigned UE IP address (empty string if not allocated)
- **UpCnxState**: User Plane Connection State (ACTIVATED, DEACTIVATED, etc.)

#### Notes
- This endpoint returns all active SM contexts in the SMF's context pool
- If no active sessions exist, a 404 response with a message is returned
- The PDUAddress field will be an empty string if the IP address is not yet allocated
- The endpoint uses the `GetAllSMContexts()` and `GetSMContextPoolSize()` helper functions

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
# When there are active sessions
curl -X GET http://127.0.0.2:8000/nsmf-oam/v1/ue-pdu-session-info/ | jq
{
  "poolSize": 2,
  "totalContexts": 2,
  "smContexts": {
    "urn:uuid:12345678-1234-1234-1234-123456789abc": {
      "Supi": "imsi-208930000000001",
      "PDUSessionID": "1",
      "Dnn": "internet",
      "Sst": "1",
      "Sd": "010203",
      "AnType": "3GPP_ACCESS",
      "PDUAddress": "10.60.0.1",
      "UpCnxState": "ACTIVATED"
    }
  }
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
3. Usage reports are collected and cleared after each retrieval to prevent memory accumulation.
4. All timestamps are in RFC3339 format (ISO 8601).
5. Volume measurements are in bytes, packet counts are in number of packets.