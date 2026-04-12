package utils

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/api"
)

func GetEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(pattern).MatchString(email)
}

func ProjectIDFromPath(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return ""
	}

	if parts[0] != "projects" {
		return ""
	}

	return parts[1]
}

func ProjectIDFromTaskListPath(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		return ""
	}

	if parts[0] != "projects" || parts[2] != "tasks" {
		return ""
	}

	return parts[1]
}

func TaskIDFromPath(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		return ""
	}

	if parts[0] != "tasks" {
		return ""
	}

	return parts[1]
}

func ValidateCreateTaskRequest(req api.CreateTaskRequest) map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(req.Title) == "" {
		fields["title"] = "is required"
	}

	status := strings.TrimSpace(req.Status)
	if status != "" && !IsValidStatus(status) {
		fields["status"] = "is invalid"
	}

	if strings.TrimSpace(req.Priority) == "" {
		fields["priority"] = "is required"
	} else if !IsValidPriority(req.Priority) {
		fields["priority"] = "is invalid"
	}

	return fields
}

func ValidateUpdateTaskRequest(req api.UpdateTaskRequest) map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(req.Title) == "" {
		fields["title"] = "is required"
	}

	if !IsValidStatus(req.Status) {
		fields["status"] = "is invalid"
	}

	if !IsValidPriority(req.Priority) {
		fields["priority"] = "is invalid"
	}

	return fields
}

func IsValidStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "todo", "in_progress", "done":
		return true
	default:
		return false
	}
}

func IsValidPriority(priority string) bool {
	switch strings.TrimSpace(priority) {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func ParseOptionalDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil || strings.TrimSpace(*dateStr) == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*dateStr))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func NormalizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}


func ValidateRegisterRequest(req api.RegisterRequest) map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "name is required"
	}

	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "email is required"
	} else if !IsValidEmail(req.Email) {
		fields["email"] = "invalid email format"
	}

	if strings.TrimSpace(req.Password) == "" {
		fields["password"] = "is required"
	} else if len(req.Password) < 6 {
		fields["password"] = "must be at least 6 characters"
	}

	return fields
}

func ValidateLoginRequest(req api.LoginRequest) map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	} else if !IsValidEmail(req.Email) {
		fields["email"] = "is invalid"
	}

	if strings.TrimSpace(req.Password) == "" {
		fields["password"] = "is required"
	}
	return fields
}
