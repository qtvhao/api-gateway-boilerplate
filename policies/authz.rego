# Authorization Policies
# Open Policy Agent (OPA) authorization rules for API Gateway
# Customize these rules based on your application requirements

package authz

import future.keywords.if
import future.keywords.in

# Default deny all requests
default allow := false

# Allow health check endpoints to everyone
allow if {
    input.path == "/health"
}

allow if {
    input.path == "/health/ready"
}

allow if {
    input.path == "/health/live"
}

# Allow public endpoints
allow if {
    startswith(input.path, "/api/v1/public/")
}

# Authenticated user rules
allow if {
    input.user.authenticated
    user_has_required_role
}

# Check if user has required role for the resource
user_has_required_role if {
    # Admin users can access everything
    "admin" in input.user.roles
}

user_has_required_role if {
    # Regular authenticated users can access non-admin endpoints
    "user" in input.user.roles
    not startswith(input.path, "/api/v1/admin/")
}

# Rate limit override for premium users (example)
rate_limit_override if {
    "premium" in input.user.tags
}
