package compiler

import (
	"errors"
	"testing"

	"github.com/sqlc-dev/sqlc/internal/sql/ast"
	"github.com/sqlc-dev/sqlc/internal/sql/sqlerr"
)

func TestParseExcludeIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		excl      string
		want      ExcludeIdentifier
		needCheck bool
	}{
		{
			name:      "unqualified",
			excl:      "user_id",
			want:      ExcludeIdentifier{ColName: "user_id", Original: "user_id"},
			needCheck: true,
		},
		{
			name:      "partially qualified",
			excl:      "posts.user_id",
			want:      ExcludeIdentifier{Table: "posts", ColName: "user_id", Original: "posts.user_id"},
			needCheck: true,
		},
		{
			name:      "fully qualified",
			excl:      "public.posts.user_id",
			want:      ExcludeIdentifier{Schema: "public", Table: "posts", ColName: "user_id", Original: "public.posts.user_id"},
			needCheck: false,
		},
		{
			name:      "invalid - too many parts",
			excl:      "a.b.c.d",
			want:      ExcludeIdentifier{},
			needCheck: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, needCheck := parseExcludeIdentifier(tt.excl)
			if needCheck != tt.needCheck {
				t.Errorf("needCheck = %v, want %v", needCheck, tt.needCheck)
			}
			if got != tt.want {
				t.Errorf("parseExcludeIdentifier() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestValidateExcludeAmbiguity(t *testing.T) {
	usersTable := &Table{
		Rel: &ast.TableName{Name: "users"},
		Columns: []*Column{
			{Name: "id"},
			{Name: "name"},
		},
	}
	postsTable := &Table{
		Rel: &ast.TableName{Name: "posts"},
		Columns: []*Column{
			{Name: "id"},
			{Name: "user_id"},
		},
	}
	commentsTable := &Table{
		Rel: &ast.TableName{Name: "comments"},
		Columns: []*Column{
			{Name: "id"},
			{Name: "user_id"},
		},
	}

	tests := []struct {
		name     string
		excludes []string
		tables   []*Table
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "unqualified exclude matches two tables - ambiguous",
			excludes: []string{"user_id"},
			tables:   []*Table{usersTable, postsTable, commentsTable},
			wantErr:  true,
			errMsg:   "user_id",
		},
		{
			name:     "unqualified exclude matches one table - ok",
			excludes: []string{"password"},
			tables: []*Table{
				{Rel: &ast.TableName{Name: "users"}, Columns: []*Column{{Name: "id"}, {Name: "password"}}},
			},
			wantErr: false,
		},
		{
			name:     "qualified exclude - ok",
			excludes: []string{"posts.user_id", "comments.user_id"},
			tables:   []*Table{usersTable, postsTable, commentsTable},
			wantErr:  false,
		},
		{
			name:     "no excludes - ok",
			excludes: nil,
			tables:   []*Table{usersTable, postsTable},
			wantErr:  false,
		},
		{
			name:     "fully qualified exclude - ok",
			excludes: []string{"public.posts.user_id"},
			tables:   []*Table{postsTable, commentsTable},
			wantErr:  false,
		},
		{
			name:     "partially qualified exclude matches two schemas - ambiguous",
			excludes: []string{"posts.user_id"},
			tables: []*Table{
				{Rel: &ast.TableName{Schema: "public", Name: "posts"}, Columns: []*Column{{Name: "id"}, {Name: "user_id"}}},
				{Rel: &ast.TableName{Schema: "enterprise", Name: "posts"}, Columns: []*Column{{Name: "id"}, {Name: "user_id"}}},
			},
			wantErr: true,
			errMsg:  "posts.user_id",
		},
		{
			name:     "partially qualified exclude matches one schema - ok",
			excludes: []string{"posts.user_id"},
			tables: []*Table{
				{Rel: &ast.TableName{Schema: "public", Name: "posts"}, Columns: []*Column{{Name: "id"}, {Name: "user_id"}}},
			},
			wantErr: false,
		},
		{
			name:     "partially qualified exclude with empty schema - counts as distinct",
			excludes: []string{"posts.user_id"},
			tables: []*Table{
				{Rel: &ast.TableName{Schema: "", Name: "posts"}, Columns: []*Column{{Name: "id"}, {Name: "user_id"}}},
				{Rel: &ast.TableName{Schema: "enterprise", Name: "posts"}, Columns: []*Column{{Name: "id"}, {Name: "user_id"}}},
			},
			wantErr: true,
			errMsg:  "posts.user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExcludeAmbiguity(tt.excludes, tt.tables)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var serr *sqlerr.Error
				if !errors.As(err, &serr) {
					t.Fatalf("expected *sqlerr.Error, got %T", err)
				}
				if serr.Message != `"`+tt.errMsg+`" is ambiguous` {
					t.Errorf("got message %q", serr.Message)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
