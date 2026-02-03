package service

import (
	"context"
	"mbmanager/internal/database"
	"mbmanager/internal/model"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// SchedulerService 调度服务
type SchedulerService struct {
	scheduler   gocron.Scheduler
	backupSvc   *BackupService
	taskJobs    map[uint]gocron.Job // 任务ID -> Job映射
	taskLocks   sync.Map            // 任务锁，防止并发执行
	mu          sync.RWMutex
}

// NewSchedulerService 创建调度服务实例
func NewSchedulerService(backupSvc *BackupService) (*SchedulerService, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &SchedulerService{
		scheduler: scheduler,
		backupSvc: backupSvc,
		taskJobs:  make(map[uint]gocron.Job),
	}, nil
}

// Start 启动调度器
func (s *SchedulerService) Start(ctx context.Context) error {
	log.Println("Starting scheduler service...")

	// 加载所有启用的任务
	var tasks []model.Task
	if err := database.DB.Where("status = ?", 1).Preload("Host").Find(&tasks).Error; err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// 为每个任务创建调度作业
	for _, task := range tasks {
		if err := s.AddTask(&task); err != nil {
			log.Printf("Failed to add task %s: %v", task.Name, err)
		}
	}

	// 启动调度器
	s.scheduler.Start()
	log.Printf("Scheduler started with %d tasks", len(tasks))

	return nil
}

// Stop 停止调度器
func (s *SchedulerService) Stop() error {
	log.Println("Stopping scheduler service...")
	return s.scheduler.Shutdown()
}

// AddTask 添加任务到调度器
func (s *SchedulerService) AddTask(task *model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果任务已存在，先移除
	if existingJob, exists := s.taskJobs[task.ID]; exists {
		if err := s.scheduler.RemoveJob(existingJob.ID()); err != nil {
			log.Printf("Failed to remove existing job: %v", err)
		}
		delete(s.taskJobs, task.ID)
	}

	// 解析调度配置
	var scheduleConfig map[string]interface{}
	if err := json.Unmarshal([]byte(task.ScheduleConfig), &scheduleConfig); err != nil {
		return fmt.Errorf("failed to parse schedule config: %w", err)
	}

	// 创建任务函数
	jobTask := s.createJobTask(task)

	// 根据调度类型创建作业
	var job gocron.Job
	var err error

	switch task.ScheduleType {
	case "once":
		// 一次性任务
		job, err = s.scheduler.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()),
			jobTask,
		)

	case "daily":
		// 每天指定时间
		timeStr, _ := scheduleConfig["time"].(string)
		if timeStr == "" {
			timeStr = "02:00"
		}
		hour, minute := parseTime(timeStr)
		job, err = s.scheduler.NewJob(
			gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(hour, minute, 0))),
			jobTask,
		)

	case "weekly":
		// 每周指定星期和时间
		weekday, _ := scheduleConfig["weekday"].(float64)
		timeStr, _ := scheduleConfig["time"].(string)
		if timeStr == "" {
			timeStr = "02:00"
		}
		hour, minute := parseTime(timeStr)
		job, err = s.scheduler.NewJob(
			gocron.WeeklyJob(1, gocron.NewWeekdays(time.Weekday(weekday)), gocron.NewAtTimes(gocron.NewAtTime(hour, minute, 0))),
			jobTask,
		)

	case "monthly":
		// 每月指定日期和时间
		day, _ := scheduleConfig["day"].(float64)
		timeStr, _ := scheduleConfig["time"].(string)
		if timeStr == "" {
			timeStr = "02:00"
		}
		hour, minute := parseTime(timeStr)
		job, err = s.scheduler.NewJob(
			gocron.MonthlyJob(1, gocron.NewDaysOfTheMonth(int(day)), gocron.NewAtTimes(gocron.NewAtTime(hour, minute, 0))),
			jobTask,
		)

	case "cron":
		// Cron表达式
		cronExpr, _ := scheduleConfig["expression"].(string)
		if cronExpr == "" {
			return fmt.Errorf("cron expression is required")
		}
		job, err = s.scheduler.NewJob(
			gocron.CronJob(cronExpr, false),
			jobTask,
		)

	default:
		return fmt.Errorf("unsupported schedule type: %s", task.ScheduleType)
	}

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// 保存任务映射
	s.taskJobs[task.ID] = job

	// 更新下次执行时间
	if nextRun, err := job.NextRun(); err == nil {
		task.NextRunAt = &nextRun
		database.DB.Save(task)
	}

	log.Printf("Task added to scheduler: %s (ID: %d, Type: %s)", task.Name, task.ID, task.ScheduleType)
	return nil
}

// RemoveTask 从调度器移除任务
func (s *SchedulerService) RemoveTask(taskID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.taskJobs[taskID]
	if !exists {
		return fmt.Errorf("task not found in scheduler")
	}

	if err := s.scheduler.RemoveJob(job.ID()); err != nil {
		return fmt.Errorf("failed to remove job: %w", err)
	}

	delete(s.taskJobs, taskID)
	log.Printf("Task removed from scheduler: ID %d", taskID)
	return nil
}

// UpdateTask 更新任务调度
func (s *SchedulerService) UpdateTask(task *model.Task) error {
	// 先移除旧任务
	if err := s.RemoveTask(task.ID); err != nil {
		log.Printf("Failed to remove old task: %v", err)
	}

	// 添加新任务
	return s.AddTask(task)
}

// RunTaskNow 立即执行任务
func (s *SchedulerService) RunTaskNow(taskID uint) error {
	// 加载任务
	var task model.Task
	if err := database.DB.Preload("Host").First(&task, taskID).Error; err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}

	// 检查任务锁
	if _, loaded := s.taskLocks.LoadOrStore(taskID, true); loaded {
		return fmt.Errorf("task is already running")
	}
	defer s.taskLocks.Delete(taskID)

	// 执行备份
	ctx := context.Background()
	return s.backupSvc.ExecuteBackup(ctx, &task)
}

// createJobTask 创建任务函数
func (s *SchedulerService) createJobTask(task *model.Task) gocron.Task {
	return gocron.NewTask(func() {
		// 检查任务锁
		if _, loaded := s.taskLocks.LoadOrStore(task.ID, true); loaded {
			log.Printf("Task %s is already running, skipping", task.Name)
			return
		}
		defer s.taskLocks.Delete(task.ID)

		// 重新加载任务（确保使用最新配置）
		var currentTask model.Task
		if err := database.DB.Preload("Host").First(&currentTask, task.ID).Error; err != nil {
			log.Printf("Failed to load task: %v", err)
			return
		}

		// 检查任务是否仍然启用
		if currentTask.Status != 1 {
			log.Printf("Task %s is disabled, skipping", currentTask.Name)
			return
		}

		// 执行备份
		ctx := context.Background()
		if err := s.backupSvc.ExecuteBackup(ctx, &currentTask); err != nil {
			log.Printf("Backup failed for task %s: %v", currentTask.Name, err)
		}

		// 更新下次执行时间
		s.mu.RLock()
		if job, exists := s.taskJobs[task.ID]; exists {
			if nextRun, err := job.NextRun(); err == nil {
				currentTask.NextRunAt = &nextRun
				database.DB.Save(&currentTask)
			}
		}
		s.mu.RUnlock()
	})
}

// GetNextRunTime 获取任务下次执行时间
func (s *SchedulerService) GetNextRunTime(taskID uint) (*time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.taskJobs[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found in scheduler")
	}

	nextRun, err := job.NextRun()
	if err != nil {
		return nil, err
	}

	return &nextRun, nil
}

// parseTime 解析时间字符串（HH:MM格式）
func parseTime(timeStr string) (uint, uint) {
	var hour, minute uint
	fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	return hour, minute
}
