package complexscheduler

import (
	"fmt"
	"math/rand"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab.ricebook.net/platform/core/types"
)

func resultLength(result map[string][]types.CPUMap) int {
	length := 0
	for _, list := range result {
		length += len(list)
	}
	return length
}

func TestSelectCPUNodes(t *testing.T) {
	coreCfg := types.Config{
		EtcdMachines:   []string{"http://127.0.0.1:2379"},
		EtcdLockPrefix: "/eru-core/_lock",
		Scheduler: types.SchedConfig{
			LockKey: "/coretest",
			LockTTL: 1,
			Type:    "complex",
		},
	}

	k, merr := New(coreCfg)
	if merr != nil {
		t.Fatalf("Create Potassim error: %v", merr)
	}

	_, _, err := k.SelectCPUNodes(map[string]types.CPUMap{}, 1, 1)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "No nodes provide to choose some")

	nodes := map[string]types.CPUMap{
		"node1": types.CPUMap{
			"0": 10,
			"1": 10,
		},
		"node2": types.CPUMap{
			"0": 10,
			"1": 10,
		},
	}

	debug, changed, _ := k.SelectCPUNodes(nodes, 0.5, 1)
	fmt.Printf("algorithm debug res: %v", debug)
	fmt.Printf("algorithm debug changed: %v", changed)

	_, _, err = k.SelectCPUNodes(nodes, 2, 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not enough")

	_, _, err = k.SelectCPUNodes(nodes, 3, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not enough")

	_, _, err = k.SelectCPUNodes(nodes, 1, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not enough")

	r, re, err := k.SelectCPUNodes(nodes, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, 2, len(re))

	for nodename, cpus := range r {
		assert.Contains(t, []string{"node1", "node2"}, nodename)
		// assert.Equal(t, len(cpus), 1)

		cpu := cpus[0]
		assert.Equal(t, cpu.Total(), 10)
	}

	// SelectCPUNodes 里有一些副作用, 粗暴地拿一个新的来测试吧
	// 下面也是因为这个
	nodes = map[string]types.CPUMap{
		"node1": types.CPUMap{
			"0": 10,
			"1": 10,
		},
		"node2": types.CPUMap{
			"0": 10,
			"1": 10,
		},
	}

	r, re, err = k.SelectCPUNodes(nodes, 1.3, 2)
	assert.NoError(t, err)

	for nodename, cpus := range r {
		assert.Contains(t, []string{"node1", "node2"}, nodename)
		assert.Equal(t, len(cpus), 1)

		cpu := cpus[0]
		assert.Equal(t, cpu.Total(), 13)
	}
}

func checkAvgPlan(res map[string][]types.CPUMap, minCon int, maxCon int, name string) error {
	var minC int
	var maxC int
	var temp int
	for _, v := range res {
		temp = len(v)
		if minC > temp || minC == 0 {
			minC = temp
		}
		if maxC < temp {
			maxC = temp
		}
	}
	if minC != minCon || maxC != maxCon {
		fmt.Println(name)
		fmt.Println("min: ", minC)
		fmt.Println("max: ", maxC)
		for k, v := range res {
			fmt.Println(k, ":", len(v))
		}
		return fmt.Errorf("alloc plan error")
	}
	return nil
}

func TestRecurrence(t *testing.T) {
	// 利用相同线上数据复现线上出现的问题

	coreCfg := types.Config{
		EtcdMachines:   []string{"http://127.0.0.1:2379"},
		EtcdLockPrefix: "/eru-core/_lock",
		Scheduler: types.SchedConfig{
			LockKey: "/coretest",
			LockTTL: 1,
			Type:    "complex",
		},
	}

	k, _ := New(coreCfg)

	nodes := map[string]types.CPUMap{
		"c2-docker-26": types.CPUMap{
			"0": 0, "10": 0, "7": 0, "8": 10, "9": 10, "13": 0, "14": 0, "15": 10, "2": 10, "5": 10, "11": 0, "12": 0, "4": 0, "1": 0, "3": 10, "6": 0,
		},
		"c2-docker-28": types.CPUMap{
			"13": 0, "14": 0, "15": 0, "4": 10, "9": 0, "1": 0, "10": 0, "12": 10, "5": 10, "6": 10, "8": 10, "0": 0, "11": 0, "2": 10, "3": 0, "7": 0,
		},
		"c2-docker-27": types.CPUMap{
			"6": 10, "10": 0, "13": 0, "14": 10, "2": 0, "7": 0, "1": 0, "11": 0, "15": 0, "8": 10, "0": 0, "3": 0, "4": 0, "5": 0, "9": 10, "12": 0,
		},
		"c2-docker-29": types.CPUMap{
			"15": 0, "3": 10, "0": 0, "10": 0, "13": 0, "7": 10, "8": 0, "9": 10, "12": 10, "2": 10, "4": 10, "1": 0, "11": 0, "14": 10, "5": 10, "6": 10,
		},
	}
	result, _, err := k.SelectCPUNodes(nodes, 0.5, 3)
	log.Infof("result: %v", result)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}
func TestComplexNodes(t *testing.T) {

	coreCfg := types.Config{
		EtcdMachines:   []string{"http://127.0.0.1:2379"},
		EtcdLockPrefix: "/eru-core/_lock",
		Scheduler: types.SchedConfig{
			LockKey: "/coretest",
			LockTTL: 1,
			Type:    "complex",
		},
	}

	k, merr := New(coreCfg)
	if merr != nil {
		t.Fatalf("Create Potassim error: %v", merr)
	}

	// nodes can offer 28 containers.
	nodes := map[string]types.CPUMap{
		"n1": types.CPUMap{ // 2 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
		},
		"n2": types.CPUMap{ // 7 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
			"12": 10, "13": 10,
		},
		"n3": types.CPUMap{ // 6 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
		},
		"n4": types.CPUMap{ // 9 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
			"12": 10, "13": 10, "14": 10, "15": 10,
			"16": 10, "17": 10,
		},
		"n5": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
	}

	k, _ = New(coreCfg)
	// test1
	res1, changed1, err := k.SelectCPUNodes(nodes, 1.7, 7)
	if err != nil {
		t.Fatalf("sth wrong")
	}
	if check := checkAvgPlan(res1, 1, 2, "res1"); check != nil {
		t.Fatalf("something went wrong")
	}
	assert.Equal(t, len(changed1), len(res1))

	// test2
	// SelectCPUNodes 里有一些副作用, 粗暴地拿一个新的来测试吧
	// 下面也是因为这个
	nodes = map[string]types.CPUMap{
		"n1": types.CPUMap{ // 2 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
		},
		"n2": types.CPUMap{ // 7 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
			"12": 10, "13": 10,
		},
		"n3": types.CPUMap{ // 6 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
		},
		"n4": types.CPUMap{ // 9 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
			"12": 10, "13": 10, "14": 10, "15": 10,
			"16": 10, "17": 10,
		},
		"n5": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
	}
	res2, changed2, err := k.SelectCPUNodes(nodes, 1.7, 11)
	if err != nil {
		t.Fatalf("something went wrong")
	}
	if check := checkAvgPlan(res2, 2, 3, "res2"); check != nil {
		t.Fatalf("something went wrong")
	}
	assert.Equal(t, len(changed2), len(res2))

	// test3
	nodes = map[string]types.CPUMap{
		"n1": types.CPUMap{ // 2 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
		},
		"n2": types.CPUMap{ // 7 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
			"12": 10, "13": 10,
		},
		"n3": types.CPUMap{ // 6 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
		},
		"n4": types.CPUMap{ // 9 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
			"12": 10, "13": 10, "14": 10, "15": 10,
			"16": 10, "17": 10,
		},
		"n5": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
	}
	res3, changed3, err := k.SelectCPUNodes(nodes, 1.7, 23)
	if err != nil {
		fmt.Println("May be we dont have plan")
		fmt.Println(err)
	}
	if check := checkAvgPlan(res3, 2, 6, "res3"); check != nil {
		t.Fatalf("something went wrong")
	}
	assert.Equal(t, len(changed3), len(res3))

	// test4
	nodes = map[string]types.CPUMap{
		"n1": types.CPUMap{ // 2 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
		},
		"n2": types.CPUMap{ // 7 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
			"12": 10, "13": 10,
		},
		"n3": types.CPUMap{ // 6 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
		},
		"n4": types.CPUMap{ // 9 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
			"12": 10, "13": 10, "14": 10, "15": 10,
			"16": 10, "17": 10,
		},
		"n5": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
	}
	_, _, newErr := k.SelectCPUNodes(nodes, 1.6, 29)
	if newErr == nil {
		t.Fatalf("how to alloc 29 containers when you only have 28?")
	}
}

func TestEvenPlan(t *testing.T) {
	coreCfg := types.Config{
		EtcdMachines:   []string{"http://127.0.0.1:2379"},
		EtcdLockPrefix: "/eru-core/_lock",
		Scheduler: types.SchedConfig{
			LockKey: "/coretest",
			LockTTL: 1,
			Type:    "complex",
		},
	}

	k, merr := New(coreCfg)
	if merr != nil {
		t.Fatalf("Create Potassim error: %v", merr)
	}

	// nodes -- n1: 2, n2: 2
	pod1 := map[string]types.CPUMap{
		"n1": types.CPUMap{ // 2 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
		},
		"n2": types.CPUMap{ // 2 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
		},
	}

	res1, rem1, err := k.SelectCPUNodes(pod1, 1.3, 2)
	if err != nil {
		t.Fatalf("sth wrong")
	}
	if check := checkAvgPlan(res1, 1, 1, "res1"); check != nil {
		t.Fatalf("something went wrong")
	}
	assert.Equal(t, len(rem1), 2)

	// nodes -- n1: 4, n2: 5, n3:6, n4: 5
	pod2 := map[string]types.CPUMap{
		"n1": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
		"n2": types.CPUMap{ // 5 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10,
		},
		"n3": types.CPUMap{ // 6 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
		},
		"n4": types.CPUMap{ // 5 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10,
		},
	}

	res2, rem2, err := k.SelectCPUNodes(pod2, 1.7, 3)
	if check := checkAvgPlan(res2, 1, 1, "res2"); check != nil {
		t.Fatalf("something went wront")
	}
	assert.Equal(t, len(rem2), 3)

	pod3 := map[string]types.CPUMap{
		"n1": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
		"n2": types.CPUMap{ // 5 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10,
		},
		"n3": types.CPUMap{ // 6 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
		},
		"n4": types.CPUMap{ // 5 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10,
		},
	}
	res3, rem3, err := k.SelectCPUNodes(pod3, 1.7, 8)
	if check := checkAvgPlan(res3, 2, 2, "res3"); check != nil {
		t.Fatalf("something went wront")
	}
	assert.Equal(t, len(rem3), 4)

	pod4 := map[string]types.CPUMap{
		"n1": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
		"n2": types.CPUMap{ // 5 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10,
		},
		"n3": types.CPUMap{ // 6 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10, "10": 10, "11": 10,
		},
		"n4": types.CPUMap{ // 5 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
			"8": 10, "9": 10,
		},
	}

	res4, rem4, err := k.SelectCPUNodes(pod4, 1.7, 10)
	if check := checkAvgPlan(res4, 2, 3, "res4"); check != nil {
		t.Fatalf("something went wrong")
	}
	assert.Equal(t, len(rem4), 4)
}

func TestSpecialCase(t *testing.T) {
	pod := map[string]types.CPUMap{
		"n1": types.CPUMap{ // 1 containers
			"0": 10, "1": 10,
		},
		"n2": types.CPUMap{ // 3 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10,
		},
		"n3": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
	}

	coreCfg := types.Config{
		EtcdMachines:   []string{"http://127.0.0.1:2379"},
		EtcdLockPrefix: "/eru-core/_lock",
		Scheduler: types.SchedConfig{
			LockKey: "/coretest",
			LockTTL: 1,
			Type:    "complex",
		},
	}

	k, _ := New(coreCfg)
	res1, _, err := k.SelectCPUNodes(pod, 1.7, 7)
	if err != nil {
		t.Fatalf("something went wrong")
	}
	checkAvgPlan(res1, 1, 3, "new test 2")

	newpod := map[string]types.CPUMap{
		"n1": types.CPUMap{ // 3 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10,
		},
		"n2": types.CPUMap{ // 4 containers
			"0": 10, "1": 10, "2": 10, "3": 10,
			"4": 10, "5": 10, "6": 10, "7": 10,
		},
	}

	res2, changed2, _ := k.SelectCPUNodes(newpod, 1.7, 4)
	assert.Equal(t, len(res2), len(changed2))
	checkAvgPlan(res2, 2, 2, "new test 2")
}

func generateNodes(nums, maxCores, seed int) *map[string]types.CPUMap {
	var name string
	var cores int
	pod := make(map[string]types.CPUMap)

	s := rand.NewSource(int64(seed))
	r1 := rand.New(s)

	for i := 0; i < nums; i++ {
		name = fmt.Sprintf("n%d", i)
		cores = r1.Intn(maxCores + 1)
		pod[name] = make(types.CPUMap)
		for j := 0; j < cores; j++ {
			coreName := fmt.Sprintf("%d", j)
			pod[name][coreName] = 10
		}
	}
	return &pod
}

var hugePod = generateNodes(10000, 24, 10086)

func getPodVol(nodes map[string]types.CPUMap, cpu float64) int {
	var res int
	var host *host
	var plan []types.CPUMap

	for _, cpuInfo := range nodes {
		host = newHost(cpuInfo, 10)
		plan = host.getContainerCores(cpu, -1)
		res += len(plan)
	}
	return res
}

func TestGetPodVol(t *testing.T) {
	nodes := map[string]types.CPUMap{
		"c2-docker-26": types.CPUMap{
			"15": 0, "3": 10, "0": 0, "10": 0, "13": 0, "7": 10, "8": 0, "9": 10, "12": 10, "2": 10, "4": 10, "1": 0, "11": 0, "14": 10, "5": 10, "6": 10,
		},
	}

	res := getPodVol(nodes, 0.5)
	assert.Equal(t, res, 18)
	res = getPodVol(nodes, 0.3)
	assert.Equal(t, res, 27)
	res = getPodVol(nodes, 1.1)
	assert.Equal(t, res, 8)
}

func Benchmark_ExtreamAlloc(b *testing.B) {
	coreCfg := types.Config{
		EtcdMachines:   []string{"http://127.0.0.1:2379"},
		EtcdLockPrefix: "/eru-core/_lock",
		Scheduler: types.SchedConfig{
			LockKey: "/coretest",
			LockTTL: 1,
			Type:    "complex",
		},
	}

	k, _ := New(coreCfg)
	b.StopTimer()
	b.StartTimer()

	vol := getPodVol(*hugePod, 1.3)
	result, changed, err := k.SelectCPUNodes(*hugePod, 1.3, vol)
	if err != nil {
		b.Fatalf("something went wrong")
	}
	assert.Equal(b, len(result), len(changed))
}

func Benchmark_AveAlloc(b *testing.B) {
	b.StopTimer()
	coreCfg := types.Config{
		EtcdMachines:   []string{"http://127.0.0.1:2379"},
		EtcdLockPrefix: "/eru-core/_lock",
		Scheduler: types.SchedConfig{
			LockKey: "/coretest",
			LockTTL: 1,
			Type:    "complex",
		},
	}

	k, _ := New(coreCfg)

	b.StartTimer()
	result, changed, err := k.SelectCPUNodes(*hugePod, 1.7, 12000)
	if err != nil {
		b.Fatalf("something went wrong")
	}
	assert.Equal(b, len(result), len(changed))
}
