package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"seasonschedule/internal/models"
)

// Configuration for authentication
const (
	JWTSecret = "your-secret-key-change-in-production"
)

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its bcrypt hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken creates a JWT token for the given user
func GenerateToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &models.Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	return tokenString, err
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*models.Claims, error) {
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// GetUserByUsername loads a user record from the database.
func GetUserByUsername(ctx context.Context, db *sql.DB, username string) (models.User, error) {
	query := `
		SELECT id, username, COALESCE(email, ''), is_admin, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user models.User
	err := db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// ValidateCredentials checks credentials against the users table in the database.
func ValidateCredentials(ctx context.Context, db *sql.DB, username, password string) (models.User, bool) {
	if db == nil {
		log.Println("Database connection not initialized")
		return models.User{}, false
	}

	query := `
		SELECT id, username, password_hash, COALESCE(email, ''), is_admin, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user models.User
	var storedHash string
	err := db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&storedHash,
		&user.Email,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error querying user: %v", err)
		return models.User{}, false
	}

	if !CheckPasswordHash(password, storedHash) {
		return models.User{}, false
	}

	return user, true
}

// CheckPermission checks if a user has a specific permission on a resource
func CheckPermission(ctx context.Context, db *sql.DB, userID uuid.UUID, resourceType string, resourceID uuid.UUID, requiredPermission string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM permissions
			WHERE user_id = $1
			AND resource_type = $2
			AND resource_id = $3
			AND permission IN ($4, 'admin')
		)
	`

	var hasPermission bool
	err := db.QueryRowContext(ctx, query, userID, resourceType, resourceID, requiredPermission).Scan(&hasPermission)
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}

// GetUserPermissions returns all permissions for a user
func GetUserPermissions(ctx context.Context, db *sql.DB, userID uuid.UUID) ([]models.Permission, error) {
	query := `
		SELECT id, user_id, resource_type, resource_id, permission, created_at, updated_at
		FROM permissions
		WHERE user_id = $1
		ORDER BY resource_type, resource_id
	`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(&perm.ID, &perm.UserID, &perm.ResourceType, &perm.ResourceID, &perm.Permission, &perm.CreatedAt, &perm.UpdatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}

// GrantPermission grants a permission to a user on a resource
func GrantPermission(ctx context.Context, db *sql.DB, userID uuid.UUID, resourceType string, resourceID uuid.UUID, permission string) error {
	query := `
		INSERT INTO permissions (user_id, resource_type, resource_id, permission)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, resource_type, resource_id, permission) DO NOTHING
	`

	_, err := db.ExecContext(ctx, query, userID, resourceType, resourceID, permission)
	return err
}

// RevokePermission revokes a permission from a user on a resource
func RevokePermission(ctx context.Context, db *sql.DB, userID uuid.UUID, resourceType string, resourceID uuid.UUID, permission string) error {
	query := `
		DELETE FROM permissions
		WHERE user_id = $1 AND resource_type = $2 AND resource_id = $3 AND permission = $4
	`

	_, err := db.ExecContext(ctx, query, userID, resourceType, resourceID, permission)
	return err
}

// IsSiteAdmin checks if a user is a site-wide administrator
func IsSiteAdmin(isAdmin bool) bool {
	return isAdmin
}

// IsOrganizationAdmin checks if a user has admin permission on a specific organization
func IsOrganizationAdmin(ctx context.Context, db *sql.DB, userID uuid.UUID, orgID uuid.UUID) (bool, error) {
	return CheckPermission(ctx, db, userID, "organization", orgID, "admin")
}

// IsTeamAdmin checks if a user has admin permission on a specific team
func IsTeamAdmin(ctx context.Context, db *sql.DB, userID uuid.UUID, teamID uuid.UUID) (bool, error) {
	return CheckPermission(ctx, db, userID, "team", teamID, "admin")
}

// GetTeamByID retrieves team information by ID
func GetTeamByID(ctx context.Context, db *sql.DB, teamID uuid.UUID) (models.Team, error) {
	query := `SELECT id, organization_id, name, created_at, updated_at FROM teams WHERE id = $1`
	var team models.Team
	err := db.QueryRowContext(ctx, query, teamID).Scan(
		&team.ID, &team.OrganizationID, &team.Name, &team.CreatedAt, &team.UpdatedAt)
	return team, err
}

// GetOrganizationByTeamID retrieves the organization_id for a given team
func GetOrganizationByTeamID(ctx context.Context, db *sql.DB, teamID uuid.UUID) (uuid.UUID, error) {
	team, err := GetTeamByID(ctx, db, teamID)
	if err != nil {
		return uuid.Nil, err
	}
	return team.OrganizationID, nil
}

// CanAccessOrganization checks if a user can access an organization (read level)
// Site admin: yes
// Organization admin: yes for that org
// Team admin: yes if they have admin on any team in that org
func CanAccessOrganization(ctx context.Context, db *sql.DB, userID uuid.UUID, orgID uuid.UUID, isSiteAdmin bool) (bool, error) {
	if isSiteAdmin {
		return true, nil
	}

	// Check if organization admin
	if isOrgAdmin, err := IsOrganizationAdmin(ctx, db, userID, orgID); err != nil {
		return false, err
	} else if isOrgAdmin {
		return true, nil
	}

	// Check if team admin in this organization
	query := `
		SELECT EXISTS(
			SELECT 1 FROM permissions p
			JOIN teams t ON p.resource_id = t.id
			WHERE p.user_id = $1
			AND p.resource_type = 'team'
			AND p.permission = 'admin'
			AND t.organization_id = $2
		)
	`
	var hasTeamAdmin bool
	err := db.QueryRowContext(ctx, query, userID, orgID).Scan(&hasTeamAdmin)
	if err != nil {
		return false, err
	}
	
	return hasTeamAdmin, nil
}

// CanManageOrganization checks if a user can manage an organization (write/admin level)
// Site admin: yes
// Organization admin: yes for that org
func CanManageOrganization(ctx context.Context, db *sql.DB, userID uuid.UUID, orgID uuid.UUID, isSiteAdmin bool) (bool, error) {
	if isSiteAdmin {
		return true, nil
	}

	return IsOrganizationAdmin(ctx, db, userID, orgID)
}

// CanManageTeam checks if a user can manage a team (write/admin level)
// Site admin: yes
// Organization admin: yes for teams in their org
// Team admin: yes for that team
func CanManageTeam(ctx context.Context, db *sql.DB, userID uuid.UUID, teamID uuid.UUID, isSiteAdmin bool) (bool, error) {
	if isSiteAdmin {
		return true, nil
	}

	// Check if team admin
	if isTeamAdmin, err := IsTeamAdmin(ctx, db, userID, teamID); err != nil {
		return false, err
	} else if isTeamAdmin {
		return true, nil
	}

	// Get organization ID for this team
	team, err := GetTeamByID(ctx, db, teamID)
	if err != nil {
		return false, err
	}

	// Check if organization admin
	return IsOrganizationAdmin(ctx, db, userID, team.OrganizationID)
}

// CanManagePermissions checks if a user can grant/revoke permissions on a resource
// Site admin: yes for any resource
// Organization admin: yes for resources within their org
// Team admin: yes for resources within their team
func CanManagePermissions(ctx context.Context, db *sql.DB, userID uuid.UUID, resourceType string, resourceID uuid.UUID, isSiteAdmin bool) (bool, error) {
	if isSiteAdmin {
		return true, nil
	}

	switch resourceType {
	case "organization":
		// Only site admins can manage organization permissions
		return false, nil
	case "team":
		// Organization admins can manage team permissions within their org
		team, err := GetTeamByID(ctx, db, resourceID)
		if err != nil {
			return false, err
		}
		return IsOrganizationAdmin(ctx, db, userID, team.OrganizationID)
	default:
		return false, nil
	}
}
