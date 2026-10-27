package tests

import (
	"math"
	"testing"

	"hotel-app/internal/expense"
)

// TestHotelKPIFormulas menguji formula Occupancy Rate, ADR, RevPAR, dan Laba/Rugi
func TestHotelKPIFormulas(t *testing.T) {
	totalRooms := int64(20)
	occupiedRooms := int64(15)
	totalRoomRevenue := 7500000.0 // Rp 7.500.000
	totalExpenses := 4200000.0    // Rp 4.200.000

	// 1. Occupancy Rate (%) = (Kamar Terisi / Total Kamar) * 100
	occupancyRate := (float64(occupiedRooms) / float64(totalRooms)) * 100
	expectedOccupancy := 75.0 // (15/20) * 100 = 75%
	if math.Abs(occupancyRate-expectedOccupancy) > 0.001 {
		t.Errorf("Ekspektasi Occupancy Rate %.2f%%, didapatkan %.2f%%", expectedOccupancy, occupancyRate)
	}

	// 2. ADR = Total Pendapatan Kamar / Kamar Terjual
	adr := totalRoomRevenue / float64(occupiedRooms)
	expectedADR := 500000.0 // Rp 500.000
	if math.Abs(adr-expectedADR) > 0.001 {
		t.Errorf("Ekspektasi ADR %.2f, didapatkan %.2f", expectedADR, adr)
	}

	// 3. RevPAR = Total Pendapatan Kamar / Total Kamar
	revpar := totalRoomRevenue / float64(totalRooms)
	expectedRevPAR := 375000.0 // Rp 375.000
	if math.Abs(revpar-expectedRevPAR) > 0.001 {
		t.Errorf("Ekspektasi RevPAR %.2f, didapatkan %.2f", expectedRevPAR, revpar)
	}

	// Pastikan properti matematis RevPAR = (OccupancyRate / 100) * ADR
	calculatedRevPAR := (occupancyRate / 100) * adr
	if math.Abs(revpar-calculatedRevPAR) > 0.001 {
		t.Errorf("RevPAR harus sama dengan (Occupancy * ADR), didapatkan %.2f vs %.2f", revpar, calculatedRevPAR)
	}

	// 4. Laba / Rugi Bersih = Pendapatan - Beban Operasional
	netIncome := totalRoomRevenue - totalExpenses
	expectedNetIncome := 3300000.0 // Rp 3.300.000 (Laba Bersih)
	if math.Abs(netIncome-expectedNetIncome) > 0.001 {
		t.Errorf("Ekspektasi Net Income %.2f, didapatkan %.2f", expectedNetIncome, netIncome)
	}
}

// TestUSALICategoryValidation menguji keabsahan kategori standar USALI
func TestUSALICategoryValidation(t *testing.T) {
	validCategories := []string{
		"Utilities",
		"Operating Supplies",
		"Repairs & Maintenance",
		"Salaries & Wages",
		"Marketing & Sales",
		"Administrative & General",
		"Food & Beverage",
		"Other Expenses",
	}

	for _, cat := range validCategories {
		if !expense.IsValidUSALICategory(cat) {
			t.Errorf("Kategori valid '%s' seharusnya diterima oleh validator USALI", cat)
		}
	}

	// Kategori tidak valid
	invalidCategories := []string{"Beli Motor Pribadi", "Liburan", "RandomCategory", ""}
	for _, cat := range invalidCategories {
		if expense.IsValidUSALICategory(cat) {
			t.Errorf("Kategori tidak valid '%s' seharusnya ditolak oleh validator USALI", cat)
		}
	}
}
