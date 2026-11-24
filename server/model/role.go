package model

import "time"

// Constantes de roles
const (
	RoleAdmin        = "admin"        // Administrador, tiene todos los permisos
	RoleAuditor      = "auditor"      // Auditor, puede ver logs e información de auditoría
	RoleConfigurator = "configurator" // Administrador de configuración, puede gestionar la configuración del sistema
	RoleUser         = "user"         // Usuario normal, permisos de operación básicos
)

// Constantes de permisos
const (
	// Permisos de gestión de usuarios
	PermUserCreate = "user:create"
	PermUserRead   = "user:read"
	PermUserUpdate = "user:update"
	PermUserDelete = "user:delete"

	// Permisos de gestión de sitios
	PermSiteCreate = "site:create"
	PermSiteRead   = "site:read"
	PermSiteUpdate = "site:update"
	PermSiteDelete = "site:delete"

	// Permisos de gestión de certificados
	PermCertCreate = "cert:create"
	PermCertRead   = "cert:read"
	PermCertUpdate = "cert:update"
	PermCertDelete = "cert:delete"

	// Permisos de gestión de configuración
	PermConfigRead   = "config:read"
	PermConfigUpdate = "config:update"

	// Permisos de logs de auditoría
	PermAuditRead = "audit:read"

	// Permisos de administración del sistema
	PermSystemRestart = "system:restart"
	PermSystemStatus  = "system:status"

	// Permisos de logs del WAF
	PermWAFLogRead = "waf:log:read"
)

// Role modelo de rol
type Role struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string    `bson:"name" json:"name"`               // Nombre del rol
	Description string    `bson:"description" json:"description"` // Descripción del rol
	Permissions []string  `bson:"permissions" json:"permissions"` // Lista de permisos
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

// Obtener el mapeo de permisos de roles por defecto
func GetDefaultRolePermissions() map[string][]string {
	return map[string][]string{
		RoleAdmin: {
			// El administrador tiene todos los permisos
			PermUserCreate, PermUserRead, PermUserUpdate, PermUserDelete,
			PermSiteCreate, PermSiteRead, PermSiteUpdate, PermSiteDelete,
			PermConfigRead, PermConfigUpdate,
			PermAuditRead,
			PermSystemRestart, PermSystemStatus,
			PermWAFLogRead,
			PermCertCreate, PermCertRead, PermCertUpdate, PermCertDelete,
		},
		RoleAuditor: {
			// El auditor puede ver usuarios, sitios, configuración y logs de auditoría
			PermUserRead,
			PermSiteRead,
			PermConfigRead,
			PermAuditRead,
			PermSystemStatus,
			PermWAFLogRead,
			PermCertRead,
		},
		RoleConfigurator: {
			// El administrador de configuración puede gestionar sitios y configuración
			PermSiteCreate, PermSiteRead, PermSiteUpdate, PermSiteDelete,
			PermConfigRead, PermConfigUpdate,
			PermSystemStatus,
			PermWAFLogRead,
			PermCertRead, PermCertUpdate, PermCertDelete,
		},
		RoleUser: {
			// Los usuarios normales solo pueden ver sitios y estado del sistema
			PermSiteRead,
			PermSystemStatus,
			PermWAFLogRead,
			PermCertRead,
		},
	}
}

// GetCollectionName devuelve el nombre de la colección
func (r *Role) GetCollectionName() string {
	return "role"
}
