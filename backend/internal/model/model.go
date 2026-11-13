package model

// Product 商品。
type Product struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Category      string  `json:"category"`
	Supplier      string  `json:"supplier"`
	PurchaseCost  float64 `json:"purchaseCost"`
	PackagingCost float64 `json:"packagingCost"`
	SalePrice     float64 `json:"salePrice"`
	Stock         int     `json:"stock"`
	SafeStock     int     `json:"safeStock"`
	Note          string  `json:"note,omitempty"`
	CreatedAt     string  `json:"createdAt"`
}

// ProductInput 新建/编辑商品入参。
type ProductInput struct {
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Category      string  `json:"category"`
	Supplier      string  `json:"supplier"`
	PurchaseCost  float64 `json:"purchaseCost"`
	PackagingCost float64 `json:"packagingCost"`
	SalePrice     float64 `json:"salePrice"`
	Stock         int     `json:"stock"`
	SafeStock     int     `json:"safeStock"`
	Note          string  `json:"note"`
}

type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type OrderItem struct {
	ProductID   int64   `json:"productId"`
	ProductName string  `json:"productName"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	SalePrice   float64 `json:"salePrice"`
	TotalAmount float64 `json:"totalAmount"`
	TotalCost   float64 `json:"totalCost"`
}

type Order struct {
	ID             int64       `json:"id"`
	OrderNo        string      `json:"orderNo"`
	Channel        string      `json:"channel"`
	CustomerName   string      `json:"customerName"`
	Items          []OrderItem `json:"items"`
	TotalAmount    float64     `json:"totalAmount"`
	DiscountAmount float64     `json:"discountAmount"`
	ShippingCost   float64     `json:"shippingCost"`
	PlatformFee    float64     `json:"platformFee"`
	GoodsCost      float64     `json:"goodsCost"`
	NetProfit      float64     `json:"netProfit"`
	CreatedAt      string      `json:"createdAt"`
	Status         string      `json:"status"`
}

// NewOrderInput 新建/编辑订单入参。
type NewOrderInput struct {
	ProductID       int64   `json:"productId"`
	CustomerName    string  `json:"customerName"`
	Channel         string  `json:"channel"`
	Quantity        int     `json:"quantity"`
	DiscountAmount  float64 `json:"discountAmount"`
	ShippingCost    float64 `json:"shippingCost"`
	PlatformFeeRate float64 `json:"platformFeeRate"`
}

type PricingRule struct {
	ID             int64   `json:"id"`
	TargetMargin   float64 `json:"targetMargin"`
	ChannelFeeRate float64 `json:"channelFeeRate"`
	ExtraCost      float64 `json:"extraCost"`
	RoundingMode   string  `json:"roundingMode"`
}

type LabInspiration struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Tag       string `json:"tag"`
	Status    string `json:"status"`
	Summary   string `json:"summary"`
	Keywords  string `json:"keywords"`
	Image     string `json:"image,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type InspirationInput struct {
	Title    string `json:"title"`
	Tag      string `json:"tag"`
	Status   string `json:"status"`
	Summary  string `json:"summary"`
	Keywords string `json:"keywords"`
	Image    string `json:"image"`
}

type ProofRecord struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Date  string `json:"date"`
	Note  string `json:"note"`
	Image string `json:"image,omitempty"`
}

type ProofRecordInput struct {
	Title string `json:"title"`
	Date  string `json:"date"`
	Note  string `json:"note"`
	Image string `json:"image"`
}

type LabProject struct {
	ID            int64         `json:"id"`
	Name          string        `json:"name"`
	Category      string        `json:"category"`
	Stage         string        `json:"stage"`
	Progress      int           `json:"progress"`
	LaunchDate    string        `json:"launchDate"`
	Note          string        `json:"note"`
	Image         string        `json:"image,omitempty"`
	SampleRecords []ProofRecord `json:"sampleRecords"`
	UpdatedAt     string        `json:"updatedAt"`
}

type LabProjectInput struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Stage      string `json:"stage"`
	Progress   int    `json:"progress"`
	LaunchDate string `json:"launchDate"`
	Note       string `json:"note"`
	Image      string `json:"image"`
}

type DrawingInspirationItem struct {
	Category string `json:"category"`
	Element  string `json:"element"`
}

type DrawingInspirationRecord struct {
	ID        int64                    `json:"id"`
	Items     []DrawingInspirationItem `json:"items"`
	Styles    []string                 `json:"styles"`
	CreatedAt string                   `json:"createdAt"`
	Completed bool                     `json:"completed"`
}

type ExpenseRecord struct {
	ID        int64   `json:"id"`
	Purpose   string  `json:"purpose"`
	Amount    float64 `json:"amount"`
	CreatedAt string  `json:"createdAt"`
}

type ExpenseInput struct {
	Purpose string  `json:"purpose"`
	Amount  float64 `json:"amount"`
}

type MarketRecord struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Date      string  `json:"date"`
	Location  string  `json:"location"`
	Income    float64 `json:"income"`
	Note      string  `json:"note,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

type MarketInput struct {
	Name     string  `json:"name"`
	Date     string  `json:"date"`
	Location string  `json:"location"`
	Income   float64 `json:"income"`
	Note     string  `json:"note"`
}

type AppSetting struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updatedAt"`
}

type PlannerModuleOptions struct {
	Rows      int      `json:"rows,omitempty"`
	Days      int      `json:"days,omitempty"`
	StartHour int      `json:"startHour,omitempty"`
	EndHour   int      `json:"endHour,omitempty"`
	CellMm    float64  `json:"cellMm,omitempty"`
	LineMm    float64  `json:"lineMm,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

type PlannerModuleConfig struct {
	ID      string               `json:"id"`
	Type    string               `json:"type"`
	Title   string               `json:"title"`
	Weight  int                  `json:"weight"`
	Accent  string               `json:"accent"`
	Options PlannerModuleOptions `json:"options"`
}

type PlannerTemplate struct {
	ID        int64                 `json:"id"`
	Name      string                `json:"name"`
	SizeKey   string                `json:"sizeKey"`
	WidthMm   float64               `json:"widthMm"`
	HeightMm  float64               `json:"heightMm"`
	PaddingMm float64               `json:"paddingMm"`
	Accent    string                `json:"accent"`
	Paper     string                `json:"paper"`
	Modules   []PlannerModuleConfig `json:"modules"`
	CreatedAt string                `json:"createdAt"`
	UpdatedAt string                `json:"updatedAt"`
}

type PlannerTemplateInput struct {
	Name      string                `json:"name"`
	SizeKey   string                `json:"sizeKey"`
	WidthMm   float64               `json:"widthMm"`
	HeightMm  float64               `json:"heightMm"`
	PaddingMm float64               `json:"paddingMm"`
	Accent    string                `json:"accent"`
	Paper     string                `json:"paper"`
	Modules   []PlannerModuleConfig `json:"modules"`
}

type InventoryTransaction struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"productId"`
	Barcode   string `json:"barcode,omitempty"`
	Type      string `json:"type"`
	Quantity  int    `json:"quantity"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type InventoryTransactionInput struct {
	ProductID int64  `json:"productId"`
	Barcode   string `json:"barcode"`
	Type      string `json:"type"`
	Quantity  int    `json:"quantity"`
	Note      string `json:"note"`
}

type InventoryProcessInput struct {
	ProductID int64  `json:"productId"`
	Type      string `json:"type"`
	Quantity  int    `json:"quantity"`
	Barcode   string `json:"barcode"`
	Note      string `json:"note"`
}

type ScanItem struct {
	ProductID int64  `json:"productId"`
	Quantity  int    `json:"quantity"`
	Barcode   string `json:"barcode"`
	Note      string `json:"note"`
}

type ScanBatchInput struct {
	Items []ScanItem `json:"items"`
	Type  string     `json:"type"`
	Note  string     `json:"note"`
}

type SKURequest struct {
	Category         string `json:"category"`
	ExcludeProductID int64  `json:"excludeProductId"`
	CreateIfMissing  bool   `json:"createIfMissing"`
}

// Bootstrap 一次性下发给前端的所有业务数据。
type Bootstrap struct {
	Products              []Product                  `json:"products"`
	Categories            []Category                 `json:"categories"`
	Orders                []Order                    `json:"orders"`
	PricingRule           PricingRule                `json:"pricingRule"`
	LabInspirations       []LabInspiration           `json:"labInspirations"`
	LabProjects           []LabProject               `json:"labProjects"`
	DrawingInspirations   []DrawingInspirationRecord `json:"drawingInspirations"`
	Expenses              []ExpenseRecord            `json:"expenses"`
	InventoryTransactions []InventoryTransaction     `json:"inventoryTransactions"`
	Markets               []MarketRecord             `json:"markets"`
	OnlineIncome          float64                    `json:"onlineIncome"`
	OnlineIncomeUpdatedAt string                     `json:"onlineIncomeUpdatedAt"`
	PlannerTemplates      []PlannerTemplate          `json:"plannerTemplates"`
}

// Snapshot 数据备份/恢复使用的完整快照。
type Snapshot struct {
	Version               int                        `json:"version"`
	ExportedAt            string                     `json:"exportedAt"`
	Products              []Product                  `json:"products"`
	Categories            []Category                 `json:"categories"`
	Orders                []Order                    `json:"orders"`
	PricingRule           PricingRule                `json:"pricingRule"`
	LabInspirations       []LabInspiration           `json:"labInspirations"`
	LabProjects           []LabProject               `json:"labProjects"`
	DrawingInspirations   []DrawingInspirationRecord `json:"drawingInspirations"`
	Expenses              []ExpenseRecord            `json:"expenses"`
	InventoryTransactions []InventoryTransaction     `json:"inventoryTransactions"`
	Markets               []MarketRecord             `json:"markets"`
	Settings              []AppSetting               `json:"settings"`
	PlannerTemplates      []PlannerTemplate          `json:"plannerTemplates"`
}
