package types

import (
	"github.com/projecteru2/core/types"
)

const supported = 3

// ResourceRequests .
type ResourceRequests [supported]ResourceRequest

// ResourceRequirement .
type ResourceRequest interface {
	Type() types.ResourceType
	Validate() error
	MakeScheduler() SchedulerV2
	Rate(types.Node) float64
}

// SchedulerV2 .
type SchedulerV2 func([]ScheduleInfo) (ResourcePlans, int, error)

// DispenseOptions .
type DispenseOptions struct {
	*types.Node
	Index            int
	ExistingInstance *types.Workload
}

// ResourcePlans .
type ResourcePlans interface {
	Type() types.ResourceType
	Capacity() map[string]int
	ApplyChangesOnNode(*types.Node, ...int)
	RollbackChangesOnNode(*types.Node, ...int)
	Dispense(DispenseOptions, *types.ResourceMeta) (*types.ResourceMeta, error)
}

type ScheduleInfo struct {
	Name          string
	CPUMap        types.CPUMap
	VolumeMap     types.VolumeMap
	InitVolumeMap types.VolumeMap
	NUMA          types.NUMA
	NUMAMemory    types.NUMAMemory
	MemCap        int64
	StorageCap    int64

	CPUPlan     []types.CPUMap
	VolumePlans []types.VolumePlan // {{"AUTO:/data:rw:1024": "/mnt0:/data:rw:1024"}}
	Capacity    int
}
