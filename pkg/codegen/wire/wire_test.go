package wire

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestProject creates a minimal project structure in a temp directory
// with the given features. Returns the temp directory path.
func setupTestProject(t *testing.T, features, wiredFeatures []string) string {
	t.Helper()

	dir := t.TempDir()

	// Create go.mod
	writeTestFile(t, dir, "go.mod", "module test-project\n\ngo 1.25.0\n")

	// Create directories for each feature
	for _, f := range features {
		entityName := toPascalCase(f)

		// Create usecase dir with New() constructor
		usecaseDir := filepath.Join("internal", "usecase", f)
		writeTestFile(t, dir, filepath.Join(usecaseDir, f+".go"), "package "+f+"\n\nfunc New() *UseCase { return &UseCase{} }\n")

		// Create handler dir
		handlerDir := filepath.Join("internal", "handlers", "http", "v1", f)
		writeTestFile(t, dir, filepath.Join(handlerDir, "handler.go"), "package "+f+"\n")

		// Create DTO dir
		dtoDir := filepath.Join("internal", "dto", f)
		writeTestFile(t, dir, filepath.Join(dtoDir, "request.go"), "package "+f+"dto\n")

		// Create persistent repo file
		repoFile := filepath.Join("internal", "repo", "persistent", f+".go")
		writeTestFile(t, dir, repoFile, "package persistent\n\nfunc New"+entityName+"Repo() {}\n")

		// Create entity file
		entityFile := filepath.Join("internal", "entity", f+".go")
		writeTestFile(t, dir, entityFile, "package entity\n\ntype "+entityName+" struct{}\n")
	}

	// Build target files with wired features
	writeTestFile(t, dir, "internal/repo/contracts.go", buildRepoContracts(wiredFeatures))
	writeTestFile(t, dir, "internal/usecase/contracts.go", buildUsecaseContracts(wiredFeatures))
	writeTestFile(t, dir, "internal/handlers/http/router.go", buildRouter(wiredFeatures))
	writeTestFile(t, dir, "internal/app/app.go", buildApp(wiredFeatures))

	return dir
}

func writeTestFile(t *testing.T, baseDir, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(baseDir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("creating dir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", relPath, err)
	}
}

func buildRepoContracts(wiredFeatures []string) string {
	var sb strings.Builder
	sb.WriteString(`package repo

import (
	"context"

	"test-project/internal/entity"
)

type (
`)
	for _, f := range wiredFeatures {
		entityName := toPascalCase(f)
		varName := toCamelCase(f)
		sb.WriteString("\t// " + entityName + "Repo defines the " + varName + " repository interface.\n")
		sb.WriteString("\t" + entityName + "Repo interface {\n")
		sb.WriteString("\t\tCreate(ctx context.Context, " + varName + " *entity." + entityName + ") error\n")
		sb.WriteString("\t\tGetByID(ctx context.Context, id uint) (*entity." + entityName + ", error)\n")
		sb.WriteString("\t\tDelete(ctx context.Context, id uint) error\n")
		sb.WriteString("\t}\n\n")
	}
	sb.WriteString(")\n")
	return sb.String()
}

func buildUsecaseContracts(wiredFeatures []string) string {
	var sb strings.Builder
	sb.WriteString("package usecase\n\nimport (\n\t\"context\"\n")
	for _, f := range wiredFeatures {
		sb.WriteString("\t\"test-project/internal/dto/" + f + "\"\n")
	}
	sb.WriteString(")\n\ntype (\n")
	for _, f := range wiredFeatures {
		entityName := toPascalCase(f)
		varName := toCamelCase(f)
		sb.WriteString("\t// " + entityName + " defines the " + varName + " use case interface.\n")
		sb.WriteString("\t" + entityName + " interface {\n")
		sb.WriteString("\t\tCreate(ctx context.Context, req " + f + "dto.CreateRequest) (*" + f + "dto.Response, error)\n")
		sb.WriteString("\t\tDelete(ctx context.Context, id uint) error\n")
		sb.WriteString("\t}\n\n")
	}
	sb.WriteString(")\n")
	return sb.String()
}

func buildRouter(wiredFeatures []string) string {
	var sb strings.Builder
	sb.WriteString("package httphandler\n\nimport (\n")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		sb.WriteString("\t" + varName + "handler \"test-project/internal/handlers/http/v1/" + f + "\"\n")
	}
	sb.WriteString("\t\"test-project/internal/usecase\"\n")
	sb.WriteString("\t\"test-project/pkg/logger\"\n")
	sb.WriteString(")\n\n")

	// SetupRoutes signature
	params := make([]string, 0)
	params = append(params, "app *struct{}")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		entityName := toPascalCase(f)
		params = append(params, varName+"UC usecase."+entityName)
	}
	params = append(params, "jwtService struct{}", "l logger.Interface", "healthChecker struct{}")
	sb.WriteString("func SetupRoutes(" + strings.Join(params, ", ") + ") {\n")
	// Call setupAPIRoutes
	callArgs := make([]string, 0)
	callArgs = append(callArgs, "app")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		callArgs = append(callArgs, varName+"UC")
	}
	callArgs = append(callArgs, "jwtService", "l")
	sb.WriteString("\tsetupAPIRoutes(" + strings.Join(callArgs, ", ") + ")\n")
	sb.WriteString("}\n\n")

	// setupAPIRoutes signature
	apiParams := make([]string, 0)
	apiParams = append(apiParams, "app *struct{}")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		entityName := toPascalCase(f)
		apiParams = append(apiParams, varName+"UC usecase."+entityName)
	}
	apiParams = append(apiParams, "jwtService struct{}", "l logger.Interface")
	sb.WriteString("func setupAPIRoutes(" + strings.Join(apiParams, ", ") + ") {\n")
	sb.WriteString("\tapiV1Group := app\n\n")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		sb.WriteString("\t" + varName + "Handler := " + varName + "handler.New(" + varName + "UC, l)\n")
		sb.WriteString("\t" + varName + "Handler.RegisterRoutes(apiV1Group)\n\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

func buildApp(wiredFeatures []string) string {
	var sb strings.Builder
	sb.WriteString("package app\n\nimport (\n")
	sb.WriteString("\t\"test-project/internal/entity\"\n")
	sb.WriteString("\thttphandler \"test-project/internal/handlers/http\"\n")
	sb.WriteString("\t\"test-project/internal/repo\"\n")
	sb.WriteString("\t\"test-project/internal/repo/persistent\"\n")
	sb.WriteString("\t\"test-project/internal/usecase\"\n")
	for _, f := range wiredFeatures {
		sb.WriteString("\t\"test-project/internal/usecase/" + f + "\"\n")
	}
	sb.WriteString("\t\"test-project/pkg/logger\"\n")
	sb.WriteString(")\n\n")

	// repositories struct
	sb.WriteString("type repositories struct {\n")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		entityName := toPascalCase(f)
		sb.WriteString("\t" + varName + " repo." + entityName + "Repo\n")
	}
	sb.WriteString("}\n\n")

	// usecases struct
	sb.WriteString("type usecases struct {\n")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		entityName := toPascalCase(f)
		sb.WriteString("\t" + varName + " usecase." + entityName + "\n")
	}
	sb.WriteString("}\n\n")

	// initRepositories
	sb.WriteString("func initRepositories(db *struct{}) *repositories {\n")
	sb.WriteString("\treturn &repositories{\n")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		entityName := toPascalCase(f)
		sb.WriteString("\t\t" + varName + ": persistent.New" + entityName + "Repo(db),\n")
	}
	sb.WriteString("\t}\n}\n\n")

	// initUseCases
	sb.WriteString("func initUseCases(repos *repositories) *usecases {\n")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		sb.WriteString("\t" + varName + "UC := " + f + ".New(repos." + varName + ")\n")
	}
	sb.WriteString("\n\treturn &usecases{\n")
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		sb.WriteString("\t\t" + varName + ": " + varName + "UC,\n")
	}
	sb.WriteString("\t}\n}\n\n")

	// initHTTPServer with SetupRoutes call
	sb.WriteString("func initHTTPServer(uc *usecases, l *logger.Logger) {\n")
	callArgs := []string{"nil", "nil"}
	for _, f := range wiredFeatures {
		varName := toCamelCase(f)
		callArgs = append(callArgs, "uc."+varName)
	}
	callArgs = append(callArgs, "jwtService", "l", "nil")
	sb.WriteString("\thttphandler.SetupRoutes(" + strings.Join(callArgs, ", ") + ")\n")
	sb.WriteString("}\n\n")

	// runAutoMigrate with AutoMigrate call
	sb.WriteString("func runAutoMigrate(db *struct{}) {\n")
	sb.WriteString("\tdb.AutoMigrate(\n")
	for _, f := range wiredFeatures {
		entityName := toPascalCase(f)
		sb.WriteString("\t\t&entity." + entityName + "{},\n")
	}
	sb.WriteString("\t)\n}\n")

	return sb.String()
}

func TestScanFeatures(t *testing.T) {
	dir := setupTestProject(t, []string{"product", "category"}, nil)

	features, err := scanFeatures(dir)
	if err != nil {
		t.Fatalf("scanFeatures failed: %v", err)
	}

	if len(features) != 2 {
		t.Fatalf("expected 2 features, got %d", len(features))
	}

	// Features should be sorted alphabetically (from ReadDir)
	if features[0].Name != "category" {
		t.Errorf("expected first feature 'category', got '%s'", features[0].Name)
	}
	if features[0].EntityName != "Category" {
		t.Errorf("expected entity name 'Category', got '%s'", features[0].EntityName)
	}
	if features[1].Name != "product" {
		t.Errorf("expected second feature 'product', got '%s'", features[1].Name)
	}
	if features[1].EntityName != "Product" {
		t.Errorf("expected entity name 'Product', got '%s'", features[1].EntityName)
	}
}

func TestScanFeatures_SkipsWithoutConstructor(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)

	// Create a usecase dir without New() constructor
	noConstructorDir := filepath.Join(dir, "internal", "usecase", "orphan")
	writeTestFile(t, dir, filepath.Join("internal", "usecase", "orphan", "orphan.go"), "package orphan\n\nfunc DoSomething() {}\n")
	_ = noConstructorDir

	features, err := scanFeatures(dir)
	if err != nil {
		t.Fatalf("scanFeatures failed: %v", err)
	}

	if len(features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(features))
	}
	if features[0].Name != "product" {
		t.Errorf("expected 'product', got '%s'", features[0].Name)
	}
}

func TestScanFeatures_SkipsWithoutArtifacts(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)

	// Create a usecase with New() but without handler/dto/repo (like auth)
	writeTestFile(t, dir, filepath.Join("internal", "usecase", "auth", "auth.go"),
		"package auth\n\nfunc New() *UseCase { return &UseCase{} }\n")

	features, err := scanFeatures(dir)
	if err != nil {
		t.Fatalf("scanFeatures failed: %v", err)
	}

	if len(features) != 1 {
		t.Fatalf("expected 1 feature (auth should be skipped), got %d", len(features))
	}
	if features[0].Name != "product" {
		t.Errorf("expected 'product', got '%s'", features[0].Name)
	}
}

func TestFindUnwired_AllUnwired(t *testing.T) {
	dir := setupTestProject(t, []string{"product", "category"}, nil)

	features, _ := scanFeatures(dir)
	unwired, err := findUnwired(dir, features)
	if err != nil {
		t.Fatalf("findUnwired failed: %v", err)
	}

	if len(unwired) != 2 {
		t.Fatalf("expected 2 unwired, got %d", len(unwired))
	}
}

func TestFindUnwired_AllWired(t *testing.T) {
	wired := []string{"product", "category"}
	dir := setupTestProject(t, wired, wired)

	features, _ := scanFeatures(dir)
	unwired, err := findUnwired(dir, features)
	if err != nil {
		t.Fatalf("findUnwired failed: %v", err)
	}

	if len(unwired) != 0 {
		t.Fatalf("expected 0 unwired, got %d: %v", len(unwired), featureNames(unwired))
	}
}

func TestFindUnwired_PartiallyWired(t *testing.T) {
	dir := setupTestProject(t, []string{"product", "category"}, []string{"product"})

	features, _ := scanFeatures(dir)
	unwired, err := findUnwired(dir, features)
	if err != nil {
		t.Fatalf("findUnwired failed: %v", err)
	}

	if len(unwired) != 1 {
		t.Fatalf("expected 1 unwired, got %d", len(unwired))
	}
	if unwired[0].Name != "category" {
		t.Errorf("expected 'category' to be unwired, got '%s'", unwired[0].Name)
	}
}

func TestWireRepoContract(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)

	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	f := Feature{Name: "product", EntityName: "Product", PackageName: "product", VarName: "product"}

	if err := wireRepoContract(cfg, f); err != nil {
		t.Fatalf("wireRepoContract failed: %v", err)
	}

	content, _ := readFileContent(dir, "internal/repo/contracts.go")

	if !strings.Contains(content, "ProductRepo interface") {
		t.Error("expected ProductRepo interface in contracts")
	}
	if !strings.Contains(content, "Create(ctx context.Context, product *entity.Product) error") {
		t.Error("expected Create method in ProductRepo")
	}
	if !strings.Contains(content, "GetByID(ctx context.Context, id uint) (*entity.Product, error)") {
		t.Error("expected GetByID method in ProductRepo")
	}
	if !strings.Contains(content, "productdto.ListRequest") {
		t.Error("expected ListRequest with productdto alias")
	}
}

func TestWireUsecaseContract(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)

	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	f := Feature{Name: "product", EntityName: "Product", PackageName: "product", VarName: "product"}

	if err := wireUsecaseContract(cfg, f); err != nil {
		t.Fatalf("wireUsecaseContract failed: %v", err)
	}

	content, _ := readFileContent(dir, "internal/usecase/contracts.go")

	if !strings.Contains(content, "Product interface") {
		t.Error("expected Product interface in contracts")
	}
	if !strings.Contains(content, "productdto.CreateRequest") {
		t.Error("expected CreateRequest with productdto alias")
	}
	if !strings.Contains(content, "productdto.ListResponse") {
		t.Error("expected ListResponse with productdto alias")
	}
}

func TestWireRouter(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)

	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	f := Feature{Name: "product", EntityName: "Product", PackageName: "product", VarName: "product"}

	if err := wireRouter(cfg, f); err != nil {
		t.Fatalf("wireRouter failed: %v", err)
	}

	content, _ := readFileContent(dir, "internal/handlers/http/router.go")

	if !strings.Contains(content, `producthandler "test-project/internal/handlers/http/v1/product"`) {
		t.Error("expected producthandler import")
	}
	if !strings.Contains(content, "productUC usecase.Product") {
		t.Error("expected productUC parameter in function signatures")
	}
	if !strings.Contains(content, "productHandler := producthandler.New(productUC, l)") {
		t.Error("expected handler creation code")
	}
	if !strings.Contains(content, "productHandler.RegisterRoutes(apiV1Group)") {
		t.Error("expected route registration code")
	}
}

func TestWireApp(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)

	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	f := Feature{Name: "product", EntityName: "Product", PackageName: "product", VarName: "product"}

	if err := wireApp(cfg, f); err != nil {
		t.Fatalf("wireApp failed: %v", err)
	}

	content, _ := readFileContent(dir, "internal/app/app.go")

	if !strings.Contains(content, `"test-project/internal/usecase/product"`) {
		t.Error("expected product usecase import")
	}
	if !strings.Contains(content, "product repo.ProductRepo") {
		t.Error("expected product field in repositories struct")
	}
	if !strings.Contains(content, "product usecase.Product") {
		t.Error("expected product field in usecases struct")
	}
	if !strings.Contains(content, "product: persistent.NewProductRepo(db),") {
		t.Error("expected product init in initRepositories")
	}
	if !strings.Contains(content, "productUC := product.New(repos.product)") {
		t.Error("expected product UC init in initUseCases")
	}
	if !strings.Contains(content, "product: productUC,") {
		t.Error("expected product field in usecases return struct")
	}
	if !strings.Contains(content, "uc.product") {
		t.Error("expected uc.product in SetupRoutes call")
	}
	if !strings.Contains(content, "&entity.Product{}") {
		t.Error("expected &entity.Product{} in AutoMigrate")
	}
}

func TestIdempotency_RepoContract(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)
	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	f := Feature{Name: "product", EntityName: "Product", PackageName: "product", VarName: "product"}

	// Wire once
	if err := wireRepoContract(cfg, f); err != nil {
		t.Fatalf("first wireRepoContract failed: %v", err)
	}
	contentAfterFirst, _ := readFileContent(dir, "internal/repo/contracts.go")

	// Wire again - should be idempotent
	if err := wireRepoContract(cfg, f); err != nil {
		t.Fatalf("second wireRepoContract failed: %v", err)
	}
	contentAfterSecond, _ := readFileContent(dir, "internal/repo/contracts.go")

	if contentAfterFirst != contentAfterSecond {
		t.Error("repo contract wiring is not idempotent - content changed on second run")
	}
}

func TestIdempotency_UsecaseContract(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)
	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	f := Feature{Name: "product", EntityName: "Product", PackageName: "product", VarName: "product"}

	if err := wireUsecaseContract(cfg, f); err != nil {
		t.Fatalf("first wireUsecaseContract failed: %v", err)
	}
	contentAfterFirst, _ := readFileContent(dir, "internal/usecase/contracts.go")

	if err := wireUsecaseContract(cfg, f); err != nil {
		t.Fatalf("second wireUsecaseContract failed: %v", err)
	}
	contentAfterSecond, _ := readFileContent(dir, "internal/usecase/contracts.go")

	if contentAfterFirst != contentAfterSecond {
		t.Error("usecase contract wiring is not idempotent")
	}
}

func TestIdempotency_Router(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)
	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	f := Feature{Name: "product", EntityName: "Product", PackageName: "product", VarName: "product"}

	if err := wireRouter(cfg, f); err != nil {
		t.Fatalf("first wireRouter failed: %v", err)
	}
	contentAfterFirst, _ := readFileContent(dir, "internal/handlers/http/router.go")

	if err := wireRouter(cfg, f); err != nil {
		t.Fatalf("second wireRouter failed: %v", err)
	}
	contentAfterSecond, _ := readFileContent(dir, "internal/handlers/http/router.go")

	if contentAfterFirst != contentAfterSecond {
		t.Error("router wiring is not idempotent")
	}
}

func TestIdempotency_App(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)
	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	f := Feature{Name: "product", EntityName: "Product", PackageName: "product", VarName: "product"}

	if err := wireApp(cfg, f); err != nil {
		t.Fatalf("first wireApp failed: %v", err)
	}
	contentAfterFirst, _ := readFileContent(dir, "internal/app/app.go")

	if err := wireApp(cfg, f); err != nil {
		t.Fatalf("second wireApp failed: %v", err)
	}
	contentAfterSecond, _ := readFileContent(dir, "internal/app/app.go")

	if contentAfterFirst != contentAfterSecond {
		t.Error("app wiring is not idempotent")
	}
}

func TestFullWiringRun(t *testing.T) {
	dir := setupTestProject(t, []string{"product", "category"}, []string{"product"})

	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	w := New(cfg)

	if err := w.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify category was wired
	repoContent, _ := readFileContent(dir, "internal/repo/contracts.go")
	if !strings.Contains(repoContent, "CategoryRepo interface") {
		t.Error("expected CategoryRepo in repo contracts after wiring")
	}

	ucContent, _ := readFileContent(dir, "internal/usecase/contracts.go")
	if !strings.Contains(ucContent, "Category interface") {
		t.Error("expected Category in usecase contracts after wiring")
	}

	routerContent, _ := readFileContent(dir, "internal/handlers/http/router.go")
	if !strings.Contains(routerContent, `v1/category"`) {
		t.Error("expected category handler import in router after wiring")
	}

	appContent, _ := readFileContent(dir, "internal/app/app.go")
	if !strings.Contains(appContent, "repo.CategoryRepo") {
		t.Error("expected CategoryRepo in app.go after wiring")
	}

	// Verify product is still intact
	if !strings.Contains(repoContent, "ProductRepo interface") {
		t.Error("expected ProductRepo to still exist in repo contracts")
	}
}

func TestFullWiringRun_Idempotent(t *testing.T) {
	dir := setupTestProject(t, []string{"product"}, nil)

	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	w := New(cfg)

	// First run
	if err := w.Run(); err != nil {
		t.Fatalf("first Run failed: %v", err)
	}

	// Capture all file contents
	repo1, _ := readFileContent(dir, "internal/repo/contracts.go")
	uc1, _ := readFileContent(dir, "internal/usecase/contracts.go")
	router1, _ := readFileContent(dir, "internal/handlers/http/router.go")
	app1, _ := readFileContent(dir, "internal/app/app.go")

	// Second run
	if err := w.Run(); err != nil {
		t.Fatalf("second Run failed: %v", err)
	}

	repo2, _ := readFileContent(dir, "internal/repo/contracts.go")
	uc2, _ := readFileContent(dir, "internal/usecase/contracts.go")
	router2, _ := readFileContent(dir, "internal/handlers/http/router.go")
	app2, _ := readFileContent(dir, "internal/app/app.go")

	if repo1 != repo2 {
		t.Error("repo contracts changed on second run")
	}
	if uc1 != uc2 {
		t.Error("usecase contracts changed on second run")
	}
	if router1 != router2 {
		t.Error("router changed on second run")
	}
	if app1 != app2 {
		t.Error("app.go changed on second run")
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"article", "Article"},
		{"user_role", "UserRole"},
		{"", ""},
		{"a", "A"},
		{"already_pascal", "AlreadyPascal"},
	}

	for _, tt := range tests {
		result := toPascalCase(tt.input)
		if result != tt.expected {
			t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"article", "article"},
		{"user_role", "userRole"},
		{"", ""},
		{"A", "a"},
	}

	for _, tt := range tests {
		result := toCamelCase(tt.input)
		if result != tt.expected {
			t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSnakeCaseFeature(t *testing.T) {
	dir := t.TempDir()

	// Setup a snake_case feature
	feature := "user_role"

	writeTestFile(t, dir, "go.mod", "module test-project\n\ngo 1.25.0\n")
	writeTestFile(t, dir, filepath.Join("internal", "usecase", feature, feature+".go"),
		"package user_role\n\nfunc New() *UseCase { return &UseCase{} }\n")
	writeTestFile(t, dir, filepath.Join("internal", "handlers", "http", "v1", feature, "handler.go"),
		"package user_role\n")
	writeTestFile(t, dir, filepath.Join("internal", "dto", feature, "request.go"),
		"package user_roledto\n")
	writeTestFile(t, dir, filepath.Join("internal", "repo", "persistent", feature+".go"),
		"package persistent\n\nfunc NewUserRoleRepo() {}\n")
	writeTestFile(t, dir, filepath.Join("internal", "entity", feature+".go"),
		"package entity\n\ntype UserRole struct{}\n")
	writeTestFile(t, dir, "internal/repo/contracts.go", buildRepoContracts(nil))
	writeTestFile(t, dir, "internal/usecase/contracts.go", buildUsecaseContracts(nil))
	writeTestFile(t, dir, "internal/handlers/http/router.go", buildRouter(nil))
	writeTestFile(t, dir, "internal/app/app.go", buildApp(nil))

	features, err := scanFeatures(dir)
	if err != nil {
		t.Fatalf("scanFeatures failed: %v", err)
	}

	if len(features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(features))
	}

	f := features[0]
	if f.EntityName != "UserRole" {
		t.Errorf("expected entity name 'UserRole', got '%s'", f.EntityName)
	}
	if f.VarName != "userRole" {
		t.Errorf("expected var name 'userRole', got '%s'", f.VarName)
	}

	// Wire it
	cfg := Config{ModuleName: "test-project", OutputDir: dir}
	if err := wireRepoContract(cfg, f); err != nil {
		t.Fatalf("wireRepoContract failed: %v", err)
	}

	content, _ := readFileContent(dir, "internal/repo/contracts.go")
	if !strings.Contains(content, "UserRoleRepo interface") {
		t.Error("expected UserRoleRepo interface")
	}
}
