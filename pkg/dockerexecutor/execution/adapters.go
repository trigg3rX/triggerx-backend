package execution

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/container"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/file"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

// ContainerManagerAdapter adapts container.ContainerManagerAPI to execution.ContainerManager
type ContainerManagerAdapter struct {
	manager container.ContainerManagerAPI
}

// NewContainerManagerAdapter creates a new adapter for container manager
func NewContainerManagerAdapter(manager container.ContainerManagerAPI) *ContainerManagerAdapter {
	return &ContainerManagerAdapter{
		manager: manager,
	}
}

// GetContainer implements execution.ContainerManager.GetContainer
func (a *ContainerManagerAdapter) GetContainer(ctx context.Context, language types.Language) (*types.PooledContainer, error) {
	return a.manager.GetContainer(ctx, language)
}

// ReturnContainer implements execution.ContainerManager.ReturnContainer
func (a *ContainerManagerAdapter) ReturnContainer(container *types.PooledContainer) error {
	return a.manager.ReturnContainer(container)
}

// ExecuteInContainer implements execution.ContainerManager.ExecuteInContainer
func (a *ContainerManagerAdapter) ExecuteInContainer(ctx context.Context, containerID string, filePath string, language types.Language) (*types.ExecutionResult, string, error) {
	return a.manager.ExecuteInContainer(ctx, containerID, filePath, language)
}

// MarkContainerAsFailed implements execution.ContainerManager.MarkContainerAsFailed
func (a *ContainerManagerAdapter) MarkContainerAsFailed(containerID string, language types.Language, err error) {
	a.manager.MarkContainerAsFailed(containerID, language, err)
}

// KillExecProcess implements execution.ContainerManager.KillExecProcess
func (a *ContainerManagerAdapter) KillExecProcess(ctx context.Context, execID string) error {
	return a.manager.KillExecProcess(ctx, execID)
}

// GetPoolStats implements execution.ContainerManager.GetPoolStats
func (a *ContainerManagerAdapter) GetPoolStats() map[types.Language]*types.PoolStats {
	return a.manager.GetPoolStats()
}

// InitializeLanguagePools implements execution.ContainerManager.InitializeLanguagePools
func (a *ContainerManagerAdapter) InitializeLanguagePools(ctx context.Context, languages []types.Language) error {
	return a.manager.InitializeLanguagePools(ctx, languages)
}

// GetSupportedLanguages implements execution.ContainerManager.GetSupportedLanguages
func (a *ContainerManagerAdapter) GetSupportedLanguages() []types.Language {
	return a.manager.GetSupportedLanguages()
}

// IsLanguageSupported implements execution.ContainerManager.IsLanguageSupported
func (a *ContainerManagerAdapter) IsLanguageSupported(language types.Language) bool {
	return a.manager.IsLanguageSupported(language)
}

// Close implements execution.ContainerManager.Close
func (a *ContainerManagerAdapter) Close(ctx context.Context) error {
	return a.manager.Close(ctx)
}

// FileManagerAdapter adapts file.FileManagerAPI to execution.FileManager
type FileManagerAdapter struct {
	manager file.FileManagerAPI
}

// NewFileManagerAdapter creates a new adapter for file manager
func NewFileManagerAdapter(manager file.FileManagerAPI) *FileManagerAdapter {
	return &FileManagerAdapter{
		manager: manager,
	}
}

// GetOrDownload implements execution.FileManager.GetOrDownload
func (a *FileManagerAdapter) GetOrDownload(ctx context.Context, fileURL string, fileLanguage string) (*types.ExecutionContext, error) {
	return a.manager.GetOrDownload(ctx, fileURL, fileLanguage)
}

// Close implements execution.FileManager.Close
func (a *FileManagerAdapter) Close() error {
	return a.manager.Close()
}
