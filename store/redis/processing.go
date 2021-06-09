package redis

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sanity-io/litter"

	"github.com/projecteru2/core/log"
	"github.com/projecteru2/core/strategy"
	"github.com/projecteru2/core/types"
)

// SaveProcessing save processing status in etcd
func (r *Rediaron) CreateProcessing(ctx context.Context, processing *types.Processing, count int) error {
	processingKey := filepath.Join(workloadProcessingPrefix, processing.BaseKey())
	return r.BatchCreate(ctx, map[string]string{processingKey: strconv.Itoa(count)})
}

func (r *Rediaron) DecrProcessing(ctx context.Context, processing *types.Processing) (err error) {
	_, err = r.cli.Decr(ctx, filepath.Join(workloadProcessingPrefix, processing.BaseKey())).Result()
	return
}

// DeleteProcessing delete processing status in etcd
func (r *Rediaron) DeleteProcessing(ctx context.Context, processing *types.Processing) error {
	processingKey := filepath.Join(workloadProcessingPrefix, processing.BaseKey())
	return r.BatchDelete(ctx, []string{processingKey})
}

func (r *Rediaron) doLoadProcessing(ctx context.Context, processing *types.Processing, strategyInfos []strategy.Info) error {
	// 显式的加 / 保证 prefix 一致性
	processingKey := filepath.Join(workloadProcessingPrefix, processing.Appname, processing.Entryname) + "/*"
	data, err := r.getByKeyPattern(ctx, processingKey, 0)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}

	nodesCount := map[string]int{}
	for k, v := range data {
		parts := strings.Split(k, "/")
		nodename := parts[len(parts)-2]
		count, err := strconv.Atoi(v)
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

	log.Debug(ctx, "[doLoadProcessing] Processing result:")
	litter.Dump(nodesCount)
	setCount(nodesCount, strategyInfos)
	return nil
}
