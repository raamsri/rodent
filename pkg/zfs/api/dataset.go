/*
 * Copyright 2024-2025 Raamsri Kumar <raam@tinkershack.in>
 * Copyright 2024-2025 The StrataSTOR Authors and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stratastor/rodent/pkg/errors"
	"github.com/stratastor/rodent/pkg/zfs/dataset"
)

func NewDatasetHandler(manager *dataset.Manager) *DatasetHandler {
	return &DatasetHandler{manager: manager}
}

func (h *DatasetHandler) listDatasets(c *gin.Context) {
	var req dataset.ListConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	result, err := h.manager.List(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *DatasetHandler) listAll(c *gin.Context) {
	var req dataset.ListConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	req.Type = "all"

	result, err := h.manager.List(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *DatasetHandler) createFilesystem(c *gin.Context) {
	var req dataset.FilesystemConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	if err := h.manager.CreateFilesystem(c.Request.Context(), req); err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *DatasetHandler) listFilesystems(c *gin.Context) {
	var req dataset.ListConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	req.Type = "filesystem"

	result, err := h.manager.List(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *DatasetHandler) listVolumes(c *gin.Context) {
	var req dataset.ListConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	req.Type = "volume"

	result, err := h.manager.List(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *DatasetHandler) createVolume(c *gin.Context) {
	var cfg dataset.VolumeConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	if err := h.manager.CreateVolume(c.Request.Context(), cfg); err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *DatasetHandler) destroyDataset(c *gin.Context) {
	var req dataset.DestroyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	if err := h.manager.Destroy(c.Request.Context(), req); err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *DatasetHandler) getProperty(c *gin.Context) {
	var req dataset.PropertyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	prop, err := h.manager.GetProperty(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": prop})
}

func (h *DatasetHandler) setProperty(c *gin.Context) {
	var req dataset.SetPropertyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	if err := h.manager.SetProperty(c.Request.Context(), req); err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *DatasetHandler) inheritProperty(c *gin.Context) {
	var req dataset.InheritConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	if err := h.manager.InheritProperty(c.Request.Context(), req); err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

// List all properties
func (h *DatasetHandler) listProperties(c *gin.Context) {
	var req dataset.NameConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	props, err := h.manager.ListProperties(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": props})
}

// Snapshot operations
func (h *DatasetHandler) createSnapshot(c *gin.Context) {
	var req dataset.SnapshotConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.CreateSnapshot(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *DatasetHandler) listSnapshots(c *gin.Context) {
	var req dataset.ListConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	req.Type = "snapshot"

	result, err := h.manager.List(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *DatasetHandler) rollbackSnapshot(c *gin.Context) {
	var req dataset.RollbackConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.Rollback(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// Clone operations
func (h *DatasetHandler) createClone(c *gin.Context) {
	var req dataset.CloneConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.Clone(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

// Rename operation
func (h *DatasetHandler) renameDataset(c *gin.Context) {
	var cfg dataset.RenameConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	if err := h.manager.Rename(c.Request.Context(), cfg); err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *DatasetHandler) sendDataset(c *gin.Context) {
	var req dataset.TransferConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	// Execute transfer
	err := h.manager.SendReceive(c.Request.Context(), req.SendConfig, req.ReceiveConfig)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *DatasetHandler) getResumeToken(c *gin.Context) {
	var req dataset.NameConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	token, err := h.manager.GetResumeToken(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": token})
}

func (h *DatasetHandler) mountDataset(c *gin.Context) {
	var req dataset.MountConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.Mount(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *DatasetHandler) unmountDataset(c *gin.Context) {
	var req dataset.UnmountConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.Unmount(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Bookmark operations
func (h *DatasetHandler) createBookmark(c *gin.Context) {
	var req dataset.BookmarkConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.CreateBookmark(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

// List bookmarks
func (h *DatasetHandler) listBookmarks(c *gin.Context) {
	var req dataset.ListConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	req.Type = "bookmark"

	result, err := h.manager.List(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// Promote clone
func (h *DatasetHandler) promoteClone(c *gin.Context) {
	var req dataset.NameConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.PromoteClone(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// Show differences between snapshots
func (h *DatasetHandler) diffDataset(c *gin.Context) {
	var req dataset.DiffConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	result, err := h.manager.Diff(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// Allow permissions
func (h *DatasetHandler) allowPermissions(c *gin.Context) {
	var req dataset.AllowConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.Allow(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

// Remove permissions
func (h *DatasetHandler) unallowPermissions(c *gin.Context) {
	var req dataset.UnallowConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.Unallow(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List permissions
func (h *DatasetHandler) listPermissions(c *gin.Context) {
	var req dataset.NameConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	result, err := h.manager.ListPermissions(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// Share dataset
func (h *DatasetHandler) shareDataset(c *gin.Context) {
	var req dataset.ShareConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.Share(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// Unshare dataset
func (h *DatasetHandler) unshareDataset(c *gin.Context) {
	var req dataset.UnshareConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		APIError(c, errors.New(errors.ServerRequestValidation, err.Error()))
		return
	}

	err := h.manager.Unshare(c.Request.Context(), req)
	if err != nil {
		APIError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
