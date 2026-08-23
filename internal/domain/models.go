package domain

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type Account struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	UserID            uint       `gorm:"index;not null" json:"user_id"`
	Platform          string     `gorm:"size:120;not null" json:"platform"`
	Username          string     `gorm:"size:255;not null" json:"username"`
	Email             string     `gorm:"size:255" json:"email"`
	Category          string     `gorm:"size:60;not null" json:"category"`
	RegisteredAt      *time.Time `json:"registered_at"`
	PasswordStrength  string     `gorm:"size:20;not null" json:"password_strength"`
	PasswordChangedAt *time.Time `json:"password_changed_at"`
	TwoFactorEnabled  bool       `gorm:"not null;default:false" json:"two_factor_enabled"`
	KnownBreach       bool       `gorm:"not null;default:false" json:"known_breach"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	Notes             string     `gorm:"type:text" json:"notes"`
	Status            string     `gorm:"size:20;not null;default:'active'" json:"status"`
	ArchivedAt        *time.Time `json:"archived_at"`
	ArchiveReason     string     `gorm:"size:255" json:"archive_reason"`
	PasswordHash      string     `gorm:"size:255" json:"-"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
type Subscription struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"index;not null" json:"user_id"`
	AccountID     *uint      `gorm:"index" json:"account_ref"`
	ServiceName   string     `gorm:"size:120;not null" json:"service_name"`
	Plan          string     `gorm:"size:80" json:"plan"`
	Amount        float64    `gorm:"not null" json:"amount"`
	Currency      string     `gorm:"size:10;not null;default:'CNY'" json:"currency"`
	Cycle         string     `gorm:"size:20;not null" json:"cycle"`
	Status        string     `gorm:"size:20;not null;default:'active'" json:"status"`
	StartedAt     *time.Time `json:"started_at"`
	NextBillingAt *time.Time `json:"next_billing_at"`
	CancelledAt   *time.Time `json:"cancelled_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
type DigitalFootprint struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	AccountID   *uint     `gorm:"index" json:"account_ref"`
	EventType   string    `gorm:"size:40;not null" json:"event_type"`
	Title       string    `gorm:"size:160;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	EventAt     time.Time `gorm:"index;not null" json:"event_at"`
	Important   bool      `gorm:"not null;default:false" json:"important"`
	CreatedAt   time.Time `json:"created_at"`
}
type DataLocation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	AccountID *uint     `gorm:"index" json:"account_ref"`
	Platform  string    `gorm:"size:120;not null" json:"platform"`
	DataType  string    `gorm:"size:40;not null" json:"data_type"`
	SizeGB    float64   `gorm:"not null" json:"size_gb"`
	Privacy   string    `gorm:"size:20;not null" json:"privacy"`
	Notes     string    `gorm:"type:text" json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type BackupRecord struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"index;not null" json:"user_id"`
	AccountID    *uint      `gorm:"index" json:"account_ref"`
	Platform     string     `gorm:"size:120;not null" json:"platform"`
	Cycle        string     `gorm:"size:20;not null" json:"cycle"`
	LastBackupAt *time.Time `json:"last_backup_at"`
	NextBackupAt *time.Time `json:"next_backup_at"`
	Notes        string     `gorm:"type:text" json:"notes"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Type      string    `gorm:"size:40;not null" json:"type"`
	Title     string    `gorm:"size:160;not null" json:"title"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	DueAt     time.Time `gorm:"index;not null" json:"due_at"`
	Channel   string    `gorm:"size:20;not null;default:'console'" json:"channel"`
	Status    string    `gorm:"size:20;not null;default:'pending'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
type AccountCategory struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"uniqueIndex;size:60;not null" json:"name"`
	Color string `gorm:"size:20" json:"color"`
}
type SecurityEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	AccountID uint      `gorm:"index;not null" json:"account_id"`
	Type      string    `gorm:"size:40;not null" json:"type"`
	Detail    string    `gorm:"type:text" json:"detail"`
	EventAt   time.Time `gorm:"index;not null" json:"event_at"`
}
type ImportBundle struct {
	Accounts      []Account          `json:"accounts"`
	Subscriptions []Subscription     `json:"subscriptions"`
	Footprints    []DigitalFootprint `json:"footprints"`
	DataLocations []DataLocation     `json:"data_locations"`
	Backups       []BackupRecord     `json:"backups"`
}
