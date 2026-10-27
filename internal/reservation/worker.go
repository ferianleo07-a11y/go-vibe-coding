package reservation

import (
	"context"
	"log"
	"time"
)

// StartReminderWorker menjalankan background goroutine untuk memeriksa dan mengirimkan
// reminder WA dan Email secara otomatis 15 menit sebelum waktu checkout tamu.
func StartReminderWorker(ctx context.Context, service Service, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		log.Printf("[Reminder Worker] Background worker pengingat check-out 15 menit aktif (interval pengecekan: %v)\n", interval)
		for {
			select {
			case <-ctx.Done():
				log.Println("[Reminder Worker] Background worker dihentikan.")
				return
			case <-ticker.C:
				count, err := service.ProcessUpcomingCheckoutReminders()
				if err != nil {
					// Abaikan log jika DB belum siap
					log.Printf("[Reminder Worker] Info pengecekan: %v\n", err)
				} else if count > 0 {
					log.Printf("[Reminder Worker] Berhasil memproses %d pengingat checkout otomatis.\n", count)
				}
			}
		}
	}()
}
