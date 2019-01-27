package consumergroup

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

type partitionConsumer struct {
	owner      *topicConsumer
	group      string
	topic      string
	partition  int32
	offset     int64
	prevOffset int64

	consumer sarama.PartitionConsumer
}

func newPartitionConsumer(owner *topicConsumer, partition int32) *partitionConsumer {
	return &partitionConsumer{
		owner:      owner,
		topic:      owner.name,
		group:      owner.owner.name,
		partition:  partition,
		offset:     0,
		prevOffset: 0,
	}
}

func (pc *partitionConsumer) start() {
	var wg sync.WaitGroup

	cg := pc.owner.owner
	err := pc.claim()
	if err != nil {
		cg.logger.WithFields(logrus.Fields{
			"group":     pc.group,
			"topic":     pc.topic,
			"partition": pc.partition,
			"err":       err,
		}).Error("Failed to claim the partition and gave up")
		goto ERROR
	}
	defer func() {
		err = pc.release()
		if err != nil {
			cg.logger.WithFields(logrus.Fields{
				"group":     pc.group,
				"topic":     pc.topic,
				"partition": pc.partition,
				"err":       err,
			}).Error("Failed to release the partition")
		} else {
			cg.logger.WithFields(logrus.Fields{
				"group":     pc.group,
				"topic":     pc.topic,
				"partition": pc.partition,
			}).Info("Success to release the partition")
		}
	}()

	err = pc.loadOffsetFromZk()
	if err != nil {
		cg.logger.WithFields(logrus.Fields{
			"group":     pc.group,
			"topic":     pc.topic,
			"partition": pc.partition,
			"err":       err,
		}).Error("Failed to fetch the partition's offset")
		goto ERROR
	}
	cg.logger.WithFields(logrus.Fields{
		"group":     pc.group,
		"topic":     pc.topic,
		"partition": pc.partition,
		"offset":    pc.offset,
	}).Info("Fetched the partition's offset from zk")
	pc.consumer, err = cg.getPartitionConsumer(pc.topic, pc.partition, pc.offset)
	if err != nil {
		cg.logger.WithFields(logrus.Fields{
			"group":     pc.group,
			"topic":     pc.topic,
			"partition": pc.partition,
			"offset":    pc.offset,
			"err":       err,
		}).Error("Failed to create the partition's consumer")
		goto ERROR
	}
	defer pc.consumer.Close()

	if cg.config.OffsetAutoCommitEnable { // start auto commit-offset thread when enable
		wg.Add(1)
		go func() {
			defer cg.callRecover()
			defer wg.Done()
			cg.logger.WithFields(logrus.Fields{
				"group":     pc.group,
				"topic":     pc.topic,
				"partition": pc.partition,
			}).Info("Start the partition's offset auto-commit thread")
			pc.autoCommitOffset()
		}()
	}

	pc.fetch()
	if cg.config.OffsetAutoCommitEnable {
		err = pc.commitOffset()
		if err != nil {
			cg.logger.WithFields(logrus.Fields{
				"group":     pc.group,
				"topic":     pc.topic,
				"partition": pc.partition,
				"offset":    pc.offset,
				"err":       err,
			}).Error("Failed to commit the partition's offset")
		}

		wg.Wait() // Wait for auto-commit-offset thread
		cg.logger.WithFields(logrus.Fields{
			"group":     pc.group,
			"topic":     pc.topic,
			"partition": pc.partition,
			"offset":    pc.offset,
		}).Info("Start the partition's offset auto-commit thread")
	}
	return

ERROR:
	cg.stop()
}

func (pc *partitionConsumer) loadOffsetFromZk() error {
	cg := pc.owner.owner
	offset, err := cg.storage.getOffset(pc.group, pc.topic, pc.partition)
	if err != nil {
		return err
	}
	if offset == -1 {
		offset = cg.config.OffsetAutoReset
	}
	pc.offset = offset
	pc.prevOffset = offset
	return nil
}

func (pc *partitionConsumer) claim() error {
	cg := pc.owner.owner
	timer := time.NewTimer(cg.config.ClaimPartitionRetryInterval)
	defer timer.Stop()
	retry := cg.config.ClaimPartitionRetryTimes
	// Claim partition would retry until success
	for i := 0; i < retry+1 || retry <= 0; i++ {
		err := cg.storage.claimPartition(pc.group, pc.topic, pc.partition, cg.id)
		if err == nil {
			return nil
		}
		if i != 0 && (i%3 == 0 || retry > 0) {
			cg.logger.WithFields(logrus.Fields{
				"group":     pc.group,
				"topic":     pc.topic,
				"partition": pc.partition,
				"retries":   i,
				"err":       err,
			}).Warn("Failed to claim the partition with retries")
		}
		select {
		case <-timer.C:
			timer.Reset(cg.config.ClaimPartitionRetryInterval)
		case <-cg.stopCh:
			return errors.New("stop signal was received when claim partition")
		}
	}
	return fmt.Errorf("failed to claim partition after %d retries", retry)
}

func (pc *partitionConsumer) release() error {
	cg := pc.owner.owner
	owner, err := cg.storage.getPartitionOwner(pc.group, pc.topic, pc.partition)
	if err != nil {
		return err
	}
	if cg.id == owner {
		return cg.storage.releasePartition(pc.group, pc.topic, pc.partition)
	}
	return fmt.Errorf("the owner of topic[%s] partition[%d] expected %s, but got %s",
		pc.topic, pc.partition, owner, cg.id)
}

func (pc *partitionConsumer) fetch() {
	cg := pc.owner.owner
	messageChan := pc.owner.messages
	errorChan := pc.owner.errors

	cg.logger.WithFields(logrus.Fields{
		"group":     pc.group,
		"topic":     pc.topic,
		"partition": pc.partition,
		"offset":    pc.offset,
	}).Info("Start to fetch the partition's messages")
PARTITION_CONSUMER_LOOP:
	for {
		select {
		case <-cg.stopCh:
			break PARTITION_CONSUMER_LOOP
		case err := <-pc.consumer.Errors():
			if err.Err == sarama.ErrOffsetOutOfRange {
				pc.restart()
				break
			}
			errorChan <- err
		case message, ok := <-pc.consumer.Messages():
			//check if the channel is closed. message channel close while the offset out of range
			if !ok {
				pc.restart()
				break
			}
			if message == nil {
				cg.logger.WithFields(logrus.Fields{
					"group":     pc.group,
					"topic":     pc.topic,
					"partition": pc.partition,
					"offset":    pc.offset,
				}).Error("Sarama partition consumer encounter error, the consumer would be exited")
				cg.stop()
				break PARTITION_CONSUMER_LOOP
			}
			select {
			case messageChan <- message:
				pc.offset = message.Offset + 1
			case <-cg.stopCh:
				break PARTITION_CONSUMER_LOOP
			}
		}
	}
}

func (pc *partitionConsumer) autoCommitOffset() {
	cg := pc.owner.owner
	defer cg.callRecover()
	timer := time.NewTimer(cg.config.OffsetAutoCommitInterval)
	for {
		select {
		case <-cg.stopCh:
			return
		case <-timer.C:
			err := pc.commitOffset()
			if err != nil {
				cg.logger.WithFields(logrus.Fields{
					"topic":     pc.topic,
					"partition": pc.partition,
					"offset":    pc.offset,
					"err":       err,
				}).Error("Failed to auto-commit the partition's offset")
			}
			timer.Reset(cg.config.OffsetAutoCommitInterval)
		}
	}
}

func (pc *partitionConsumer) commitOffset() error {
	cg := pc.owner.owner
	offset := pc.offset
	if pc.prevOffset == offset {
		return nil
	}
	err := cg.storage.commitOffset(pc.group, pc.topic, pc.partition, offset)
	if err != nil {
		return err
	}
	pc.prevOffset = offset
	return nil
}

func (pc *partitionConsumer) restart() {
	cg := pc.owner.owner
	cg.logger.WithFields(logrus.Fields{
		"group":     pc.group,
		"topic":     pc.topic,
		"partition": pc.partition,
	}).Infof("Restart partition consumer while the offset out of range")
	err := pc.consumer.Close()
	if err != nil {
		cg.logger.WithFields(logrus.Fields{
			"group":     pc.group,
			"topic":     pc.topic,
			"partition": pc.partition,
		}).Error("Stop consumer group because the old partition consumer cannot be closed")
		cg.stop()
		return
	}
	pc.consumer, err = cg.getPartitionConsumer(pc.topic, pc.partition, sarama.OffsetOldest)
	if err != nil {
		cg.logger.WithFields(logrus.Fields{
			"group":     pc.group,
			"topic":     pc.topic,
			"partition": pc.partition,
			"err":       err,
		}).Error("Stop consumer group because the new partition consumer cannot be start")
		cg.stop()
	}
}

func (pc *partitionConsumer) getOffset() map[string]interface{} {
	offset := make(map[string]interface{})
	offset["offset"] = pc.offset
	offset["prev_offset"] = pc.prevOffset
	return offset
}
