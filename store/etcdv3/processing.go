package etcdv3

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/projecteru2/core/log"
	"github.com/projecteru2/core/strategy"
	"github.com/projecteru2/core/types"
	"go.etcd.io/etcd/v3/clientv3"
)

// CreateProcessing save processing status in etcd
func (m *Mercury) CreateProcessing(ctx context.Context, processing *types.Processing, count int) error {
	_, err := m.Create(ctx, filepath.Join(workloadProcessingPrefix, processing.BaseKey()), fmt.Sprintf("%d", count))
	return err
}

func (m *Mercury) DecrProcessing(ctx context.Context, processing *types.Processing) error {
	return m.Decr(ctx, filepath.Join(workloadProcessingPrefix, processing.BaseKey()))
}

// DeleteProcessing delete processing status in etcd
func (m *Mercury) DeleteProcessing(ctx context.Context, processing *types.Processing) error {
	_, err := m.Delete(ctx, filepath.Join(workloadProcessingPrefix, processing.BaseKey()))
	return err
}

func (m *Mercury) doLoadProcessing(ctx context.Context, processing *types.Processing, strategyInfos []strategy.Info) error {
	// 显式的加 / 保证 prefix 一致性
	processingKey := filepath.Join(workloadProcessingPrefix, processing.Appname, processing.Entryname) + "/"
	resp, err := m.Get(ctx, processingKey, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	if resp.Count == 0 {
		return nil
	}
	nodesCount := map[string]int{}
	for _, ev := range resp.Kvs {
		key := string(ev.Key)
		parts := strings.Split(key, "/")
		nodename := parts[len(parts)-2]
		count, err := strconv.Atoi(string(ev.Value))
		if err != nil {
			log.Errorf(ctx, "[doLoadProcessing] Load processing status failed %v", err)
			continue
		}
		if _, ok := nodesCount[nodename]; !ok {
			nodesCount[nodename] = count
			continue
		}
		nodesCount[nodename] += count
	}

	log.Debug(ctx, "[doLoadProcessing] Processing result: %+v", nodesCount)
	setCount(nodesCount, strategyInfos)
	return nil
}
