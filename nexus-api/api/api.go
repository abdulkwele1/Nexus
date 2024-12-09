package api

type UserCookies = map[string]string

type ErrorResponse struct {
	Error string `json:"error"`
}

type LoginRequest struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type LoginResponse struct {
	RedirectURL string `json:"redirect_url"`
	Match       bool   `json:"match"`
	Cookie      string `json:"cookie"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type Panel struct {
	PanelID          int     `json:"panel_id"`          // Unique ID of the panel
	Name             string  `json:"name"`              // Name of the panel
	Location         string  `json:"location"`          // Location of the panel
	InstallationDate string  `json:"installation_date"` // Installation date in YYYY-MM-DD format
	CapacityKW       float64 `json:"capacity_kW"`       // Capacity in kilowatts
}

// UpdatePanelRequest represents the payload for updating panel details.
type UpdatePanelRequest struct {
	Name             string  `json:"name"`              // Name of the panel
	Location         string  `json:"location"`          // Location of the panel
	InstallationDate string  `json:"installation_date"` // Installation date in YYYY-MM-DD format
	CapacityKW       float64 `json:"capacity_kW"`       // Capacity in kilowatts
}

// ErrorResponse represents a common error response structure.
type ErrorResponse struct {
	Error string `json:"error"` // Error message
}

// SuccessResponse represents a generic success response structure.
type SuccessResponse struct {
	Message string `json:"message"` // Success message
}
