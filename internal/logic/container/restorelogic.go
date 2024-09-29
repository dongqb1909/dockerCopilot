package container

import (
	"context"
	"github.com/google/uuid"
	"github.com/onlyLTY/dockerCopilot/internal/utiles"

	"github.com/onlyLTY/dockerCopilot/internal/svc"
	"github.com/onlyLTY/dockerCopilot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRestoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreLogic {
	return &RestoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RestoreLogic) Restore(req *types.ContainerRestoreReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	taskID := uuid.New().String()
	fileName := CleanFilename(req.Filename)
	go func() {
		// Catch any panic and log the error
		defer func() {
			if r := recover(); r != nil {
				l.Errorf("Recovered from panic in restoreContainer: %v", r)
			}
		}()
		err := utiles.RestoreContainer(l.svcCtx, fileName, taskID)
		if err != nil {
			l.Errorf("Error in restoreContainer: %v", err)
		}
	}()
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = map[string]interface{}{
		"taskID": taskID,
	}
	return resp, nil
}
