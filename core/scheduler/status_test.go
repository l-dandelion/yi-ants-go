package scheduler

import (
	"testing"
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

func TestCheckStatus(t *testing.T) {
	var currentStatus, wantedStatus int8
	var currentStatusList, wantedStatusList []int8

	//initializing\starting\pausing\stopping
	currentStatusList = []int8{
		constant.RUNNING_STATUS_PREPARING,
		constant.RUNNING_STATUS_STARTING,
		constant.RUNNING_STATUS_PAUSING,
		constant.RUNNING_STATUS_STOPPING,
	}
	wantedStatus = constant.RUNNING_STATUS_PREPARING
	for _, currentStatus := range currentStatusList {
		if yierr := checkStatus(currentStatus, wantedStatus); yierr == nil {
			t.Fatalf("It still can check status with incorrect current status %q!",
				GetStatusDescription(currentStatus))
		}
	}

	// wanted status should be initializing, starting, pausing or stopping
	currentStatus = constant.RUNNING_STATUS_UNPREPARED
	wantedStatusList = []int8{
		constant.RUNNING_STATUS_UNPREPARED,
		constant.RUNNING_STATUS_PREPARED,
		constant.RUNNING_STATUS_STARTED,
		constant.RUNNING_STATUS_PAUSED,
		constant.RUNNING_STATUS_STOPPED,
	}
	for _, wantedStatus := range wantedStatusList {
		if yierr := checkStatus(currentStatus, wantedStatus); yierr == nil {
			t.Fatalf("It still can check status with incorrect wanted status %q!",
				GetStatusDescription(wantedStatus))
		}
	}

	//uninitialized can't -> starting, pausing, stopping
	currentStatus = constant.RUNNING_STATUS_UNPREPARED
	wantedStatusList = []int8{
		constant.RUNNING_STATUS_STARTING,
		constant.RUNNING_STATUS_PAUSING,
		constant.RUNNING_STATUS_STOPPING,
	}
	for _, wantedStatus := range wantedStatusList {
		if yierr := checkStatus(currentStatus, wantedStatus); yierr == nil {
			t.Fatalf("It still can check status with current status %q wanted status %q!",
				GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
		}
	}
	wantedStatus = constant.RUNNING_STATUS_PREPARING
	if yierr := checkStatus(currentStatus, wantedStatus); yierr != nil {
		t.Fatalf("An error occurs when checking status: %s (currentStatus: %q, wantedStatus: %q)!",
			yierr, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
	}

	//started can't -> initializing, starting
	currentStatus = constant.RUNNING_STATUS_STARTED
	wantedStatusList = []int8{
		constant.RUNNING_STATUS_PREPARING,
		constant.RUNNING_STATUS_STARTING,
	}
	for _, wantedStatus := range wantedStatusList {
		if yierr := checkStatus(currentStatus, wantedStatus); yierr == nil {
			t.Fatalf("It still can check status with current status %q wanted status %q!",
				GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
		}
	}
	wantedStatusList = []int8{
		constant.RUNNING_STATUS_STOPPING,
		constant.RUNNING_STATUS_PAUSING,
	}
	for _, wantedStatus := range wantedStatusList {
		if yierr := checkStatus(currentStatus, wantedStatus); yierr != nil {
			t.Fatalf("An error occurs when checking status: %s (currentStatus: %q, wantedStatus: %q)!",
				yierr, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
		}
	}

	// !(started or paused) can't to stopping
	currentStatusList = []int8{
		constant.RUNNING_STATUS_UNPREPARED,
		constant.RUNNING_STATUS_PREPARING,
		constant.RUNNING_STATUS_PREPARED,
		constant.RUNNING_STATUS_STARTING,
		constant.RUNNING_STATUS_PAUSING,
		constant.RUNNING_STATUS_STOPPING,
		constant.RUNNING_STATUS_STOPPED,
	}
	wantedStatus = constant.RUNNING_STATUS_STOPPING
	for _, currentStatus := range currentStatusList {
		if yierr := checkStatus(currentStatus, wantedStatus); yierr == nil {
			t.Fatalf("It still can check status with current status %q wanted status %q!",
				GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
		}
	}
	currentStatus = constant.RUNNING_STATUS_STARTED
	if err := checkStatus(currentStatus, wantedStatus); err != nil {
		t.Fatalf("An error occurs when checking status: %s (currentStatus: %q, wantedStatus: %q)!",
			err, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
	}
	currentStatus = constant.RUNNING_STATUS_PAUSED
	if err := checkStatus(currentStatus, wantedStatus); err != nil {
		t.Fatalf("An error occurs when checking status: %s (currentStatus: %q, wantedStatus: %q)!",
			err, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
	}

	// !(started) can't to pausing
	currentStatusList = []int8{
		constant.RUNNING_STATUS_UNPREPARED,
		constant.RUNNING_STATUS_PREPARING,
		constant.RUNNING_STATUS_PREPARED,
		constant.RUNNING_STATUS_STARTING,
		constant.RUNNING_STATUS_PAUSING,
		constant.RUNNING_STATUS_PAUSED,
		constant.RUNNING_STATUS_STOPPING,
		constant.RUNNING_STATUS_STOPPED,
	}
	wantedStatus = constant.RUNNING_STATUS_PAUSING
	for _, currentStatus := range currentStatusList {
		if yierr := checkStatus(currentStatus, wantedStatus); yierr == nil {
			t.Fatalf("It still can check status with current status %q wanted status %q!",
				GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
		}
	}
	currentStatus = constant.RUNNING_STATUS_STARTED
	if err := checkStatus(currentStatus, wantedStatus); err != nil {
		t.Fatalf("An error occurs when checking status: %s (currentStatus: %q, wantedStatus: %q)!",
			err, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
	}
}
