package taskmanager

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/StellarServer/internal/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// QueueManager 队列管理器
type QueueManager struct {
	db          *mongo.Database
	redisClient *redis.Client
	queues      map[string]*TaskQueue
	queuesMutex sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewQueueManager 创建队列管理器
func NewQueueManager(db *mongo.Database, redisClient *redis.Client) *QueueManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &QueueManager{
		db:          db,
		redisClient: redisClient,
		queues:      make(map[string]*TaskQueue),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动队列管理器
func (qm *QueueManager) Start() error {
	// 从数据库加载队列
	if err := qm.loadQueues(); err != nil {
		return err
	}

	// 启动队列处理
	go qm.processQueues()

	return nil
}

// Stop 停止队列管理器
func (qm *QueueManager) Stop() {
	qm.cancel()
}

// CreateQueue 创建队列
func (qm *QueueManager) CreateQueue(name string, queueType string, priority int, maxSize int) (*TaskQueue, error) {
	qm.queuesMutex.Lock()
	defer qm.queuesMutex.Unlock()

	// 检查队列是否已存在
	if _, exists := qm.queues[name]; exists {
		return nil, errors.New("队列已存在")
	}

	// 创建队列
	queue := &TaskQueue{
		Name:      name,
		Type:      queueType,
		Priority:  priority,
		MaxSize:   maxSize,
		Tasks:     make([]*models.Task, 0),
		TaskCount: 0,
		Mutex:     sync.Mutex{},
	}

	// 保存队列到数据库
	queueDoc := models.TaskQueue{
		ID:        primitive.NewObjectID(),
		Name:      name,
		Type:      queueType,
		Priority:  priority,
		MaxSize:   maxSize,
		TaskCount: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := qm.db.Collection("task_queues").InsertOne(qm.ctx, queueDoc)
	if err != nil {
		return nil, err
	}

	// 添加到内存中
	qm.queues[name] = queue

	return queue, nil
}

// GetQueue 获取队列
func (qm *QueueManager) GetQueue(name string) (*TaskQueue, error) {
	qm.queuesMutex.RLock()
	defer qm.queuesMutex.RUnlock()

	queue, exists := qm.queues[name]
	if !exists {
		return nil, errors.New("队列不存在")
	}

	return queue, nil
}

// EnqueueTask 将任务加入队列
func (qm *QueueManager) EnqueueTask(queueName string, task *models.Task) error {
	queue, err := qm.GetQueue(queueName)
	if err != nil {
		return err
	}

	queue.Mutex.Lock()
	defer queue.Mutex.Unlock()

	// 检查队列是否已满
	if queue.MaxSize > 0 && queue.TaskCount >= queue.MaxSize {
		return errors.New("队列已满")
	}

	// 更新任务状态
	task.Status = string(models.TaskStatusQueued)

	// 添加到队列
	queue.Tasks = append(queue.Tasks, task)
	queue.TaskCount++

	// 更新数据库中的队列
	_, err = qm.db.Collection("task_queues").UpdateOne(
		qm.ctx,
		bson.M{"name": queueName},
		bson.M{
			"$inc": bson.M{"taskCount": 1},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		return err
	}

	// 更新数据库中的任务状态
	_, err = qm.db.Collection("tasks").UpdateOne(
		qm.ctx,
		bson.M{"_id": task.ID},
		bson.M{
			"$set": bson.M{
				"status": string(models.TaskStatusQueued),
			},
		},
	)
	if err != nil {
		return err
	}

	// 将任务ID添加到Redis队列
	err = qm.redisClient.LPush(qm.ctx, "task_queue:"+queueName, task.ID.Hex()).Err()
	if err != nil {
		return err
	}

	return nil
}

// DequeueTask 从队列中取出任务
func (qm *QueueManager) DequeueTask(queueName string) (*models.Task, error) {
	queue, err := qm.GetQueue(queueName)
	if err != nil {
		return nil, err
	}

	queue.Mutex.Lock()
	defer queue.Mutex.Unlock()

	// 检查队列是否为空
	if queue.TaskCount == 0 || len(queue.Tasks) == 0 {
		return nil, errors.New("队列为空")
	}

	// 获取第一个任务
	task := queue.Tasks[0]
	queue.Tasks = queue.Tasks[1:]
	queue.TaskCount--

	// 更新数据库中的队列
	_, err = qm.db.Collection("task_queues").UpdateOne(
		qm.ctx,
		bson.M{"name": queueName},
		bson.M{
			"$inc": bson.M{"taskCount": -1},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		return nil, err
	}

	// 从Redis队列中移除任务
	err = qm.redisClient.LRem(qm.ctx, "task_queue:"+queueName, 1, task.ID.Hex()).Err()
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetNextTask 获取下一个要执行的任务
func (qm *QueueManager) GetNextTask() (*models.Task, string, error) {
	qm.queuesMutex.RLock()
	defer qm.queuesMutex.RUnlock()

	// 按优先级排序队列
	var queues []*TaskQueue
	for _, queue := range qm.queues {
		queues = append(queues, queue)
	}

	sort.Slice(queues, func(i, j int) bool {
		return queues[i].Priority > queues[j].Priority
	})

	// 从高优先级队列开始查找任务
	for _, queue := range queues {
		queue.Mutex.Lock()
		if queue.TaskCount > 0 && len(queue.Tasks) > 0 {
			task := queue.Tasks[0]
			queue.Tasks = queue.Tasks[1:]
			queue.TaskCount--
			queue.Mutex.Unlock()

			// 更新数据库中的队列
			_, err := qm.db.Collection("task_queues").UpdateOne(
				qm.ctx,
				bson.M{"name": queue.Name},
				bson.M{
					"$inc": bson.M{"taskCount": -1},
					"$set": bson.M{"updatedAt": time.Now()},
				},
			)
			if err != nil {
				return nil, "", err
			}

			// 从Redis队列中移除任务
			err = qm.redisClient.LRem(qm.ctx, "task_queue:"+queue.Name, 1, task.ID.Hex()).Err()
			if err != nil {
				return nil, "", err
			}

			return task, queue.Name, nil
		}
		queue.Mutex.Unlock()
	}

	return nil, "", errors.New("所有队列都为空")
}

// loadQueues 从数据库加载队列
func (qm *QueueManager) loadQueues() error {
	// 查询所有队列
	cursor, err := qm.db.Collection("task_queues").Find(qm.ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(qm.ctx)

	// 清空当前队列
	qm.queuesMutex.Lock()
	qm.queues = make(map[string]*TaskQueue)
	qm.queuesMutex.Unlock()

	// 加载队列
	for cursor.Next(qm.ctx) {
		var queueDoc models.TaskQueue
		if err := cursor.Decode(&queueDoc); err != nil {
			return err
		}

		// 创建队列
		queue := &TaskQueue{
			Name:      queueDoc.Name,
			Type:      queueDoc.Type,
			Priority:  queueDoc.Priority,
			MaxSize:   queueDoc.MaxSize,
			Tasks:     make([]*models.Task, 0),
			TaskCount: 0,
			Mutex:     sync.Mutex{},
		}

		// 加载队列中的任务
		taskIDs, err := qm.redisClient.LRange(qm.ctx, "task_queue:"+queueDoc.Name, 0, -1).Result()
		if err != nil && err != redis.Nil {
			return err
		}

		if len(taskIDs) > 0 {
			for _, taskID := range taskIDs {
				objID, err := primitive.ObjectIDFromHex(taskID)
				if err != nil {
					continue
				}

				var task models.Task
				err = qm.db.Collection("tasks").FindOne(qm.ctx, bson.M{"_id": objID}).Decode(&task)
				if err != nil {
					continue
				}

				queue.Tasks = append(queue.Tasks, &task)
				queue.TaskCount++
			}
		}

		// 添加到内存中
		qm.queuesMutex.Lock()
		qm.queues[queueDoc.Name] = queue
		qm.queuesMutex.Unlock()
	}

	return nil
}

// processQueues 处理队列
func (qm *QueueManager) processQueues() {
	// 这里可以实现定期检查队列状态、清理过期任务等逻辑
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-qm.ctx.Done():
			return
		case <-ticker.C:
			// 定期检查队列状态
			qm.checkQueues()
		}
	}
}

// checkQueues 检查队列状态
func (qm *QueueManager) checkQueues() {
	qm.queuesMutex.RLock()
	defer qm.queuesMutex.RUnlock()

	for name, queue := range qm.queues {
		// 检查Redis队列与内存队列是否一致
		taskIDs, err := qm.redisClient.LRange(qm.ctx, "task_queue:"+name, 0, -1).Result()
		if err != nil && err != redis.Nil {
			continue
		}

		queue.Mutex.Lock()
		if len(taskIDs) != queue.TaskCount {
			// 如果不一致，重新同步
			queue.Tasks = make([]*models.Task, 0)
			queue.TaskCount = 0

			for _, taskID := range taskIDs {
				objID, err := primitive.ObjectIDFromHex(taskID)
				if err != nil {
					continue
				}

				var task models.Task
				err = qm.db.Collection("tasks").FindOne(qm.ctx, bson.M{"_id": objID}).Decode(&task)
				if err != nil {
					continue
				}

				queue.Tasks = append(queue.Tasks, &task)
				queue.TaskCount++
			}

			// 更新数据库中的队列
			_, _ = qm.db.Collection("task_queues").UpdateOne(
				qm.ctx,
				bson.M{"name": name},
				bson.M{
					"$set": bson.M{
						"taskCount": queue.TaskCount,
						"updatedAt": time.Now(),
					},
				},
			)
		}
		queue.Mutex.Unlock()
	}
}
