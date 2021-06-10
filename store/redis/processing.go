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

func (r *Rediaron) getProcessingKey(processing *types.Processing) string {
	return filepath.Join(workloadProcessingPrefix, processing.Appname, processing.Entryname, processing.Nodename, processing.Ident)
}

// SaveProcessing save processing status in etcd
func (r *Rediaron) CreateProcessing(ctx context.Context, processing *types.Processing, count int) error {
	processingKey := r.getProcessingKey(processing)
	return r.BatchCreate(ctx, map[string]string{processingKey: strconv.Itoa(count)})
}

// DeleteProcessing delete processing status in etcd
func (r *Rediaron) DeleteProcessing(ctx context.Context, processing *types.Processing) error {
	return r.BatchDelete(ctx, []string{r.getProcessingKey(processing)})
}

func (r *Rediaron) doLoadProcessing(ctx context.Context, appname, entryname string, strategyInfos []strategy.Info) error {
	// 显式的加 / 保证 prefix 一致性
	processingKey := filepath.Join(workloadProcessingPrefix, appname, entryname) + "/*"
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
