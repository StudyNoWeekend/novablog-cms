package logic

import (
	"context"
	"fmt"
	"time"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/pkg/clouddrive"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CloudImportLogic 云盘导入业务逻辑：创建任务 → goroutine 逐文件
// 「云盘下载 → 流式转存媒体库」，进度写 DB 计数器由前端轮询。
type CloudImportLogic struct {
	importModel  *model.CloudImportModel
	accountModel *model.CloudDriveAccountModel
	driveLogic   *CloudDriveLogic
	mediaLogic   *MediaLogic
	logger       *zap.Logger
}

// NewCloudImportLogic 创建 CloudImportLogic 实例。
func NewCloudImportLogic(driveLogic *CloudDriveLogic, mediaLogic *MediaLogic, logger *zap.Logger) *CloudImportLogic {
	return &CloudImportLogic{
		importModel:  model.NewCloudImport(),
		accountModel: model.NewCloudDriveAccount(),
		driveLogic:   driveLogic,
		mediaLogic:   mediaLogic,
		logger:       logger,
	}
}

// StartImport 发起云盘导入：校验账号与文件上限 → 创建任务与条目 → 异步执行。
func (l *CloudImportLogic) StartImport(ctx context.Context, r *req.CloudImportStartReq) (string, error) {
	account, err := l.accountModel.GetByID(ctx, r.AccountID)
	if err != nil {
		return "", enum.ErrCloudDriveAccountNotFound
	}
	if !clouddrive.Supported(account.Provider) {
		return "", enum.ErrCloudDriveProviderInvalid
	}

	if len(r.Files) > CloudImportBatchLimit() {
		return "", fmt.Errorf("%w（上限 %d 个）", enum.ErrCloudDriveImportLimit, CloudImportBatchLimit())
	}
	maxSize := MaxCloudImportFileSize()
	for _, f := range r.Files {
		if f.Size > maxSize {
			return "", fmt.Errorf("%w：%s（%d MB > %d MB）", enum.ErrCloudDriveFileTooLarge, f.Name, f.Size>>20, maxSize>>20)
		}
	}

	// 同账号串行：避免并发任务互相争抢网盘限速配额
	running, err := l.importModel.GetRunningTaskCountByAccount(ctx, account.ID)
	if err != nil {
		return "", fmt.Errorf("查询导入任务状态失败: %w", err)
	}
	if running > 0 {
		return "", enum.ErrCloudDriveTaskRunning
	}

	// 预检客户端可用（token 过期/账号失效在开始前暴露）
	if _, err := l.driveLogic.BuildClientForAccount(ctx, account); err != nil {
		return "", err
	}

	now := time.Now()
	task := &model.CloudImportTask{
		ID:             uuid.New().String(),
		AccountID:      account.ID,
		Provider:       account.Provider,
		Status:         "running",
		Total:          len(r.Files),
		StartedAt:      &now,
		TargetFolderID: r.FolderID,
	}
	if err := l.importModel.CreateTask(ctx, task); err != nil {
		return "", fmt.Errorf("创建导入任务失败: %w", err)
	}

	items := make([]model.CloudImportItem, 0, len(r.Files))
	for _, f := range r.Files {
		items = append(items, model.CloudImportItem{
			ID:           uuid.New().String(),
			TaskID:       task.ID,
			SourceFileID: f.FileID,
			SourcePath:   f.Path,
			Filename:     f.Name,
			FileSize:     f.Size,
			Status:       "pending",
		})
	}
	if err := l.importModel.CreateItems(ctx, items); err != nil {
		l.failTaskSync(task, "创建导入条目失败: "+err.Error())
		return "", fmt.Errorf("创建导入条目失败: %w", err)
	}

	taskID := task.ID
	go l.runImport(taskID, account.ID)
	return taskID, nil
}

// runImport 异步执行导入（context.Background：任务生命周期独立于请求）。
func (l *CloudImportLogic) runImport(taskID, accountID string) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			l.logger.Error("云盘导入任务 panic", zap.String("task_id", taskID), zap.Any("panic", r))
			l.updateTaskFailed(ctx, taskID, fmt.Sprintf("导入任务 panic: %v", r))
		}
	}()

	task, err := l.importModel.GetTaskByID(ctx, taskID)
	if err != nil {
		l.logger.Error("查询导入任务失败", zap.String("task_id", taskID), zap.Error(err))
		return
	}
	account, err := l.accountModel.GetByID(ctx, accountID)
	if err != nil {
		l.updateTaskFailed(ctx, taskID, "云盘账号不存在")
		return
	}
	client, err := l.driveLogic.BuildClientForAccount(ctx, account)
	if err != nil {
		l.updateTaskFailed(ctx, taskID, "构建云盘客户端失败: "+err.Error())
		return
	}

	items, err := l.importModel.GetItemsByTaskID(ctx, taskID)
	if err != nil {
		l.updateTaskFailed(ctx, taskID, "查询导入条目失败: "+err.Error())
		return
	}

	targetFolderID := ""
	if task.TargetFolderID != nil {
		targetFolderID = *task.TargetFolderID
	}

	var succeeded, failed int
	for _, item := range items {
		if err := l.importOne(ctx, client, &item, targetFolderID); err != nil {
			failed++
			item.Status = "failed"
			item.Error = err.Error()
			l.logger.Error("云盘导入文件失败",
				zap.String("task_id", taskID),
				zap.String("file", item.Filename),
				zap.Error(err))
		} else {
			succeeded++
			item.Status = "success"
		}
		if uerr := l.importModel.UpdateItem(ctx, &item); uerr != nil {
			l.logger.Error("更新导入条目失败", zap.Error(uerr))
		}

		// 每处理完一个文件立即回写进度计数（前端轮询依赖）
		task.Succeeded = succeeded
		task.Failed = failed
		if uerr := l.importModel.UpdateTask(ctx, task); uerr != nil {
			l.logger.Error("更新导入进度失败", zap.Error(uerr))
		}
	}

	finishedAt := time.Now()
	task.FinishedAt = &finishedAt
	if failed > 0 {
		task.Status = "failed"
		task.Error = fmt.Sprintf("导入完成，但 %d 个文件失败", failed)
	} else {
		task.Status = "completed"
	}
	if err := l.importModel.UpdateTask(ctx, task); err != nil {
		l.logger.Error("更新导入任务状态失败", zap.Error(err))
	}
}

// importOne 导入单个文件：云盘下载 → 流式写入媒体库 → 回填 media_id。
func (l *CloudImportLogic) importOne(ctx context.Context, client clouddrive.Client, item *model.CloudImportItem, targetFolderID string) error {
	reader, size, contentType, err := client.Download(ctx, item.SourceFileID)
	if err != nil {
		return err
	}
	defer reader.Close()

	filename := item.Filename
	if filename == "" {
		filename = item.SourceFileID
	}
	if size <= 0 {
		size = item.FileSize
	}

	media, err := l.mediaLogic.SaveRemoteStream(ctx, filename, contentType, size, reader,
		MediaUploadOptions{FolderID: targetFolderID})
	if err != nil {
		return err
	}
	mediaID := media.ID
	item.MediaID = &mediaID
	return nil
}

// GetImportStatus 查询导入进度与失败明细。
func (l *CloudImportLogic) GetImportStatus(ctx context.Context, taskID string) (*res.CloudImportStatusRes, error) {
	task, err := l.importModel.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询导入任务失败: %w", err)
	}
	items, err := l.importModel.GetItemsByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询导入条目失败: %w", err)
	}

	failed := make([]res.CloudImportItemRes, 0)
	for _, item := range items {
		if item.Status == "failed" {
			failed = append(failed, l.toItemRes(item))
		}
	}
	return &res.CloudImportStatusRes{
		Task:   l.toTaskRes(task),
		Failed: failed,
	}, nil
}

// GetImportTasks 分页查询导入任务历史。
func (l *CloudImportLogic) GetImportTasks(ctx context.Context, page, pageSize int) ([]res.CloudImportTaskRes, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	tasks, total, err := l.importModel.GetTasks(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询导入任务列表失败: %w", err)
	}
	result := make([]res.CloudImportTaskRes, 0, len(tasks))
	for i := range tasks {
		result = append(result, l.toTaskRes(&tasks[i]))
	}
	return result, total, nil
}

// toItemRes 条目模型转 DTO。
func (l *CloudImportLogic) toItemRes(item model.CloudImportItem) res.CloudImportItemRes {
	return res.CloudImportItemRes{
		ID:           item.ID,
		SourceFileID: item.SourceFileID,
		SourcePath:   item.SourcePath,
		Filename:     item.Filename,
		FileSize:     item.FileSize,
		Status:       item.Status,
		MediaID:      ptrToString(item.MediaID),
		Error:        item.Error,
	}
}

// toTaskRes 任务模型转 DTO。
func (l *CloudImportLogic) toTaskRes(task *model.CloudImportTask) res.CloudImportTaskRes {
	return res.CloudImportTaskRes{
		ID:             task.ID,
		AccountID:      task.AccountID,
		Provider:       task.Provider,
		TargetFolderID: task.TargetFolderID,
		Status:         task.Status,
		Total:          task.Total,
		Succeeded:      task.Succeeded,
		Failed:         task.Failed,
		StartedAt:      task.StartedAt,
		FinishedAt:     task.FinishedAt,
		Error:          task.Error,
		CreatedAt:      task.CreatedAt,
	}
}

// updateTaskFailed 标记任务失败（goroutine 内使用）。
func (l *CloudImportLogic) updateTaskFailed(ctx context.Context, taskID, errMsg string) {
	task, err := l.importModel.GetTaskByID(ctx, taskID)
	if err != nil {
		l.logger.Error("查询任务失败", zap.String("task_id", taskID), zap.Error(err))
		return
	}
	task.Status = "failed"
	task.Error = errMsg
	finishedAt := time.Now()
	task.FinishedAt = &finishedAt
	if err := l.importModel.UpdateTask(ctx, task); err != nil {
		l.logger.Error("更新任务状态失败", zap.Error(err))
	}
}

// failTaskSync 同步标记任务失败（StartImport 前置错误）。
func (l *CloudImportLogic) failTaskSync(task *model.CloudImportTask, errMsg string) {
	task.Status = "failed"
	task.Error = errMsg
	finishedAt := time.Now()
	task.FinishedAt = &finishedAt
	if err := l.importModel.UpdateTask(context.Background(), task); err != nil {
		l.logger.Error("更新任务状态失败", zap.Error(err))
	}
}

// RecoverInterruptedTasks 启动自愈：将上次运行遗留的 running 任务置为失败。
func (l *CloudImportLogic) RecoverInterruptedTasks(ctx context.Context) {
	count, err := l.importModel.RecoverRunningTasks(ctx)
	if err != nil {
		l.logger.Error("恢复中断导入任务失败", zap.Error(err))
		return
	}
	if count > 0 {
		l.logger.Warn("已将上次服务中断遗留的导入任务标记为失败", zap.Int64("count", count))
	}
}
