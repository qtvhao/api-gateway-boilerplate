# Authorization Policies
# Open Policy Agent (OPA) authorization rules for API Gateway

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
    resource_accessible_to_user
}

# Check if user has required role for the resource
user_has_required_role if {
    # Admin users can access everything
    "admin" in input.user.roles
}

user_has_required_role if {
    # System admins can access everything
    "system_admin" in input.user.roles
}

user_has_required_role if {
    # Developer role can access project and task management
    "developer" in input.user.roles
    startswith(input.path, "/api/v1/projects/")
}

user_has_required_role if {
    # Manager role can access goals and analytics
    "manager" in input.user.roles
    path_allowed_for_manager
}

path_allowed_for_manager if {
    startswith(input.path, "/api/v1/goals/")
}

path_allowed_for_manager if {
    startswith(input.path, "/api/v1/analytics/")
}

path_allowed_for_manager if {
    startswith(input.path, "/api/v1/hr/")
}

user_has_required_role if {
    # HR role can access employee and wellbeing data
    "hr" in input.user.roles
    path_allowed_for_hr
}

path_allowed_for_hr if {
    startswith(input.path, "/api/v1/hr/")
}

path_allowed_for_hr if {
    startswith(input.path, "/api/v1/wellbeing/")
}

user_has_required_role if {
    # Regular users can view their own data
    "user" in input.user.roles
    input.method == "GET"
}

# Check if resource is accessible to the user
resource_accessible_to_user if {
    # Admins can access all resources
    "admin" in input.user.roles
}

resource_accessible_to_user if {
    # System admins can access all resources
    "system_admin" in input.user.roles
}

resource_accessible_to_user if {
    # Users can access resources in their tenant
    input.user.tenant_id == input.resource.tenant_id
}

resource_accessible_to_user if {
    # Users can access their own resources
    input.user.user_id == input.resource.owner_id
}

# Rate limit override for premium users
rate_limit_override if {
    "premium" in input.user.tags
}
