package worker

import (
	"context"
	"time"

	"go-crontab/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var GLogManager *LogManager

type LogManager struct {
	client         *mongo.Client
	collection     *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.BatchJobLog
}

func InitLogManager() error {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://106.54.212.69:27017"))
	if err != nil {
		return err
	}
	// 获取collection对象
	collection := client.Database("job").Collection("logs")

	GLogManager = &LogManager{
		client:         client,
		collection:     collection,
		logChan:        make(chan *common.JobLog, 100),
		autoCommitChan: make(chan *common.BatchJobLog, 100),
	}

	go GLogManager.WriteLog()

	return nil
}

// 监听log队列，将任务保存到MongoDB中
func (l *LogManager) WriteLog() {
	var (
		batchLogs    *common.BatchJobLog
		timeOutTimer *time.Timer
	)

	for {
		select {
		case log := <-l.logChan:
			if batchLogs == nil {
				batchLogs = &common.BatchJobLog{}

				// 因为 time.AfterFunc会启动一个协程执行，为了不改变batchLogs的指针
				// 我们将batchLogs通过参数传递进去，然后返回一个 func
				timeOutTimer = time.AfterFunc(3*time.Second, func(batch *common.BatchJobLog) func() {
					return func() {
						l.autoCommitChan <- batchLogs
					}
				}(batchLogs))
			}
			batchLogs.Logs = append(batchLogs.Logs, log)

			if len(batchLogs.Logs) >= 100 {
				// 发送给MongoDB进行保存
				l.SaveLog(batchLogs)
				batchLogs = nil
				timeOutTimer.Stop()
			}
		case autoCommit := <-l.autoCommitChan:
			if autoCommit != batchLogs {
				continue
			}
			l.SaveLog(autoCommit)
		}
	}
}

// 批量插入到MongoDB中
func (l *LogManager) SaveLog(logs *common.BatchJobLog) {
	l.collection.InsertMany(context.Background(), logs.Logs)
}

// 将lag放入到队列里
func (l *LogManager) AddLog(log *common.JobLog) {
	select {
	case l.logChan <- log:
	default:
		// 这里如果队列满了，就会丢弃这个log
	}
}
