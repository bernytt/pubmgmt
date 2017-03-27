package pub

type Error string

func (e Error) Error() string { return string(e) }

// General errors.
const (
	ErrUnauthorized         = Error("Unauthorized")
	ErrResourceAccessDenied = Error("Access denied to resource")
	ErrModelNotFound        = Error("Model not found")
	ErrObjNotFound          = Error("Object not found")
	ErrModelSetEmpty        = Error("ModelSet is empty")
	ErrIDCanNotSet          = Error("Id can not set")
)

// User errors.
const (
	ErrUserNotFound            = Error("User not found")
	ErrUsernameAlreadyExists   = Error("Username already exists")
	ErrEmailAlreadyExists      = Error("Email already exists")
	ErrUserInactive            = Error("User is inactive")
	ErrAdminAlreadyInitialized = Error("Admin user already initialized")
)

// Hostgroup errors
const (
	ErrHostgroupSetEmpty      = Error("Not any hostgroups yet")
	ErrHostgroupNotFound      = Error("Hostgroup not found")
	ErrHostgroupAlreadyExists = Error("Hostgroup already exists")
)

// Host errors
const (
	ErrHostSetEmpty      = Error("Not any hosts yet")
	ErrHostNotFound      = Error("Host not found")
	ErrHostAlreadyExists = Error("Host already exists")
	ErrHostInactive      = Error("Host is inactive")
)

// Email errors
const (
	ErrEmailNotFound    = Error("Email not found")
	ErrEmailOutboxEmpty = Error("Not any emails in outbox")
)

// Task errors
const (
	ErrTaskNotFound = Error("Task not found")
	ErrTaskSetEmpty = Error("Not any tasks yet")
	ErrCronNotFound = Error("Cron job not found")
	ErrCronSetEmpty = Error("Not any cron jobs yet")
)

// Modules errors
const (
	ErrSvnInfoSetEmpty = Error("Not any svn infos yet")
	ErrSvnInfoNotFound = Error("Svn info not found")
)

// Crypto errors.
const (
	ErrCryptoHashFailure = Error("Unable to hash data")
)

// JWT errors.
const (
	ErrSecretGeneration   = Error("Unable to generate secret key")
	ErrInvalidJWTToken    = Error("Invalid JWT token")
	ErrMissingContextData = Error("Unable to find JWT data in request context")
)
