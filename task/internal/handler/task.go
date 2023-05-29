package handler

import (
	"context"
	"fmt"
	"task/internal/repository"
	"task/internal/service"
	"task/pkg/e"
)

type TaskService struct {
}

func NewTaskService() *TaskService {
	return &TaskService{}
}

func (*TaskService) TaskCreate(ctx context.Context, req *service.TaskRequest) (resp *service.CommonResponse, err error) {
	var task repository.Task
	resp = new(service.CommonResponse)
	resp.Code = e.SUCCESS
	err = task.Create(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(e.ERROR)
		resp.Data = err.Error()
		return resp, err
	}
	resp.Msg = e.GetMsg(uint(resp.Code))
	return resp, nil
}

func (*TaskService) TaskDelete(ctx context.Context, req *service.TaskRequest) (resp *service.CommonResponse, err error) {
	var task repository.Task
	resp = new(service.CommonResponse)
	resp.Code = e.SUCCESS
	err = task.Delete(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(e.ERROR)
		resp.Data = err.Error()
		return resp, err
	}
	resp.Msg = e.GetMsg(uint(resp.Code))
	return resp, nil
}

func (*TaskService) TaskUpdate(ctx context.Context, req *service.TaskRequest) (resp *service.CommonResponse, err error) {
	var task repository.Task
	resp = new(service.CommonResponse)
	resp.Code = e.SUCCESS
	err = task.Update(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(e.ERROR)
		resp.Data = err.Error()
		return resp, err
	}
	resp.Msg = e.GetMsg(uint(resp.Code))
	fmt.Println("resp", resp)
	return resp, nil
}

func (*TaskService) TaskShow(ctx context.Context, req *service.TaskRequest) (resp *service.TaskDetailResponse, err error) {
	var task repository.Task
	resp = new(service.TaskDetailResponse)
	resp.Code = e.SUCCESS
	tlist, err := task.Show(req)
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	resp.TaskDetail = repository.BuildTasks(tlist)
	return resp, nil
}
