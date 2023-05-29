package repository

import (
	"task/internal/service"
	"task/pkg/util"
)

// Task :gorm.Model
type Task struct {
	TaskID    uint `gorm:"primarykey"` // 主键
	UserID    uint `gorm:"index"`      // 多个字段使用相同名称 创建的复合索引
	Status    int  `gorm:"default:0"`
	Title     string
	Context   string `gorm:"type:longtext"`
	StartTime int64
	EndTime   int64
}

func (*Task) Show(req *service.TaskRequest) (taskList []Task, err error) {
	err = DB.Model(Task{}).Where("user_id=?", req.UserID).Find(&taskList).Error
	// 方式 recordNotFound Error
	if taskList == nil {
		return taskList, err
	}
	return taskList, nil
}

// Create 插入一个新的task
func (*Task) Create(req *service.TaskRequest) error {
	task := Task{
		UserID:    uint(req.UserID),
		Title:     req.Title,
		Context:   req.Content,
		StartTime: int64(req.StartTime),
		EndTime:   int64(req.EndTime),
		Status:    int(req.Status),
	}
	if err := DB.Create(&task).Error; err != nil {
		util.LogrusObj.Error("Insert Error:", err.Error())
		return err
	}
	return nil
}

func (*Task) Delete(req *service.TaskRequest) error {
	//err := DB.Where("task_id=?", req.TaskID).Delete(Task{}).Error
	err := DB.Delete(Task{}, req.TaskID).Error
	return err
}

func (*Task) Update(req *service.TaskRequest) error {
	t := Task{}
	err := DB.Where("user_id=?", req.UserID).First(&t).Error
	if err != nil {
		util.LogrusObj.Error("Update get item Error:", err.Error())
		return err
	}
	t.Title = req.Title
	t.Context = req.Content
	t.Status = int(req.Status)
	t.StartTime = int64(req.StartTime)
	t.EndTime = int64(req.EndTime)
	err = DB.Save(&t).Error
	if err != nil {
		util.LogrusObj.Error("Update push item Error:", err.Error())
		return err
	}
	return err
}

// BuildTask 序列化
func BuildTask(item Task) *service.TaskModel {
	return &service.TaskModel{
		TaskID:    uint32(item.TaskID),
		UserID:    uint32(item.UserID),
		Status:    uint32(item.Status),
		Title:     item.Title,
		Content:   item.Context,
		StartTime: uint32(item.StartTime),
		EndTime:   uint32(item.EndTime),
	}
}

func BuildTasks(items []Task) (tlist []*service.TaskModel) {
	for _, v := range items {
		f := BuildTask(v)
		tlist = append(tlist, f)
	}
	return tlist
}
