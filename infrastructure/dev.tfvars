environment = "dev"

min_instances = 0
max_instances = 1

# Resources
cpu_limit    = "1"
memory_limit = "512Mi"

# Timeouts
read_timeout  = "15"
write_timeout = "15"
idle_timeout  = "60"

allow_public_access = true
allowed_origins     = ["http://localhost:5173", "https://localhost:5173"]
cookie_domain       = ""

# Logging
log_levels = "debug"
log_format = "text"