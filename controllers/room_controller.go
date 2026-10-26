package controllers

import (
	"net/http"

	"hotel-app/config"
	"hotel-app/models"

	"github.com/gin-gonic/gin"
)

type RoomInput struct {
	RoomNumber    string  `json:"room_number" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	PricePerNight float64 `json:"price_per_night" binding:"required,gt=0"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
}

// GetRooms mengambil semua data kamar milik pengguna yang sedang login.
// Mendukung filter query: ?status=...
func GetRooms(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	statusQuery := c.Query("status")

	query := config.DB.Where("user_id = ?", userID)
	if statusQuery != "" {
		query = query.Where("status = ?", statusQuery)
	}

	var rooms []models.Room
	if err := query.Order("room_number asc").Find(&rooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kamar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rooms,
	})
}

// GetRoomByID mengambil detail 1 kamar dengan validasi kepemilikan user_id.
func GetRoomByID(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	roomID := c.Param("id")

	var room models.Room
	if err := config.DB.Where("id = ? AND user_id = ?", roomID, userID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data kamar tidak ditemukan atau bukan milik Anda"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": room,
	})
}

// CreateRoom menambahkan kamar baru untuk pengguna yang login.
func CreateRoom(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	var input RoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: nomor kamar, tipe, dan harga sewa wajib diisi dengan benar"})
		return
	}

	// Status default adalah 'Tersedia' jika tidak disertakan
	status := input.Status
	if status == "" {
		status = models.RoomStatusAvailable
	} else if !models.IsValidRoomStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status tidak valid. Pilihan: Tersedia, Terisi, Sedang Dibersihkan, Dalam Perbaikan",
		})
		return
	}

	// Cek apakah nomor kamar sudah terdaftar pada pengguna ini
	var existing models.Room
	if err := config.DB.Where("user_id = ? AND room_number = ?", userID, input.RoomNumber).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor kamar " + input.RoomNumber + " sudah terdaftar di hotel Anda"})
		return
	}

	room := models.Room{
		UserID:        userID,
		RoomNumber:    input.RoomNumber,
		Type:          input.Type,
		PricePerNight: input.PricePerNight,
		Status:        status,
		Description:   input.Description,
	}

	if err := config.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data kamar"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kamar berhasil ditambahkan",
		"data":    room,
	})
}

// UpdateRoom memperbarui data kamar dengan validasi kepemilikan user_id.
func UpdateRoom(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	roomID := c.Param("id")

	var room models.Room
	if err := config.DB.Where("id = ? AND user_id = ?", roomID, userID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data kamar tidak ditemukan atau bukan milik Anda"})
		return
	}

	var input RoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	if input.Status != "" && !models.IsValidRoomStatus(input.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status tidak valid. Pilihan: Tersedia, Terisi, Sedang Dibersihkan, Dalam Perbaikan",
		})
		return
	}

	// Cek duplikasi nomor kamar jika nomor diubah
	if input.RoomNumber != room.RoomNumber {
		var existing models.Room
		if err := config.DB.Where("user_id = ? AND room_number = ? AND id != ?", userID, input.RoomNumber, room.ID).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor kamar sudah digunakan oleh kamar lain"})
			return
		}
	}

	room.RoomNumber = input.RoomNumber
	room.Type = input.Type
	room.PricePerNight = input.PricePerNight
	if input.Status != "" {
		room.Status = input.Status
	}
	room.Description = input.Description

	if err := config.DB.Save(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data kamar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data kamar berhasil diperbarui",
		"data":    room,
	})
}

// DeleteRoom menghapus data kamar milik pengguna.
func DeleteRoom(c *gin.Context) {
	if config.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	roomID := c.Param("id")

	var room models.Room
	if err := config.DB.Where("id = ? AND user_id = ?", roomID, userID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data kamar tidak ditemukan atau bukan milik Anda"})
		return
	}

	// Cegah penghapusan jika kamar sedang terisi oleh tamu
	if room.Status == models.RoomStatusOccupied {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kamar sedang terisi oleh tamu, selesaikan proses check-out terlebih dahulu sebelum menghapus",
		})
		return
	}

	if err := config.DB.Delete(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kamar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data kamar berhasil dihapus",
	})
}
