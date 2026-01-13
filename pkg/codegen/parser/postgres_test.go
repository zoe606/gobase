package parser

import (
	"testing"
)

func TestParseCreateTable(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		expectedTable string
		expectedCols  int
		wantErr       bool
	}{
		{
			name: "simple table",
			sql: `CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				email VARCHAR(255) NOT NULL UNIQUE,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);`,
			expectedTable: "users",
			expectedCols:  4,
			wantErr:       false,
		},
		{
			name: "table with IF NOT EXISTS",
			sql: `CREATE TABLE IF NOT EXISTS profiles (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL,
				bio TEXT
			);`,
			expectedTable: "profiles",
			expectedCols:  3,
			wantErr:       false,
		},
		{
			name: "table with foreign key",
			sql: `CREATE TABLE articles (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				title VARCHAR(255) NOT NULL
			);`,
			expectedTable: "articles",
			expectedCols:  3,
			wantErr:       false,
		},
		{
			name: "table with JSONB",
			sql: `CREATE TABLE settings (
				id SERIAL PRIMARY KEY,
				config JSONB,
				metadata JSONB DEFAULT '{}'
			);`,
			expectedTable: "settings",
			expectedCols:  3,
			wantErr:       false,
		},
		{
			name:    "no CREATE TABLE statement",
			sql:     `SELECT * FROM users;`,
			wantErr: true,
		},
	}

	parser := NewPostgresParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Table.Name != tt.expectedTable {
				t.Errorf("table name = %q, want %q", result.Table.Name, tt.expectedTable)
			}

			if len(result.Table.Columns) != tt.expectedCols {
				t.Errorf("column count = %d, want %d", len(result.Table.Columns), tt.expectedCols)
			}
		})
	}
}

func TestParseColumn(t *testing.T) {
	parser := NewPostgresParser()

	tests := []struct {
		name           string
		sql            string
		columnName     string
		expectedType   string
		expectedNull   bool
		expectedPK     bool
		expectedUnique bool
		expectedFK     bool
	}{
		{
			name: "serial primary key",
			sql: `CREATE TABLE test (
				id SERIAL PRIMARY KEY
			);`,
			columnName:   "id",
			expectedType: "SERIAL",
			expectedNull: false,
			expectedPK:   true,
		},
		{
			name: "varchar not null",
			sql: `CREATE TABLE test (
				name VARCHAR(255) NOT NULL
			);`,
			columnName:   "name",
			expectedType: "VARCHAR(255)",
			expectedNull: false,
			expectedPK:   false,
		},
		{
			name: "nullable text",
			sql: `CREATE TABLE test (
				bio TEXT
			);`,
			columnName:   "bio",
			expectedType: "TEXT",
			expectedNull: true,
			expectedPK:   false,
		},
		{
			name: "unique column",
			sql: `CREATE TABLE test (
				email VARCHAR(255) NOT NULL UNIQUE
			);`,
			columnName:     "email",
			expectedType:   "VARCHAR(255)",
			expectedNull:   false,
			expectedUnique: true,
		},
		{
			name: "foreign key reference",
			sql: `CREATE TABLE test (
				user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
			);`,
			columnName:   "user_id",
			expectedType: "INTEGER",
			expectedNull: false,
			expectedFK:   true,
		},
		{
			name: "timestamp with time zone",
			sql: `CREATE TABLE test (
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);`,
			columnName:   "created_at",
			expectedType: "TIMESTAMP WITH TIME ZONE",
			expectedNull: true,
		},
		{
			name: "boolean with default",
			sql: `CREATE TABLE test (
				is_active BOOLEAN DEFAULT true
			);`,
			columnName:   "is_active",
			expectedType: "BOOLEAN",
			expectedNull: true,
		},
		{
			name: "jsonb type",
			sql: `CREATE TABLE test (
				data JSONB
			);`,
			columnName:   "data",
			expectedType: "JSONB",
			expectedNull: true,
		},
		{
			name: "bigint type",
			sql: `CREATE TABLE test (
				count BIGINT NOT NULL
			);`,
			columnName:   "count",
			expectedType: "BIGINT",
			expectedNull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.sql)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Table.Columns) == 0 {
				t.Fatal("expected at least one column")
			}

			col := result.Table.Columns[0]

			if col.Name != tt.columnName {
				t.Errorf("column name = %q, want %q", col.Name, tt.columnName)
			}

			if col.SQLType != tt.expectedType {
				t.Errorf("SQL type = %q, want %q", col.SQLType, tt.expectedType)
			}

			if col.IsNullable != tt.expectedNull {
				t.Errorf("nullable = %v, want %v", col.IsNullable, tt.expectedNull)
			}

			if col.IsPrimaryKey != tt.expectedPK {
				t.Errorf("primary key = %v, want %v", col.IsPrimaryKey, tt.expectedPK)
			}

			if col.IsUnique != tt.expectedUnique {
				t.Errorf("unique = %v, want %v", col.IsUnique, tt.expectedUnique)
			}

			if (col.ForeignKey != nil) != tt.expectedFK {
				t.Errorf("has foreign key = %v, want %v", col.ForeignKey != nil, tt.expectedFK)
			}
		})
	}
}

func TestParseIndexes(t *testing.T) {
	parser := NewPostgresParser()

	tests := []struct {
		name           string
		sql            string
		expectedCount  int
		expectedNames  []string
		expectedUnique []bool
	}{
		{
			name: "single index",
			sql: `CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				email VARCHAR(255)
			);
			CREATE INDEX idx_users_email ON users(email);`,
			expectedCount:  1,
			expectedNames:  []string{"idx_users_email"},
			expectedUnique: []bool{false},
		},
		{
			name: "unique index",
			sql: `CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				email VARCHAR(255)
			);
			CREATE UNIQUE INDEX idx_users_email ON users(email);`,
			expectedCount:  1,
			expectedNames:  []string{"idx_users_email"},
			expectedUnique: []bool{true},
		},
		{
			name: "multiple indexes",
			sql: `CREATE TABLE articles (
				id SERIAL PRIMARY KEY,
				user_id INTEGER,
				slug VARCHAR(255),
				status VARCHAR(20)
			);
			CREATE INDEX idx_articles_user_id ON articles(user_id);
			CREATE UNIQUE INDEX idx_articles_slug ON articles(slug);
			CREATE INDEX idx_articles_status ON articles(status);`,
			expectedCount:  3,
			expectedNames:  []string{"idx_articles_user_id", "idx_articles_slug", "idx_articles_status"},
			expectedUnique: []bool{false, true, false},
		},
		{
			name: "no indexes",
			sql: `CREATE TABLE simple (
				id SERIAL PRIMARY KEY
			);`,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.sql)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Table.Indexes) != tt.expectedCount {
				t.Errorf("index count = %d, want %d", len(result.Table.Indexes), tt.expectedCount)
			}

			for i, idx := range result.Table.Indexes {
				if i < len(tt.expectedNames) && idx.Name != tt.expectedNames[i] {
					t.Errorf("index[%d].Name = %q, want %q", i, idx.Name, tt.expectedNames[i])
				}
				if i < len(tt.expectedUnique) && idx.Unique != tt.expectedUnique[i] {
					t.Errorf("index[%d].Unique = %v, want %v", i, idx.Unique, tt.expectedUnique[i])
				}
			}
		})
	}
}

func TestParseTableComment(t *testing.T) {
	parser := NewPostgresParser()

	tests := []struct {
		name            string
		sql             string
		expectedComment string
	}{
		{
			name: "with comment",
			sql: `CREATE TABLE articles (
				id SERIAL PRIMARY KEY
			);
			COMMENT ON TABLE articles IS 'Blog articles and posts';`,
			expectedComment: "Blog articles and posts",
		},
		{
			name: "without comment",
			sql: `CREATE TABLE simple (
				id SERIAL PRIMARY KEY
			);`,
			expectedComment: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.sql)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Table.Comment != tt.expectedComment {
				t.Errorf("comment = %q, want %q", result.Table.Comment, tt.expectedComment)
			}
		})
	}
}

func TestGenerateGoTypes(t *testing.T) {
	parser := NewPostgresParser()

	sql := `CREATE TABLE profiles (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		bio TEXT,
		avatar_url VARCHAR(500),
		settings JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE
	);`

	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check fields were generated
	if len(result.Fields) == 0 {
		t.Fatal("expected Go fields to be generated")
	}

	// Find specific fields and verify their types
	fieldMap := make(map[string]GoField)
	for _, f := range result.Fields {
		fieldMap[f.Name] = f
	}

	// Check ID field
	if id, ok := fieldMap["ID"]; ok {
		if id.Type != "uint" {
			t.Errorf("ID type = %q, want %q", id.Type, "uint")
		}
	} else {
		t.Error("expected ID field")
	}

	// Check UserID (foreign key should be uint)
	if userID, ok := fieldMap["UserID"]; ok {
		if userID.Type != "uint" {
			t.Errorf("UserID type = %q, want %q", userID.Type, "uint")
		}
	} else {
		t.Error("expected UserID field")
	}

	// Check nullable Text field
	if bio, ok := fieldMap["Bio"]; ok {
		if bio.Type != "*string" {
			t.Errorf("Bio type = %q, want %q", bio.Type, "*string")
		}
	} else {
		t.Error("expected Bio field")
	}

	// Check JSONB field
	if settings, ok := fieldMap["Settings"]; ok {
		if settings.Type != "JSONMap" {
			t.Errorf("Settings type = %q, want %q", settings.Type, "JSONMap")
		}
	} else {
		t.Error("expected Settings field")
	}

	// Check deleted_at (should be gorm.DeletedAt)
	if deletedAt, ok := fieldMap["DeletedAt"]; ok {
		if deletedAt.Type != "gorm.DeletedAt" {
			t.Errorf("DeletedAt type = %q, want %q", deletedAt.Type, "gorm.DeletedAt")
		}
	} else {
		t.Error("expected DeletedAt field")
	}

	// Check relations were generated for foreign key
	if len(result.Relations) == 0 {
		t.Error("expected relations for foreign key")
	} else {
		rel := result.Relations[0]
		if rel.Name != "User" {
			t.Errorf("relation name = %q, want %q", rel.Name, "User")
		}
		if rel.Type != "*User" {
			t.Errorf("relation type = %q, want %q", rel.Type, "*User")
		}
	}
}

func TestParseForeignKeyDetails(t *testing.T) {
	parser := NewPostgresParser()

	sql := `CREATE TABLE comments (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE SET NULL,
		article_id INTEGER REFERENCES articles(id) ON DELETE SET NULL
	);`

	result, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find user_id column
	var userIDCol *Column
	var articleIDCol *Column
	for i := range result.Table.Columns {
		if result.Table.Columns[i].Name == "user_id" {
			userIDCol = &result.Table.Columns[i]
		}
		if result.Table.Columns[i].Name == "article_id" {
			articleIDCol = &result.Table.Columns[i]
		}
	}

	if userIDCol == nil {
		t.Fatal("expected user_id column")
	}
	if userIDCol.ForeignKey == nil {
		t.Fatal("expected user_id to have foreign key")
	}

	// Check FK details
	fk := userIDCol.ForeignKey
	if fk.RefTable != "users" {
		t.Errorf("FK ref table = %q, want %q", fk.RefTable, "users")
	}
	if fk.RefColumn != "id" {
		t.Errorf("FK ref column = %q, want %q", fk.RefColumn, "id")
	}
	if fk.OnDelete != "CASCADE" {
		t.Errorf("FK on delete = %q, want %q", fk.OnDelete, "CASCADE")
	}
	if fk.OnUpdate != "SET NULL" {
		t.Errorf("FK on update = %q, want %q", fk.OnUpdate, "SET NULL")
	}

	// Check nullable FK
	if articleIDCol == nil {
		t.Fatal("expected article_id column")
	}
	if !articleIDCol.IsNullable {
		t.Error("expected article_id to be nullable")
	}
}
