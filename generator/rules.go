package generator

import (
	"github.com/riftzen-bit/context-gen/analyzer"
)

// rulesMap maps language/framework names to practical coding rules for AI assistants.
// Keys must match exact values from analyzer.ProjectInfo.Languages[].Name and ProjectInfo.Frameworks.
var rulesMap = map[string][]string{
	// Languages
	"Go": {
		"Use MixedCaps for names, not underscores. Short variable names in small scopes. Acronyms all-caps (HTTP, URL, ID)",
		"Use Result error wrapping with fmt.Errorf and %w. Don't panic for expected errors",
		"Accept interfaces, return structs. Define small, focused interfaces (1-3 methods)",
		"Pass context.Context as the first parameter for cancellation and timeouts",
	},
	"Rust": {
		"Use Result<T, E> for all fallible operations. Define custom error types with thiserror. No unwrap() in production",
		"Use serde derive macros for serialization. Use #[serde(rename_all = \"camelCase\")] for API responses",
		"Use async with tokio for I/O-bound operations",
	},
	"Python": {
		"Use type hints on all function signatures. Use Pydantic BaseModel for validation",
		"Use snake_case for functions and variables, PascalCase for classes",
	},
	"TypeScript": {
		"Use strict TypeScript — no `any` types, use `unknown` and narrow with type guards",
		"Prefer interfaces for object shapes, type aliases for unions/intersections",
		"Use `as const` objects instead of enums for better tree-shaking",
	},

	// Frameworks
	"React": {
		"Use functional components with hooks, no class components",
		"Memoize expensive renders with React.memo, useMemo, useCallback",
		"Use stable, unique keys for list items — never array index as key if items reorder",
		"Prefer composition over inheritance",
		"Extract reusable stateful logic into custom hooks (useXxx)",
	},
	"Tauri": {
		"Use #[tauri::command] for IPC — keep commands thin, delegate to library functions, return Result<T, String>",
		"Use tauri::State<T> for shared state with Mutex/RwLock for thread safety",
		"Use @tauri-apps/api for frontend IPC, not raw window.__TAURI__",
		"Configure CSP and capability permissions in tauri.conf.json",
	},
	"Vue.js": {
		"Use Composition API (setup script) over Options API",
		"Use ref/reactive for state, computed for derived state",
		"Use defineProps/defineEmits for component interface",
	},
	"Next.js": {
		"Use App Router conventions: page.tsx, layout.tsx, loading.tsx, error.tsx",
		"Use Server Components by default, add 'use client' only when needed",
		"Use Server Actions for mutations instead of API routes",
	},
	"Svelte": {
		"Use $state, $derived, $effect runes (Svelte 5)",
		"Use SvelteKit load functions for data fetching",
		"Use form actions for mutations",
	},
	"Angular": {
		"Use standalone components, no NgModules",
		"Use signals for reactive state",
		"Use inject() instead of constructor injection",
	},
	"Tailwind CSS": {
		"Use Tailwind utility classes, avoid custom CSS",
		"Use @apply only in component libraries, not in application code",
	},
	"Express": {
		"Use middleware for cross-cutting concerns (auth, logging, validation)",
		"Validate request bodies at the boundary",
	},
	"Fastify": {
		"Use middleware for cross-cutting concerns (auth, logging, validation)",
		"Validate request bodies at the boundary",
	},
	"Hono": {
		"Use middleware for cross-cutting concerns (auth, logging, validation)",
		"Validate request bodies at the boundary",
	},

	// Additional languages
	"Java": {
		"Use records for data carriers, sealed classes for restricted hierarchies",
		"Use Optional for nullable return values — never return null from public methods",
		"Prefer streams and lambdas over imperative loops for collection processing",
		"Use try-with-resources for all AutoCloseable resources",
	},
	"Kotlin": {
		"Use data classes for DTOs, sealed classes for algebraic types",
		"Use coroutines (suspend functions) for async operations, not callbacks",
		"Leverage null safety — avoid !! operator, use safe calls (?.) and elvis (?:)",
		"Use extension functions to add behavior without inheritance",
	},
	"C#": {
		"Use records for immutable data, init-only setters for controlled mutation",
		"Use pattern matching (switch expressions) for type-based dispatch",
		"Use async/await for I/O operations, avoid blocking calls (.Result, .Wait())",
		"Enable nullable reference types and handle null explicitly",
	},
	"Ruby": {
		"Follow RuboCop conventions for style consistency",
		"Use blocks, procs, and lambdas for functional patterns",
		"Prefer symbols over strings for hash keys and identifiers",
		"Use modules for mixins and namespace organization",
	},
	"PHP": {
		"Use strict types (declare(strict_types=1)) in all files",
		"Use typed properties and constructor promotion (PHP 8+)",
		"Use match expressions instead of switch for value mapping",
		"Use enums (PHP 8.1+) instead of class constants for fixed sets",
	},
	"Swift": {
		"Use structs over classes by default — classes only when identity or inheritance is needed",
		"Use guard for early returns to reduce nesting",
		"Use Codable for JSON serialization/deserialization",
		"Use async/await for concurrency instead of completion handlers",
	},
	"Dart": {
		"Use null safety — never use late unless absolutely necessary",
		"Use const constructors for immutable widgets to optimize rebuilds",
		"Use named parameters for improved readability in function calls",
		"Use freezed or built_value for immutable data models",
	},
	"Elixir": {
		"Use pattern matching in function heads for control flow",
		"Use pipes (|>) for data transformation chains",
		"Use GenServer for stateful processes, Tasks for one-off async work",
		"Use with for chaining operations that may fail",
	},
	"Lua": {
		"Use local variables — global lookups are slower",
		"Use metatables and __index for OOP patterns",
		"Use tables as the primary data structure for both arrays and maps",
	},

	// Additional frameworks
	"Django": {
		"Use class-based views for standard CRUD, function views for custom logic",
		"Use model managers and querysets — keep business logic out of views",
		"Use Django REST Framework serializers for API validation",
		"Use signals sparingly — prefer explicit method calls for clarity",
	},
	"Flask": {
		"Use application factory pattern with create_app()",
		"Use blueprints to organize routes by feature/domain",
		"Use Flask-SQLAlchemy for database with migration support via Alembic",
	},
	"FastAPI": {
		"Use Pydantic models for request/response validation",
		"Use dependency injection for shared resources (DB sessions, auth)",
		"Use async def for I/O-bound endpoints, def for CPU-bound",
		"Use APIRouter to organize endpoints by domain",
	},
	"Nuxt": {
		"Use auto-imports for composables and components",
		"Use useFetch/useAsyncData for data fetching with SSR support",
		"Use server routes (server/api/) for backend logic",
	},
	"Remix": {
		"Use loader functions for data fetching, action functions for mutations",
		"Use Form component for progressive enhancement",
		"Use nested routes to colocate data requirements with UI",
	},
	"Astro": {
		"Use .astro components for static content, framework components for interactivity",
		"Use content collections for structured content with type safety",
		"Minimize client-side JavaScript — use client:* directives only when needed",
	},
	"Electron": {
		"Separate main process (Node.js) and renderer process (browser) concerns",
		"Use contextBridge and preload scripts for secure IPC — never expose Node APIs directly",
		"Use IPC channels with strict message validation between processes",
	},
	"Gatsby": {
		"Use GraphQL data layer for content sourcing",
		"Use gatsby-node.js for dynamic page creation",
		"Use gatsby-image/gatsby-plugin-image for optimized images",
	},
	"styled-components": {
		"Use styled() for extending existing components",
		"Use ThemeProvider for consistent design tokens",
		"Use css helper for shared style fragments",
	},
	"Emotion": {
		"Use css prop or styled API consistently — don't mix approaches",
		"Use theme via useTheme hook for design tokens",
	},
	"Sass": {
		"Use variables, mixins, and functions for reusable styles",
		"Use nesting sparingly — max 3 levels deep",
		"Use partials (_filename.scss) and @use instead of @import",
	},
}

// generalRules are always included regardless of detected tech stack.
var generalRules = []string{
	"No hardcoded secrets — use environment variables or config files",
	"Document complex logic — add comments explaining WHY, not WHAT",
	"Write clear, descriptive variable and function names",
}

// RelevantRules returns coding rules matching the detected languages and frameworks.
func RelevantRules(info *analyzer.ProjectInfo) []string {
	seen := make(map[string]bool)
	var rules []string

	addRules := func(items []string) {
		for _, r := range items {
			if !seen[r] {
				seen[r] = true
				rules = append(rules, r)
			}
		}
	}

	for _, lang := range info.Languages {
		if r, ok := rulesMap[lang.Name]; ok {
			addRules(r)
		}
	}

	for _, fw := range info.Frameworks {
		if r, ok := rulesMap[fw]; ok {
			addRules(r)
		}
	}

	addRules(generalRules)

	return rules
}
