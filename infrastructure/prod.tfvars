environment = "prod"

# Scaling
min_instances = 1
max_instances = 2

# Resources
cpu_limit    = "1"
memory_limit = "512Mi"

# Timeouts
read_timeout  = "15"
write_timeout = "15"
idle_timeout  = "60"

# Access & Security
allow_public_access = true
allowed_origins     = ["http://localhost:3000", "http://localhost:5173", "https://localhost:5173", "https://histopathai.com", "https://histopathai.com.tr"]
cookie_domain       = ""
log_levels          = "info"
log_format          = "json"