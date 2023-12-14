package logic

import (
	"context"

	"github.com/scroll-tech/go-ethereum/log"
	"gorm.io/gorm"

	"scroll-tech/bridge-history-api/internal/orm"
)

// EventUpdateLogic the logic of insert/update the database
type EventUpdateLogic struct {
	db              *gorm.DB
	crossMessageOrm *orm.CrossMessage
	batchEventOrm   *orm.BatchEvent
}

// NewEventUpdateLogic create a EventUpdateLogic instance
func NewEventUpdateLogic(db *gorm.DB) *EventUpdateLogic {
	return &EventUpdateLogic{
		db:              db,
		crossMessageOrm: orm.NewCrossMessage(db),
		batchEventOrm:   orm.NewBatchEvent(db),
	}
}

// GetL1SyncHeight get the l1 sync height from db
func (b *EventUpdateLogic) GetL1SyncHeight(ctx context.Context) (uint64, uint64, error) {
	messageSyncedHeight, err := b.crossMessageOrm.GetMessageSyncedHeightInDB(ctx, orm.MessageTypeL1SentMessage)
	if err != nil {
		log.Error("failed to get L1 cross message synced height", "error", err)
		return 0, 0, err
	}

	batchSyncedHeight, err := b.batchEventOrm.GetBatchEventSyncedHeightInDB(ctx)
	if err != nil {
		log.Error("failed to get L1 batch event synced height", "error", err)
		return 0, 0, err
	}

	return messageSyncedHeight, batchSyncedHeight, nil
}

// GetL2MessageSyncedHeightInDB get l2 messages synced height
func (b *EventUpdateLogic) GetL2MessageSyncedHeightInDB(ctx context.Context) (uint64, error) {
	l2SentMessageSyncedHeight, err := b.crossMessageOrm.GetMessageSyncedHeightInDB(ctx, orm.MessageTypeL2SentMessage)
	if err != nil {
		log.Error("failed to get L2 cross message processed height", "err", err)
		return 0, err
	}
	return l2SentMessageSyncedHeight, nil
}

// GetL2LatestWithdrawal get l2 latest withdrawal message
func (b *EventUpdateLogic) GetL2LatestWithdrawal(ctx context.Context) (*orm.CrossMessage, error) {
	message, err := b.crossMessageOrm.GetLatestL2Withdrawal(ctx)
	if err != nil {
		log.Error("failed to get latest L2 sent message event", "err", err)
		return nil, err
	}
	return message, nil
}

// L1InsertOrUpdate insert or update l1 messages
func (b *EventUpdateLogic) L1InsertOrUpdate(ctx context.Context, l1FetcherResult *L1FilterResult) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		if txErr := b.crossMessageOrm.InsertOrUpdateL1Messages(ctx, l1FetcherResult.DepositMessages, tx); txErr != nil {
			log.Error("failed to insert L1 deposit messages", "err", txErr)
			return txErr
		}

		if txErr := b.crossMessageOrm.InsertOrUpdateL1RelayedMessagesOfL2Withdrawals(ctx, l1FetcherResult.RelayedMessages, tx); txErr != nil {
			log.Error("failed to update L1 relayed messages of L2 withdrawals", "err", txErr)
			return txErr
		}

		if txErr := b.batchEventOrm.InsertOrUpdateBatchEvents(ctx, l1FetcherResult.BatchEvents, tx); txErr != nil {
			log.Error("failed to insert or update batch events", "err", txErr)
			return txErr
		}

		if txErr := b.crossMessageOrm.UpdateL1MessageQueueEventsInfo(ctx, l1FetcherResult.MessageQueueEvents, tx); txErr != nil {
			log.Error("failed to insert L1 message queue events", "err", txErr)
			return txErr
		}

		if txErr := b.crossMessageOrm.InsertFailedGatewayRouterTxs(ctx, l1FetcherResult.FailedGatewayRouterTxs, tx); txErr != nil {
			log.Error("failed to insert L1 failed gateway router transactions", "err", txErr)
			return txErr
		}
		return nil
	})

	if err != nil {
		log.Error("failed to update db of L1 events", "err", err)
		return err
	}

	return nil
}

// UpdateL1BatchIndexAndStatus update l1 batch index and status
func (b *EventUpdateLogic) UpdateL1BatchIndexAndStatus(ctx context.Context, height uint64) error {
	batches, err := b.batchEventOrm.GetBatchesLEBlockHeight(ctx, height)
	if err != nil {
		log.Error("failed to get batches >= block height", "error", err)
		return err
	}

	for _, batch := range batches {
		log.Info("update batch info of L2 withdrawals", "index", batch.BatchIndex, "start", batch.StartBlockNumber, "end", batch.EndBlockNumber)
		if err = b.crossMessageOrm.UpdateBatchStatusOfL2Withdrawals(ctx, batch.StartBlockNumber, batch.EndBlockNumber, batch.BatchIndex); err != nil {
			log.Error("failed to update batch status of L2 sent messages", "start", batch.StartBlockNumber, "end", batch.EndBlockNumber, "index", batch.BatchIndex, "error", err)
			return err
		}
		if err = b.batchEventOrm.UpdateBatchEventStatus(ctx, batch.BatchIndex); err != nil {
			log.Error("failed to update batch event status as updated", "start", batch.StartBlockNumber, "end", batch.EndBlockNumber, "index", batch.BatchIndex, "error", err)
			return err
		}
	}

	return nil
}

// L2InsertOrUpdate insert or update l2 messages
func (b *EventUpdateLogic) L2InsertOrUpdate(ctx context.Context, l2FetcherResult *L2FilterResult) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		if txErr := b.crossMessageOrm.InsertOrUpdateL2Messages(ctx, l2FetcherResult.WithdrawMessages, tx); txErr != nil {
			log.Error("failed to insert L2 withdrawal messages", "err", txErr)
			return txErr
		}
		if txErr := b.crossMessageOrm.InsertOrUpdateL2RelayedMessagesOfL1Deposits(ctx, l2FetcherResult.RelayedMessages, tx); txErr != nil {
			log.Error("failed to update L2 relayed messages of L1 deposits", "err", txErr)
			return txErr
		}
		if txErr := b.crossMessageOrm.InsertOrUpdateL2RevertedRelayedMessagesOfL1Deposits(ctx, l2FetcherResult.RevertedRelayedMessages, tx); txErr != nil {
			log.Error("failed to update L2 relayed messages of L1 deposits", "err", txErr)
			return txErr
		}
		if txErr := b.crossMessageOrm.InsertFailedGatewayRouterTxs(ctx, l2FetcherResult.FailedGatewayRouterTxs, tx); txErr != nil {
			log.Error("failed to insert L2 failed gateway router transactions", "err", txErr)
			return txErr
		}
		return nil
	})

	if err != nil {
		log.Error("failed to update db of L2 events", "err", err)
		return err
	}
	return nil
}