package room

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler menangani request HTTP untuk domain Kamar.
type Handler struct {
	service Service
	dbReady func() bool
}

// NewHandler membuat instance baru dari Room Handler.
func NewHandler(service Service, dbReady func() bool) *Handler {
	return &Handler{
		service: service,
		dbReady: dbReady,
	}
}

// GetRooms mengambil semua data kamar milik pengguna yang sedang login.
// Mendukung filter query: ?status=...
func (h *Handler) GetRooms(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	statusQuery := c.Query("status")

	rooms, err := h.service.GetRooms(userID, statusQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kamar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rooms,
	})
}

// GetRoomByID mengambil detail 1 kamar dengan validasi kepemilikan user_id.
func (h *Handler) GetRoomByID(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data kamar tidak ditemukan atau bukan milik Anda"})
		return
	}

	room, err := h.service.GetRoomByID(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kamar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": room,
	})
}

// CreateRoom menambahkan kamar baru untuk pengguna yang login.
func (h *Handler) CreateRoom(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	var input RoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: nomor kamar, tipe, dan harga sewa wajib diisi dengan benar"})
		return
	}

	createdRoom, err := h.service.CreateRoom(userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidRoomStatus) || errors.Is(err, ErrRoomNumberExists) || (err != nil && err.Error() == "Nomor kamar "+input.RoomNumber+" sudah terdaftar di hotel Anda") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data kamar"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kamar berhasil ditambahkan",
		"data":    createdRoom,
	})
}

// UpdateRoom memperbarui data kamar dengan validasi kepemilikan user_id.
func (h *Handler) UpdateRoom(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data kamar tidak ditemukan atau bukan milik Anda"})
		return
	}

	var input RoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	updatedRoom, err := h.service.UpdateRoom(uint(id), userID, input)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrInvalidRoomStatus) || err.Error() == "Nomor kamar sudah digunakan oleh kamar lain" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data kamar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data kamar berhasil diperbarui",
		"data":    updatedRoom,
	})
}

// DeleteRoom menghapus data kamar milik pengguna.
func (h *Handler) DeleteRoom(c *gin.Context) {
	if h.dbReady != nil && !h.dbReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data kamar tidak ditemukan atau bukan milik Anda"})
		return
	}

	err = h.service.DeleteRoom(uint(id), userID)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrRoomOccupied) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kamar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data kamar berhasil dihapus",
	})
}
