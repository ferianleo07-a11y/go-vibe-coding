package report

// CategoryExpenseBreakdown mendefinisikan rekapitulasi total nominal per kategori biaya.
type CategoryExpenseBreakdown struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

// DashboardMetricsResponse memuat indikator metrik KPI hotel dan rekapitulasi Laba/Rugi.
type DashboardMetricsResponse struct {
	TotalRooms         int64                      `json:"total_rooms"`
	AvailableRooms     int64                      `json:"available_rooms"`
	OccupiedRooms      int64                      `json:"occupied_rooms"`
	CleaningRooms      int64                      `json:"cleaning_rooms"`
	MaintenanceRooms   int64                      `json:"maintenance_rooms"`
	OccupancyRate      float64                    `json:"occupancy_rate"` // %
	RoomRevenue        float64                    `json:"room_revenue"`
	ADR                float64                    `json:"adr"`    // Average Daily Rate
	RevPAR             float64                    `json:"revpar"` // Revenue per Available Room
	TotalExpenses      float64                    `json:"total_expenses"`
	NetIncome          float64                    `json:"net_income"` // Laba / Rugi Bersih
	ExpensesByCategory []CategoryExpenseBreakdown `json:"expenses_by_category"`
}
