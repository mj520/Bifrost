/*
Copyright [2018] [jc3wish]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kafka

import (
	"context"
	"github.com/Shopify/sarama"
)

func (c *InputKafka) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *InputKafka) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *InputKafka) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ConsumeClaimCtx, ConsumeClaimCancle := context.WithCancel(c.kafkaGroupCtx)
	defer ConsumeClaimCancle()
	go c.ConsumePluginPosition(sess, ConsumeClaimCtx)
	for {
		select {
		case kafkaMsg := <-claim.Messages():
			if kafkaMsg == nil {
				return nil
			}
			c.ToChildCallback(kafkaMsg)
			break
		case _ = <-c.kafkaGroupCtx.Done():
			return nil
		}
	}
	return nil
}
