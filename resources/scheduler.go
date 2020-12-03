package resources

import (
	"math"

	resourcetypes "github.com/projecteru2/core/resources/types"
	"github.com/projecteru2/core/types"
	"github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"
)

// SelectNodesByResourceRequests select nodes by resource requests
func SelectNodesByResourceRequests(resourceRequests resourcetypes.ResourceRequests, nodeMap map[string]*types.Node) (
	scheduleType types.ResourceType,
	total int,
	plans []resourcetypes.ResourcePlans,
	err error,
) {
	total = math.MaxInt64
	scheduleInfos := []resourcetypes.ScheduleInfo{}
	for _, node := range nodeMap {
		scheduleInfo := resourcetypes.ScheduleInfo{
			Name:          node.Name,
			CPUMap:        node.CPU,
			VolumeMap:     node.Volume,
			InitVolumeMap: node.InitVolume,
			NUMA:          node.NUMA,
			NUMAMemory:    node.NUMAMemory,
			MemCap:        node.MemCap,
			StorageCap:    node.StorageCap,
			Capacity:      0,
		}
		scheduleInfos = append(scheduleInfos, scheduleInfo)
	}
	log.Debugf("[SelectNodesByResourceRequests] nodesInfo: %+v", scheduleInfos)
	for _, resourceRequest := range resourceRequests {
		plan, subTotal, err := resourceRequest.MakeScheduler()(scheduleInfos)
		if err != nil {
			return scheduleType, total, plans, err
		}
		plans = append(plans, plan)
		total = utils.Min(total, subTotal)

		// calculate schedule type
		if resourceRequest.Type()&types.ResourceCPUBind != 0 {
			scheduleType |= types.ResourceCPU
		}
		if resourceRequest.Type()&types.ResourceScheduledVolume != 0 {
			scheduleType |= types.ResourceVolume
		}
	}

	if scheduleType == 0 {
		scheduleType = types.ResourceMemory
	}
	return // nolint:nakedret
}
