package kafka

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"

	inputDriver "github.com/brokercap/Bifrost/input/driver"
)

func (c *InputKafka) ConsumePluginPosition(sess sarama.ConsumerGroupSession, ctx context.Context) {
	log.Println("ConsumePluginPosition starting")
	var lastEventId uint64 = 0
	for {
		select {
		case p := <-c.waitCommitOffset:
			if p == nil {
				return
			}
			// 由上一层定时将最小的位点提交回input 插件层，所以一直没有数据，一直重新重复提交相同的位点进来
			// 所以这里需要判断一下只要和上一次eventId不一样，则需要保存
			// 这里为什么不判断 > lastEventId ，是因为存在可能EventID被更新了，强制变小了的可能
			if p.EventID == lastEventId {
				break
			}
			data := c.TransferWaitCommitOffsetList(p)
			for _, pluginPosition := range data {
				sess.MarkOffset(pluginPosition.topic, pluginPosition.partition, pluginPosition.offset+1, "")
				sess.Commit()
			}
			lastEventId = p.EventID
			break
		case <-ctx.Done():
			return
		}
	}
}

func (c *InputKafka) TransferWaitCommitOffsetList(p *inputDriver.PluginPosition) (data []*waitCommitOffset) {
	if p == nil {
		return
	}
	if p.GTID == "" {
		return
	}
	for _, gtid := range strings.Split(p.GTID, ",") {
		gtidInfoArr := strings.Split(gtid, ":")
		if len(gtidInfoArr) != 3 {
			continue
		}
		partition, err := strconv.ParseInt(gtidInfoArr[1], 10, 32)
		if err != nil {
			continue
		}
		offset, err := strconv.ParseInt(gtidInfoArr[2], 10, 64)
		if err != nil {
			continue
		}
		data = append(data, &waitCommitOffset{
			topic:     gtidInfoArr[0],
			partition: int32(partition),
			offset:    offset,
		})
	}
	return
}

func (c *InputKafka) DoneMinPosition(p *inputDriver.PluginPosition) (err error) {
	if p == nil {
		return
	}
	// 这里加一个超时，防止 waitCommitOffset 被阻塞，导致上一层被阻塞
	timer := time.NewTimer(15 * time.Second)
	defer timer.Stop()
	select {
	case c.waitCommitOffset <- p:
		break
	case <-timer.C:
		break
	}
	return nil
}
