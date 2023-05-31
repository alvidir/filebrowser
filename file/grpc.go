package file

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type FileGrpcService struct {
	proto.UnimplementedFileServiceServer
	fileApp   *FileApplication
	logger    *zap.Logger
	uidHeader string
}

func NewFileGrpcServer(fileApp *FileApplication, authHeader string, logger *zap.Logger) *FileGrpcService {
	return &FileGrpcService{
		fileApp:   fileApp,
		logger:    logger,
		uidHeader: authHeader,
	}
}

func NewPermissions(userId int32, perm Permission) *proto.Permissions {
	return &proto.Permissions{
		UserId: userId,
		Read:   perm&Read != 0,
		Write:  perm&Write != 0,
		Owner:  perm&Owner != 0,
	}
}

func NewProtoFile(file *File) *proto.File {
	descriptor := &proto.File{
		Id:          file.id,
		Name:        file.name,
		Metadata:    make([]*proto.Metadata, 0, len(file.metadata)),
		Directory:   file.directory,
		Permissions: make([]*proto.Permissions, 0, len(file.permissions)),
		Flags:       uint32(file.flags),
		Data:        file.data,
	}

	for key, value := range file.metadata {
		descriptor.Metadata = append(descriptor.Metadata, &proto.Metadata{
			Key:   key,
			Value: value,
		})
	}

	for uid, perm := range file.permissions {
		descriptor.Permissions = append(descriptor.Permissions, NewPermissions(uid, perm))
	}

	return descriptor
}

func (server *FileGrpcService) Create(ctx context.Context, req *proto.File) (*proto.File, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.uidHeader, server.logger)
	if err != nil {
		return nil, err
	}

	options := CreateOptions{
		Name:      req.GetName(),
		Directory: req.GetDirectory(),
		Meta:      make(Metadata),
		Data:      req.GetData(),
	}

	for _, meta := range req.GetMetadata() {
		options.Meta[meta.GetKey()] = meta.GetValue()
	}

	file, err := server.fileApp.Create(ctx, uid, &options)
	if err != nil {
		return nil, err
	}

	return NewProtoFile(file), nil
}

func (server *FileGrpcService) Get(ctx context.Context, req *proto.File) (*proto.File, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.uidHeader, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.fileApp.Get(ctx, uid, req.GetId())
	if err != nil {
		return nil, err
	}

	return NewProtoFile(file), nil
}

func (server *FileGrpcService) Update(ctx context.Context, req *proto.File) (*proto.File, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.uidHeader, server.logger)
	if err != nil {
		return nil, err
	}

	options := UpdateOptions{
		Name: req.GetName(),
		Meta: make(Metadata),
		Data: req.GetData(),
	}

	metadata := make(Metadata)
	for _, meta := range req.GetMetadata() {
		metadata[meta.Key] = meta.Value
	}

	file, err := server.fileApp.Update(ctx, uid, req.GetId(), &options)
	if err != nil {
		return nil, err
	}

	return NewProtoFile(file), nil
}

func (server *FileGrpcService) Delete(ctx context.Context, req *proto.File) (*proto.File, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.uidHeader, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.fileApp.Delete(ctx, uid, req.GetId())
	if err != nil {
		return nil, err
	}

	return NewProtoFile(file), nil
}
