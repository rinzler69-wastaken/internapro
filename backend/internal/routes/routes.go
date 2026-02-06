package routes

import (
	"database/sql"
	"net/http"

	"dsi_interna_sys/internal/handlers"
	"dsi_interna_sys/internal/middleware"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, db *sql.DB) {
	authHandler := handlers.NewAuthHandler(db)
	internHandler := handlers.NewInternHandler(db)
	taskHandler := handlers.NewTaskHandler(db)
	attendanceHandler := handlers.NewAttendanceHandler(db)
	leaveHandler := handlers.NewLeaveHandler(db)
	assessmentHandler := handlers.NewAssessmentHandler(db)
	reportHandler := handlers.NewReportHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)
	settingHandler := handlers.NewSettingHandler(db)
	supervisorHandler := handlers.NewSupervisorHandler(db)
	exportImportHandler := handlers.NewExportImportHandler(db)
	profileHandler := handlers.NewProfileHandler(db)
	passwordResetHandler := handlers.NewPasswordResetHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)

	api := router.PathPrefix("/api").Subrouter()

	// Public
	api.HandleFunc("/health", healthCheck).Methods("GET")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/google", authHandler.StartGoogleOAuth).Methods("GET")
	api.HandleFunc("/auth/google/callback", authHandler.HandleGoogleCallback).Methods("GET")
	api.HandleFunc("/auth/password/forgot", passwordResetHandler.RequestReset).Methods("POST")
	api.HandleFunc("/auth/password/reset", passwordResetHandler.Reset).Methods("POST")

	// Protected
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// Holidays
	protected.HandleFunc("/holidays", handlers.GetHolidays).Methods("GET")

	admin := protected.PathPrefix("").Subrouter()
	admin.Use(middleware.RequireRole("admin"))
	admin.HandleFunc("/auth/register", authHandler.Register).Methods("POST")

	// Auth
	protected.HandleFunc("/auth/me", authHandler.GetCurrentUser).Methods("GET")
	protected.HandleFunc("/auth/2fa/setup", authHandler.Setup2FA).Methods("POST")
	protected.HandleFunc("/auth/2fa/verify", authHandler.Verify2FA).Methods("POST")
	protected.HandleFunc("/auth/2fa/disable", authHandler.Disable2FA).Methods("POST")
	protected.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")

	// Profile
	protected.HandleFunc("/profile", profileHandler.Get).Methods("GET")
	protected.HandleFunc("/profile", profileHandler.Update).Methods("PUT")
	protected.HandleFunc("/profile/password", profileHandler.UpdatePassword).Methods("PUT")

	// Interns
	protected.HandleFunc("/interns", internHandler.GetAll).Methods("GET")
	protected.HandleFunc("/interns", internHandler.Create).Methods("POST")
	protected.HandleFunc("/interns/{id}", internHandler.GetByID).Methods("GET")
	protected.HandleFunc("/interns/{id}", internHandler.Update).Methods("PUT")
	protected.HandleFunc("/interns/{id}", internHandler.Delete).Methods("DELETE")

	// Tasks
	protected.HandleFunc("/tasks", taskHandler.GetAll).Methods("GET")
	protected.HandleFunc("/tasks", taskHandler.Create).Methods("POST")
	protected.HandleFunc("/tasks/intern/{id}", taskHandler.GetByInternID).Methods("GET")
	protected.HandleFunc("/tasks/search-interns", taskHandler.SearchInterns).Methods("GET")
	protected.HandleFunc("/tasks/{id}/attachments", taskHandler.UploadAttachment).Methods("POST")
	protected.HandleFunc("/tasks/{id}/complete", taskHandler.MarkComplete).Methods("POST")
	protected.HandleFunc("/tasks/{id}/status", taskHandler.UpdateStatus).Methods("POST")
	protected.HandleFunc("/tasks/{id}/submit", taskHandler.Submit).Methods("POST")
	protected.HandleFunc("/tasks/{id}/review", taskHandler.Review).Methods("POST")
	protected.HandleFunc("/tasks/{id}", taskHandler.GetByID).Methods("GET")
	protected.HandleFunc("/tasks/{id}", taskHandler.Update).Methods("PUT")
	protected.HandleFunc("/tasks/{id}", taskHandler.Delete).Methods("DELETE")

	// Task Assignments (grouped)
	protected.HandleFunc("/task-assignments", taskHandler.GetAssignments).Methods("GET")
	protected.HandleFunc("/task-assignments/{id}", taskHandler.GetAssignmentByID).Methods("GET")

	// Attendance - ORDER IS CRITICAL HERE
	protected.HandleFunc("/attendance", attendanceHandler.GetAll).Methods("GET")
	protected.HandleFunc("/attendance/today", attendanceHandler.GetToday).Methods("GET")     // SPECIFIC
	protected.HandleFunc("/attendance/checkin", attendanceHandler.CheckIn).Methods("POST")   // SPECIFIC
	protected.HandleFunc("/attendance/checkout", attendanceHandler.CheckOut).Methods("POST") // SPECIFIC
	protected.HandleFunc("/attendance/permission", attendanceHandler.SubmitPermission).Methods("POST")
	protected.HandleFunc("/attendance/intern/{id}", attendanceHandler.GetByInternID).Methods("GET")
	protected.HandleFunc("/attendance/{id}", attendanceHandler.GetByID).Methods("GET") // GENERIC LAST

	// Leaves
	protected.HandleFunc("/leaves", leaveHandler.GetAll).Methods("GET")
	protected.HandleFunc("/leaves", leaveHandler.Create).Methods("POST")
	protected.HandleFunc("/leaves/intern/{id}", leaveHandler.GetByInternID).Methods("GET")
	protected.HandleFunc("/leaves/{id}/approve", leaveHandler.Approve).Methods("POST")
	protected.HandleFunc("/leaves/{id}/reject", leaveHandler.Reject).Methods("POST")
	protected.HandleFunc("/leaves/{id}/attachment", leaveHandler.UploadAttachment).Methods("POST")
	protected.HandleFunc("/leaves/{id}", leaveHandler.GetByID).Methods("GET")
	protected.HandleFunc("/leaves/{id}", leaveHandler.Update).Methods("PUT")

	// Assessments
	protected.HandleFunc("/assessments", assessmentHandler.GetAll).Methods("GET")
	protected.HandleFunc("/assessments", assessmentHandler.Create).Methods("POST")
	protected.HandleFunc("/assessments/intern/{id}", assessmentHandler.GetByInternID).Methods("GET")
	protected.HandleFunc("/assessments/{id}", assessmentHandler.GetByID).Methods("GET")
	protected.HandleFunc("/assessments/{id}", assessmentHandler.Update).Methods("PUT")
	protected.HandleFunc("/assessments/{id}", assessmentHandler.Delete).Methods("DELETE")

	// Reports
	protected.HandleFunc("/reports", reportHandler.GetAll).Methods("GET")
	protected.HandleFunc("/reports", reportHandler.Create).Methods("POST")
	protected.HandleFunc("/reports/{id}", reportHandler.GetByID).Methods("GET")
	protected.HandleFunc("/reports/{id}", reportHandler.Update).Methods("PUT")
	protected.HandleFunc("/reports/{id}", reportHandler.Delete).Methods("DELETE")
	protected.HandleFunc("/reports/{id}/feedback", reportHandler.AddFeedback).Methods("POST")
	protected.HandleFunc("/reports/intern/{id}", reportHandler.GetInternReport).Methods("GET")
	protected.HandleFunc("/reports/attendance/{id}", reportHandler.GetAttendanceReport).Methods("GET")
	protected.HandleFunc("/reports/assessments/{id}", reportHandler.GetAssessmentReport).Methods("GET")
	protected.HandleFunc("/reports/certificate/{id}", reportHandler.GetCertificate).Methods("GET")
	protected.HandleFunc("/reports/certificate/{id}/generate", reportHandler.GenerateCertificate).Methods("POST")

	// Analytics
	analytics := protected.PathPrefix("/analytics").Subrouter()
	analytics.Use(middleware.RequireRole("admin", "pembimbing", "supervisor", "intern"))
	analytics.HandleFunc("/trends/weekly/{id:[0-9]+}", analyticsHandler.GetWeeklyTrends).Methods("GET")
	analytics.HandleFunc("/patterns/checkin/{id:[0-9]+}", analyticsHandler.GetCheckInPatterns).Methods("GET")
	analytics.HandleFunc("/insights/{id:[0-9]+}", analyticsHandler.GetPerformanceInsights).Methods("GET")

	// Dashboard (all authenticated users)
	dashboard := protected.PathPrefix("/dashboard").Subrouter()
	dashboard.HandleFunc("/intern", dashboardHandler.GetInternDashboard).Methods("GET")
	dashboard.HandleFunc("/admin", dashboardHandler.GetAdminDashboard).Methods("GET")

	// Notifications
	protected.HandleFunc("/notifications", notificationHandler.GetAll).Methods("GET")
	protected.HandleFunc("/notifications/{id}/read", notificationHandler.MarkAsRead).Methods("POST")
	protected.HandleFunc("/notifications/mark-all-read", notificationHandler.MarkAllRead).Methods("POST")
	protected.HandleFunc("/notifications/{id}", notificationHandler.Delete).Methods("DELETE")

	// Settings (admin only)
	settings := protected.PathPrefix("/settings").Subrouter()
	settings.Use(middleware.RequireRole("admin"))
	settings.HandleFunc("", settingHandler.GetAll).Methods("GET")
	settings.HandleFunc("", settingHandler.Update).Methods("POST")

	// Supervisors (admin only)
	admin.HandleFunc("/supervisors", supervisorHandler.GetAll).Methods("GET")
	admin.HandleFunc("/supervisors", supervisorHandler.Create).Methods("POST")
	admin.HandleFunc("/supervisors/{id}", supervisorHandler.GetByID).Methods("GET")
	admin.HandleFunc("/supervisors/{id}", supervisorHandler.Update).Methods("PUT")
	admin.HandleFunc("/supervisors/{id}", supervisorHandler.Delete).Methods("DELETE")
	admin.HandleFunc("/supervisors/{id}/approve", supervisorHandler.Approve).Methods("POST")
	admin.HandleFunc("/supervisors/{id}/reject", supervisorHandler.Reject).Methods("POST")

	// Export/Import (admin & pembimbing)
	manager := protected.PathPrefix("").Subrouter()
	manager.Use(middleware.RequireRole("admin", "pembimbing"))
	manager.HandleFunc("/export/interns", exportImportHandler.ExportInterns).Methods("GET")
	manager.HandleFunc("/export/attendances", exportImportHandler.ExportAttendances).Methods("GET")
	manager.HandleFunc("/export/tasks", exportImportHandler.ExportTasks).Methods("GET")
	manager.HandleFunc("/import/interns", exportImportHandler.ImportInterns).Methods("POST")
	manager.HandleFunc("/import/template", exportImportHandler.DownloadTemplate).Methods("GET")
	manager.HandleFunc("/interns/{id}/download-report", reportHandler.DownloadInternReport).Methods("GET")
	manager.HandleFunc("/interns/{id}/certificate", reportHandler.DownloadCertificate).Methods("GET")

	router.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/",
			middleware.AuthMiddleware(
				http.FileServer(http.Dir("./uploads")),
			),
		),
	)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"INTERNA API is running"}`))
}
