package observiqreceiver

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	observiq "github.com/observiq/carbon/agent"
	obsentry "github.com/observiq/carbon/entry"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.uber.org/zap"
)

type observiqReceiver struct {
	startOnce sync.Once
	stopOnce  sync.Once
	done      chan struct{}

	config   *Config
	agent    *observiq.LogAgent
	logsChan chan obsentry.Entry
	consumer consumer.LogsConsumer
	logger   *zap.Logger
}

// Ensure this factory adheres to required interface
var _ component.LogsReceiver = (*observiqReceiver)(nil)

// Start tells the receiver to start
func (r *observiqReceiver) Start(ctx context.Context, host component.Host) error {
	err := componenterror.ErrAlreadyStarted
	r.startOnce.Do(func() {
		err = nil

		r.agent.Start()

		go func() {
			for {
				select {
				case <-r.done:
					return
				case obsLog := <-r.logsChan:
					// TODO batch logs in agent output and handle mapping of slice here
					if err := r.consumer.ConsumeLogs(ctx, convert(obsLog)); err != nil {
						// TODO determine handling for unhandleable but non-fatal error
					}
				}
			}
		}()
	})

	return err
}

func convert(obsLog obsentry.Entry) pdata.Logs {
	out := pdata.NewLogs()
	logs := out.ResourceLogs()
	logs.Resize(1)
	rls := logs.At(0)
	rls.Resource().InitEmpty()
	rls.InstrumentationLibraryLogs().Resize(1)
	logSlice := rls.InstrumentationLibraryLogs().At(0).Logs()

	lr := pdata.NewLogRecord()
	lr.InitEmpty()
	lr.Body().InitEmpty()

	lr.SetTimestamp(pdata.TimestampUnixNano(obsLog.Timestamp.UnixNano()))
	sevText, sevNum := convertSeverity(obsLog.Severity)
	lr.SetSeverityText(sevText)
	lr.SetSeverityNumber(sevNum)

	attributes := lr.Attributes()
	for lblKey, lblVal := range obsLog.Labels {
		attributes.InsertString(lblKey, lblVal)
	}

	insertToAttributeVal(obsLog.Record, lr.Body())
	logSlice.Append(&lr)

	return out
}

// Shutdown is invoked during service shutdown
func (r *observiqReceiver) Shutdown(context.Context) error {
	r.stopOnce.Do(func() {
		close(r.done)
	})
	return nil
}

func insertToAttributeVal(value interface{}, dest pdata.AttributeValue) {
	switch t := value.(type) {
	case bool:
		dest.SetBoolVal(t)
	case string:
		dest.SetStringVal(t)
	case []byte:
		dest.SetStringVal(string(t))
	case int64:
		dest.SetIntVal(t)
	case int32:
		dest.SetIntVal(int64(t))
	case int16:
		dest.SetIntVal(int64(t))
	case int8:
		dest.SetIntVal(int64(t))
	case uint64:
		dest.SetIntVal(int64(t))
	case uint32:
		dest.SetIntVal(int64(t))
	case uint16:
		dest.SetIntVal(int64(t))
	case uint8:
		dest.SetIntVal(int64(t))
	case float64:
		dest.SetDoubleVal(t)
	case float32:
		dest.SetDoubleVal(float64(t))
	case map[string]interface{}:
		dest.SetMapVal(toAttributeMap(t))
	case []interface{}:
		dest.SetMapVal(toAttributeMap(sliceToMap(t)))
	default:
		dest.SetStringVal(fmt.Sprintf("%v", t))
	}
}

func toAttributeMap(obsMap map[string]interface{}) pdata.AttributeMap {
	attMap := pdata.NewAttributeMap()
	for k, v := range obsMap {
		switch t := v.(type) {
		case bool:
			attMap.InsertBool(k, t)
		case string:
			attMap.InsertString(k, t)
		case []byte:
			attMap.InsertString(k, string(t))
		case int64:
			attMap.InsertInt(k, t)
		case int32:
			attMap.InsertInt(k, int64(t))
		case int16:
			attMap.InsertInt(k, int64(t))
		case int8:
			attMap.InsertInt(k, int64(t))
		case uint64:
			attMap.InsertInt(k, int64(t))
		case uint32:
			attMap.InsertInt(k, int64(t))
		case uint16:
			attMap.InsertInt(k, int64(t))
		case uint8:
			attMap.InsertInt(k, int64(t))
		case float64:
			attMap.InsertDouble(k, t)
		case float32:
			attMap.InsertDouble(k, float64(t))
		case map[string]interface{}:
			subMap := toAttributeMap(t)
			subMapVal := pdata.NewAttributeValueMap()
			subMapVal.SetMapVal(subMap)
			attMap.Insert(k, subMapVal)
		case []interface{}:
			subMap := toAttributeMap(sliceToMap(t))
			subMapVal := pdata.NewAttributeValueMap()
			subMapVal.SetMapVal(subMap)
			attMap.Insert(k, subMapVal)
		default:
			attMap.InsertString(k, fmt.Sprintf("%v", t))
		}
	}
	return attMap
}

// This returns a map of stringified index to value,
// rather than an array because the pdata package does not support arrays
func sliceToMap(arr []interface{}) map[string]interface{} {
	sliceAsMap := make(map[string]interface{})
	for i, v := range arr {
		sliceAsMap[strconv.Itoa(i)] = v
	}
	return sliceAsMap
}

func convertSeverity(s obsentry.Severity) (string, pdata.SeverityNumber) {
	switch {
	case s == obsentry.Catastrophe:
		return "Fatal", pdata.SeverityNumberFATAL4

	case s > obsentry.Emergency:
		return "Fatal", pdata.SeverityNumberFATAL2
	case s == obsentry.Emergency:
		return "Error", pdata.SeverityNumberFATAL

	case s > obsentry.Alert:
		return "Error", pdata.SeverityNumberERROR4
	case s == obsentry.Alert:
		return "Error", pdata.SeverityNumberERROR3

	case s > obsentry.Critical:
		return "Error", pdata.SeverityNumberERROR3
	case s == obsentry.Critical:
		return "Error", pdata.SeverityNumberERROR2

	case s > obsentry.Error:
		return "Error", pdata.SeverityNumberERROR2
	case s == obsentry.Error:
		return "Error", pdata.SeverityNumberERROR

	case s > obsentry.Warning:
		return "Info", pdata.SeverityNumberINFO4
	case s == obsentry.Warning:
		return "Info", pdata.SeverityNumberINFO4

	case s > obsentry.Notice:
		return "Info", pdata.SeverityNumberINFO3
	case s == obsentry.Notice:
		return "Info", pdata.SeverityNumberINFO3

	case s > obsentry.Info:
		return "Info", pdata.SeverityNumberINFO2
	case s == obsentry.Info:
		return "Info", pdata.SeverityNumberINFO

	case s > obsentry.Debug:
		return "Debug", pdata.SeverityNumberDEBUG2
	case s == obsentry.Debug:
		return "Debug", pdata.SeverityNumberDEBUG

	case s > obsentry.Trace:
		return "Trace", pdata.SeverityNumberTRACE3
	case s == obsentry.Trace:
		return "Trace", pdata.SeverityNumberTRACE2

	case s > obsentry.Default:
		return "Trace", pdata.SeverityNumberTRACE
	default:
		return "Undefined", pdata.SeverityNumberUNDEFINED
	}
}
