# Massdriver MCP Server — Tool Reference

This document describes all 87 tools available in the Massdriver MCP server.

## Conventions

- **Behavioral annotations.** Every tool advertises MCP hints: read-only tools set `readOnlyHint`; mutating tools set `destructiveHint`/`idempotentHint` so clients can gate or auto-approve calls appropriately.
- **Enum validation.** Fields with a closed set of values (e.g. `effect`, `scope`, deployment `action`/`status`, resource `origin`, `get_url` `type`) are constrained with JSON Schema enums and rejected client-side if invalid.
- **Handled failures.** When the API rejects a mutation (e.g. a missing required attribute), the tool returns a result with `isError: true` and a human-readable message rather than appearing to succeed.
- **Trimmed payloads.** Bundle and OCI-repo responses omit the inline SVG `icon` field to keep responses compact.
- **Pagination.** List tools return `{ items, has_more, next_cursor }` and accept `cursor`/`page_size` (default 25, max 100). `list_components` is unpaginated (a project's blueprint is bounded).

## Projects

| Tool | Description |
|------|-------------|
| `list_projects` | Lists all projects in the organization, including their environments. |
| `get_project` | Gets a specific project by ID, including its environments. |
| `create_project` | Creates a new project. Requires `id` and `name`; accepts optional `description` and custom `attributes` (required by orgs that define required project attributes). |
| `update_project` | Updates a project's name, description, or custom `attributes`. |
| `delete_project` | Deletes a project. All environments must be empty first. |

## Environments

| Tool | Description |
|------|-------------|
| `list_environments` | Lists all environments. Optionally filter by `project_id`. |
| `get_environment` | Gets an environment by its identifier (e.g., `myproj-staging`). |
| `create_environment` | Creates an environment within a project. Requires `project_id`, `id`, `name`; accepts optional `description` and custom `attributes`. |
| `update_environment` | Updates an environment's name, description, or custom `attributes`. |
| `delete_environment` | Deletes an environment. All instances must be decommissioned first. |
| `set_environment_default` | Sets a resource as the default of its type for an environment. The resource must first be shared to the environment via `create_resource_grant`. |
| `remove_environment_default` | Removes a default resource binding. |

## Instances

| Tool | Description |
|------|-------------|
| `list_instances` | Lists instances. Optionally filter by `project_id`, `environment_id`, or `status`. |
| `get_instance` | Gets an instance by ID, including environment, project, and release info. |
| `update_instance` | Updates an instance's version pin. |
| `set_instance_secret` | Sets or updates a secret on an instance. |
| `remove_instance_secret` | Removes a secret from an instance. |
| `list_alarms` | Lists alarms. Optionally filter by project, environment, component, instance, or bundle. |

## Deployments

| Tool | Description |
|------|-------------|
| `list_deployments` | Lists deployments (newest first). Optionally filter by `instance_id`, `status`, or `action`. |
| `get_deployment` | Gets a deployment by ID. |
| `get_deployment_logs` | Gets a deployment's logs. With `follow: true`, blocks until the deployment reaches a terminal status and returns the final status plus complete logs (optional `timeout_seconds`, default 300, max 600). |
| `create_deployment` | Creates and starts a deployment. Actions: `PROVISION`, `DECOMMISSION`, `PLAN`. |
| `propose_deployment` | Proposes a deployment for approval (enters `PROPOSED` status). Actions: `PROVISION`, `DECOMMISSION`. |
| `approve_deployment` | Approves a proposed deployment. |
| `reject_deployment` | Rejects a proposed deployment. |
| `abort_deployment` | Aborts a running deployment. |

## Components

| Tool | Description |
|------|-------------|
| `list_components` | Lists all components in a project's blueprint. Requires `project_id`. |
| `get_component` | Gets a component by ID. |
| `add_component` | Adds a component to a project blueprint. Requires `project_id`, `bundle_name`, `id`, `name`; accepts optional `description` and custom `attributes`. |
| `update_component` | Updates a component's name, description, or custom `attributes`. |
| `remove_component` | Removes a component from a blueprint. |
| `link_components` | Links two components (source output field to destination input field). |
| `unlink_components` | Removes a link between components. |

## Bundles

| Tool | Description |
|------|-------------|
| `get_bundle` | Gets a bundle by ID. Supports version constraints (e.g., `aws-aurora-postgres@~1`). To list available bundles, use `list_oci_repos` with `artifact_type` set to `BUNDLE`; to list a bundle's versions, use `get_oci_repo` (version tags live on the repository). |

## Resources

| Tool | Description |
|------|-------------|
| `list_resources` | Lists resources. Optionally filter by `origin`, `resource_type`, `environment_id`, or `search`. |
| `get_resource` | Gets a resource by ID (payload values are masked). |
| `create_resource` | Imports a resource. Requires `resource_type_id` and `name`. |
| `update_resource` | Updates a resource's name or payload. |
| `delete_resource` | Deletes an imported resource. |
| `export_resource` | Exports a resource with unmasked payload (audit-logged). |
| `create_resource_grant` | Creates a sharing grant on a resource. |
| `delete_resource_grant` | Deletes a sharing grant. |
| `list_resource_grants` | Lists sharing grants on a resource. |

## Organization

| Tool | Description |
|------|-------------|
| `get_organization` | Gets the current organization's details. |
| `create_custom_attribute` | Creates a custom attribute definition. Requires `key` and `scope`. |
| `update_custom_attribute` | Updates a custom attribute's required flag or allowed values. |
| `delete_custom_attribute` | Deletes a custom attribute definition. |

## Viewer

| Tool | Description |
|------|-------------|
| `get_viewer` | Gets the currently authenticated identity (user or service account). |

## Audit Logs

| Tool | Description |
|------|-------------|
| `get_audit_log` | Gets a specific audit log entry by ID. |
| `list_audit_logs` | Lists audit log entries. Optionally filter by `type`, `actor_id`, or `actor_search`. |
| `list_audit_log_event_types` | Lists all available audit log event types. |

## Groups

| Tool | Description |
|------|-------------|
| `list_groups` | Lists all access control groups. |
| `get_group` | Gets a group by ID, including members, service accounts, and policies. |
| `create_group` | Creates a new group. Requires `name`. |
| `update_group` | Updates a group's name or description. |
| `delete_group` | Deletes a group. |
| `add_group_user` | Adds a user to a group by email (sends invitation if not yet a member). |
| `remove_group_user` | Removes a user from a group. |
| `revoke_group_invitation` | Revokes a pending group invitation. |
| `add_group_service_account` | Adds a service account to a group. |
| `remove_group_service_account` | Removes a service account from a group. |

## Service Accounts

| Tool | Description |
|------|-------------|
| `list_service_accounts` | Lists all service accounts. Optionally filter by `search`. |
| `get_service_account` | Gets a service account by ID. |
| `create_service_account` | Creates a service account. Response includes the bearer token (shown once). |
| `update_service_account` | Updates a service account's name or description. |
| `delete_service_account` | Deletes a service account. |

## OCI Repos

| Tool | Description |
|------|-------------|
| `list_oci_repos` | Lists OCI repositories. Optionally filter by `search` or `artifact_type`. |
| `get_oci_repo` | Gets an OCI repository by ID, including its published version tags. |
| `create_oci_repo` | Creates an OCI repository. Requires `id` and `artifact_type`. |
| `update_oci_repo` | Updates an OCI repository's attributes. |
| `delete_oci_repo` | Deletes an OCI repository. |
| `create_oci_repo_grant` | Creates a sharing grant on an OCI repository. |
| `delete_oci_repo_grant` | Deletes an OCI repository sharing grant. |
| `list_oci_repo_grants` | Lists sharing grants on an OCI repository. |

## Policies

| Tool | Description |
|------|-------------|
| `get_policy` | Gets an ABAC policy by ID. |
| `create_policy` | Creates a policy on a group. Requires `group_id` and `effect` (`ALLOW`/`DENY`). |
| `update_policy` | Updates a policy's effect, actions, or conditions. |
| `delete_policy` | Deletes a policy. |
| `list_policy_actions` | Lists all available policy actions. |
| `list_policy_entities` | Lists all entity kinds that policies can target. |
| `evaluate_policy` | Checks if the caller is allowed to perform an action on an entity. |
| `evaluate_policies_batch` | Checks multiple action/entity pairs in one request (max 10). |
| `explain_policy` | Returns a human-readable explanation of a policy configuration. |
| `get_policy_attribute_schema` | Gets the JSON Schema for valid condition attributes for a policy action. |
| `list_policy_attribute_values` | Lists permitted values for a custom attribute key at a given scope. |

## Server

| Tool | Description |
|------|-------------|
| `get_server` | Gets server metadata including version and authentication methods. |

## URLs

| Tool | Description |
|------|-------------|
| `get_url` | Generates a deep link URL into the Massdriver web UI. Supported types: `organization`, `projects`, `project`, `environment`, `instance`, `bundle`, `repo_instances`. `id` is required for all types except `organization`/`projects`; `version` is additionally required for `bundle`/`repo_instances`. |
