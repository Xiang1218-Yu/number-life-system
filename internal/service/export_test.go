package service

import (
	"encoding/json"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"number-life-system/internal/domain"
)

// newTestDB 建立一个内存 SQLite 库并迁移出与生产一致的表结构，
// 让 Import 的事务能在不依赖 Postgres 的情况下端到端验证跨表关联。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	if err := db.AutoMigrate(
		&domain.Account{},
		&domain.Subscription{},
		&domain.DigitalFootprint{},
		&domain.DataLocation{},
		&domain.BackupRecord{},
	); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	// 每个测试独占一张干净表，避免共享内存库的脏数据。
	t.Cleanup(func() {
		db.Migrator().DropTable(&domain.BackupRecord{}, &domain.DataLocation{}, &domain.DigitalFootprint{}, &domain.Subscription{}, &domain.Account{})
	})
	return db
}

func ptr[T any](v T) *T { return &v }

// mustTime 用固定值填充 not null 的时间字段，避免污染断言。
func mustTime() time.Time {
	return time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
}

// TestImportRemapsAccountRefsAcrossTables 验证导入旧档案时，订阅、数字足迹、
// 数据位置、备份记录里的 account_ref 被正确重映射到新账户 ID。
// 这是回归测试：旧实现对映射值 +1，导致子表指向别的账户或悬空。
func TestImportRemapsAccountRefsAcrossTables(t *testing.T) {
	db := newTestDB(t)
	svc := &ExportService{DB: db}

	bundle := domain.ImportBundle{
		Accounts: []domain.Account{
			{ID: 1, Platform: "GitHub", Username: "alice", Category: "development", PasswordStrength: "strong", Status: "active"},
			{ID: 2, Platform: "iCloud", Username: "alice@x.com", Category: "cloud", PasswordStrength: "medium", Status: "active"},
		},
		Subscriptions: []domain.Subscription{
			{ID: 10, AccountID: ptr(uint(1)), ServiceName: "GitHub Pro", Amount: 4, Currency: "USD", Cycle: "month", Status: "active"},
			{ID: 11, AccountID: ptr(uint(2)), ServiceName: "iCloud+", Amount: 6, Currency: "CNY", Cycle: "month", Status: "active"},
			// 未关联账户的订阅应保持 nil，不应被错误映射到某个账户。
			{ID: 12, AccountID: nil, ServiceName: "独立服务", Amount: 9, Currency: "CNY", Cycle: "year", Status: "active"},
		},
		Footprints: []domain.DigitalFootprint{
			{ID: 20, AccountID: ptr(uint(1)), EventType: "login", Title: "异地登录", EventAt: mustTime()},
		},
		DataLocations: []domain.DataLocation{
			{ID: 30, AccountID: ptr(uint(2)), Platform: "iCloud", DataType: "photo", SizeGB: 50, Privacy: "private"},
		},
		Backups: []domain.BackupRecord{
			{ID: 40, AccountID: ptr(uint(1)), Platform: "GitHub", Cycle: "month", LastBackupAt: ptr(mustTime()), NextBackupAt: ptr(mustTime().AddDate(0, 1, 0))},
		},
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	if err := svc.Import(7, data); err != nil {
		t.Fatalf("导入失败: %v", err)
	}

	// 期望：账户获得新自增 ID，旧 ID 1->新ID(a1)、2->新ID(a2)。
	var accounts []domain.Account
	if err := db.Order("id asc").Find(&accounts).Error; err != nil {
		t.Fatalf("查询账户失败: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("账户数 = %d, 期望 2", len(accounts))
	}
	// 用平台名而非数值大小锁定，避免与 sqlite 自增从 1 开始的情况误判。
	idByPlatform := map[string]uint{}
	for _, a := range accounts {
		idByPlatform[a.Platform] = a.ID
	}
	a1, a2 := idByPlatform["GitHub"], idByPlatform["iCloud"]

	// 订阅：account_ref 必须精确等于新账户 ID，而不是新 ID+1 或旧 ID。
	var subs []domain.Subscription
	if err := db.Order("id asc").Find(&subs).Error; err != nil {
		t.Fatalf("查询订阅失败: %v", err)
	}
	wantRefs := map[string]*uint{"GitHub Pro": &a1, "iCloud+": &a2, "独立服务": nil}
	for _, sub := range subs {
		want := wantRefs[sub.ServiceName]
		if want == nil {
			if sub.AccountID != nil {
				t.Errorf("订阅 %s 的 account_ref = %v, 期望 nil", sub.ServiceName, *sub.AccountID)
			}
			continue
		}
		if sub.AccountID == nil {
			t.Errorf("订阅 %s 的 account_ref 为 nil, 期望 %d", sub.ServiceName, *want)
		} else if *sub.AccountID != *want {
			// 这正是回归 bug 的断言：+1 会落到相邻账户 a2 上。
			t.Errorf("订阅 %s 的 account_ref = %d, 期望 %d（不可偏移到别的账户）", sub.ServiceName, *sub.AccountID, *want)
		}
	}

	// 数字足迹、数据位置、备份同样校验：ref 精确等于新账户 ID。
	var fp domain.DigitalFootprint
	if err := db.First(&fp).Error; err != nil {
		t.Fatalf("查询足迹失败: %v", err)
	}
	if fp.AccountID == nil || *fp.AccountID != a1 {
		got := (*uint)(nil)
		if fp.AccountID != nil {
			got = fp.AccountID
		}
		t.Errorf("足迹 account_ref = %v, 期望 %d", got, a1)
	}

	var dl domain.DataLocation
	if err := db.First(&dl).Error; err != nil {
		t.Fatalf("查询数据位置失败: %v", err)
	}
	if dl.AccountID == nil || *dl.AccountID != a2 {
		t.Errorf("数据位置 account_ref 指向错误, 期望 %d", a2)
	}

	var bk domain.BackupRecord
	if err := db.First(&bk).Error; err != nil {
		t.Fatalf("查询备份失败: %v", err)
	}
	if bk.AccountID == nil || *bk.AccountID != a1 {
		t.Errorf("备份 account_ref 指向错误, 期望 %d", a1)
	}
}

// TestImportDropsDanglingAccountRef 验证当旧档案里的引用指向一个并未打包进来的账户时，
// 导入后该引用变为 nil（成为独立记录），而不是误指向某个编号相邻的账户。
func TestImportDropsDanglingAccountRef(t *testing.T) {
	db := newTestDB(t)
	svc := &ExportService{DB: db}

	bundle := domain.ImportBundle{
		Accounts: []domain.Account{
			{ID: 5, Platform: "OnlyOne", Username: "u", Category: "other", PasswordStrength: "weak", Status: "active"},
		},
		// 指向未导入的账户 999。
		Subscriptions: []domain.Subscription{
			{ID: 1, AccountID: ptr(uint(999)), ServiceName: "孤儿订阅", Amount: 1, Currency: "CNY", Cycle: "month", Status: "active"},
		},
		Footprints: []domain.DigitalFootprint{
			{ID: 2, AccountID: ptr(uint(998)), EventType: "event", Title: "孤儿足迹", EventAt: mustTime()},
		},
	}
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	if err := svc.Import(1, data); err != nil {
		t.Fatalf("导入失败: %v", err)
	}

	var sub domain.Subscription
	if err := db.First(&sub).Error; err != nil {
		t.Fatalf("查询订阅失败: %v", err)
	}
	if sub.AccountID != nil {
		t.Errorf("孤儿订阅的 account_ref = %d, 期望 nil（不应误指到任何已导入账户）", *sub.AccountID)
	}

	var fp domain.DigitalFootprint
	if err := db.First(&fp).Error; err != nil {
		t.Fatalf("查询足迹失败: %v", err)
	}
	if fp.AccountID != nil {
		t.Errorf("孤儿足迹的 account_ref = %d, 期望 nil", *fp.AccountID)
	}
}

// TestRemapAccountID 纯函数级断言：映射值不得被偏移。
func TestRemapAccountID(t *testing.T) {
	mapping := map[uint]uint{1: 100, 2: 200}

	if got := remapAccountID(ptr(uint(1)), mapping); got == nil || *got != 100 {
		t.Errorf("旧ID 1 应映射到 100, 得到 %v", got)
	}
	if got := remapAccountID(ptr(uint(2)), mapping); got == nil || *got != 200 {
		t.Errorf("旧ID 2 应映射到 200, 得到 %v", got)
	}
	if got := remapAccountID(ptr(uint(999)), mapping); got != nil {
		t.Errorf("未登记的旧ID 应返回 nil, 得到 %v", *got)
	}
	if got := remapAccountID(nil, mapping); got != nil {
		t.Errorf("nil 引用应保持 nil, 得到 %v", *got)
	}
}
