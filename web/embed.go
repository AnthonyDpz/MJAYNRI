// Package web fournit les assets web embarqués dans le binaire Go.
// L'utilisation de //go:embed permet un déploiement monobinaire :
// aucun fichier externe n'est nécessaire en production.
package web

import "embed"

// TemplateFS contient tous les templates HTML (web/templates/*.html).
//
//go:embed templates
var TemplateFS embed.FS

// StaticFS contient les assets statiques (CSS, JS) sous web/static/.
//
//go:embed static
var StaticFS embed.FS
