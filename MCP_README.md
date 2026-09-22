# Massdriver MCP Server — Tool Reference

This document describes all 112 tools available in the Massdriver MCP server.

## Conventions

- **Behavioral annotations.** Every tool advertises MCP hints: read-only tools set `readOnlyHint`; mutating tools set `destructiveHint`/`idempotentHint` so clients can gate or auto-approve calls appropriately.
- **Enum validation.** Fields with a closed set of values (e.g. `effect`, `scope`, deployment `action`/`status`, resource `origin`, `get_url` `type`) are constrained with JSON Schema enums and rejected client-side if invalid.
- **Handled failures.** When the API rejects a mutation (e.g. a missing required attribute), the tool returns a result with `isError: true` and a human-readable message rather than appearing to succeed.
- **Trimmed payloads.** Bundle and OCI-repo responses omit the inline SVG `icon` field to keep responses compact.
- **Pagination.** List tools return `{ items, has_more, next_cursor }` and accept `cursor`/`page_size` (default 25, max 100). `list_components` is unpaginated (a project's blueprint is bounded).

## Projects

| Tool | Description |
|------|-------------|
| `list_projects` | Lists all projects in the organization, including their environments. Optionally filter by `search` (free-text over name and description), `name` (exact match), `name_in` (any of several exact names), or `created_after`/`created_before` (RFC 3339 creation window). |
| `get_project` | Gets a specific project by ID, including its environments. |
| `create_project` | Creates a new project. Requires `id` and `name`; accepts optional `description` and custom `attributes` (required by orgs that define required project attributes). |
| `clone_project` | Clones an existing project into a new project, duplicating its blueprint structure (components and wiring) without copying any environments. Requires `source_project_id`, `id`, and `name`. |
| `update_project` | Updates a project's name, description, or custom `attributes`. |
| `delete_project` | Deletes a project. All environments must be empty first. |

## Environments

| Tool | Description |
|------|-------------|
| `list_environments` | Lists all environments. Optionally filter by `project_id`. |
| `get_environment` | Gets an environment by its identifier (e.g., `myproj-staging`). |
| `create_environment` | Creates an environment within a project. Requires `project_id`, `id`, `name`; accepts optional `description`, custom `attributes`, `decommission_protection`, and `separation_of_duty`. |
| `update_environment` | Updates an environment's name, description, custom `attributes`, `decommission_protection`, or `separation_of_duty`. Only the fields you provide change; omitted fields are left unchanged. |
| `delete_environment` | Deletes an environment. All instances must be decommissioned first. |
| `set_environment_default` | Sets a resource as the default of its type for an environment. The resource must first be shared to the environment via `create_resource_grant`. |
| `remove_environment_default` | Removes a default resource binding. |
| `compare_environments` | Compares two environments in the same project instance-by-instance (paired by component), reporting the resolved bundle version on each side and a leaf-level diff of configured params. |
| `fork_environment` | Forks a new environment from a parent environment in the same project. Requires `parent_id`, `id`, `name`; optional toggles `copy_secrets`, `copy_remote_references`, and `copy_environment_defaults` (all default false) control what carries over; optional `decommission_protection` and `separation_of_duty` set the fork's guards. |
| `deploy_environment` | Schedules a deployment of every instance in the environment in dependency order. Cancels any in-flight environment deployment and enqueues a fresh provision wave; changes happen asynchronously. |
| `decommission_environment` | Schedules a teardown of every instance in the environment in reverse dependency order (the environment shell stays; use `delete_environment` to remove it afterwards). Blocked when decommission protection is enabled. |
| `list_environment_links` | Lists the blueprint links in effect in an environment given the component versions its instances actually run — the subset of the project's links whose version constraints match both ends. |
| `list_environment_unfulfilled_dependencies` | Lists required dependency inputs across the environment's instances that nothing fills (no link, no remote reference, no matching environment default) — each entry is one input a deploy would block on. |

## Instances

| Tool | Description |
|------|-------------|
| `list_instances` | Lists instances. Optionally filter by `project_id`, `environment_id`, or `status`. |
| `get_instance` | Gets an instance by ID, including environment, project, and release info. |
| `update_instance` | Updates an instance's version pin. |
| `set_instance_secret` | Sets or updates a secret on an instance. |
| `remove_instance_secret` | Removes a secret from an instance. |
| `set_remote_reference` | Overrides one of an instance's connection slots (`field`) with a resource from another project or an imported resource (`resource_id`). Takes priority over any blueprint link on that slot. |
| `remove_remote_reference` | Removes a remote-reference override from a connection slot, reverting it to the blueprint link or environment default. |
| `copy_instance` | Copies a source instance's configuration onto an existing destination instance (overwriting its params). Source and destination must be instances of the same blueprint component (same manifest). Requires `source_id` and `destination_id`; optional `overrides` (deep-merged onto source params), `copy_secrets`, and `copy_remote_references`. |
| `orphan_instance` | Break-glass reset of a permanently-stuck instance to INITIALIZED, clearing state locks and bulk-aborting active deployments. Optional `delete_state` also removes the remote IaC state files (IRREVERSIBLE — next deploy provisions from scratch). |
| `list_alarms` | Lists alarms. Optionally filter by project, environment, component, instance, or bundle. |

## Deployments

| Tool | Description |
|------|-------------|
| `list_deployments` | Lists deployments (newest first). Optionally filter by `instance_id`, `status`, or `action`. |
| `get_deployment` | Gets a deployment by ID. |
| `get_deployment_logs` | Gets a deployment's logs. With `follow: true`, blocks until the deployment reaches a terminal status and returns the final status plus complete logs (optional `timeout_seconds`, default 300, max 600). |
| `create_deployment` | Creates and starts a deployment. Actions: `PROVISION`, `DECOMMISSION`, `PLAN`. `params` is required for every action (full set, validated against the instance's params schema) — read `get_instance.params` to reuse current config. |
| `propose_deployment` | Proposes a deployment for approval (enters `PROPOSED` status). Actions: `PROVISION`, `DECOMMISSION`. `params` is required (full set, validated against the instance's params schema). |
| `approve_deployment` | Approves a proposed deployment. In environments with separation of duty enabled, the proposer's own approval is rejected — a second reviewer must approve. |
| `reject_deployment` | Rejects a proposed deployment. |
| `abort_deployment` | Aborts a running deployment. |
| `plan_deployment` | Runs a fresh PLAN (dry-run preview) against an existing deployment's params. Mutates nothing on the source; returns a new PLAN deployment. |
| `rollback_deployment` | Proposes a rollback to a past deployment's exact state. The source must be a `COMPLETED` `PROVISION`; creates a `PROPOSED` `PROVISION` for review (approve or reject). |
| `compare_deployments` | Compares two deployments' snapshotted config: bundle version on each side plus a leaf-level diff of params. |

## Components

| Tool | Description |
|------|-------------|
| `list_components` | Lists all components in a project's blueprint. Requires `project_id`. |
| `get_component` | Gets a component by ID. |
| `add_component` | Adds a component to a project blueprint. Requires `project_id`, `bundle_name`, `id`, `name`; accepts optional `description` and custom `attributes`. |
| `update_component` | Updates a component's name, description, or custom `attributes`. |
| `set_component_position` | Sets a component's pixel position (`x`, `y`) on the project's visual canvas. Purely cosmetic — does not affect configuration or deployments. |
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
| `list_resources` | Lists resources. Optionally filter by `origin`, `resource_type` (optionally version-pinned, e.g. `aws-iam-role@1.2.3`), `environment_id`, `search`, `created_after`/`created_before` (RFC 3339 creation window), or `attributes` (effective-attribute filters, AND'd). |
| `get_resource` | Gets a resource by ID (payload values are masked). |
| `create_resource` | Imports a resource. Requires `resource_type_id` (optionally pinned to an exact version, e.g. `aws-iam-role@1.2.3`) and `name`; the payload must conform to the resource type's schema (see `get_resource_type`). |
| `update_resource` | Updates a resource's name or payload. |
| `delete_resource` | Deletes an imported resource. |
| `export_resource` | Exports a resource with unmasked payload (audit-logged). |
| `create_resource_grant` | Creates a sharing grant on a resource. `action` must be `resource:export` (the only grantable action). |
| `delete_resource_grant` | Deletes a sharing grant. |
| `list_resource_grants` | Lists sharing grants on a resource. |

## Resource Types

| Tool | Description |
|------|-------------|
| `get_resource_type` | Gets a resource type — the contract behind the connection system — including its full JSON schema, import instructions, and connection orientation. The ID's version portion may be exact (`@1.2.3`), a range (`@~1`), a channel (`@latest`), or omitted; the response carries the fully resolved version. |
| `list_resource_type_dependents` | Lists the instances in an environment that depend on a resource type, one entry per (instance, dependency field) pair. Use before changing or removing a resource of that type. |

## Organization

| Tool | Description |
|------|-------------|
| `get_organization` | Gets the current organization's details (id, name, subscription status, timestamps). Custom attributes and members are not included here — use `list_custom_attributes` and `list_organization_members` for those. |
| `create_custom_attribute` | Creates a custom attribute definition. Requires `key` and `scope`. `required` defaults to false here; setting it true makes the attribute mandatory org-wide at its scope. |
| `update_custom_attribute` | Updates a custom attribute's required flag or allowed values. |
| `delete_custom_attribute` | Deletes a custom attribute definition. |
| `list_organization_members` | Lists the organization's members (user accounts), paginated. Requires the `organization:manageProfile` permission (non-admin tokens get a forbidden error). |
| `list_custom_attributes` | Lists the organization's custom attribute definitions (keys, scopes, allowed values), paginated. Use to discover declared attribute keys before setting attributes or writing policy conditions. |
| `get_organization_settings` | Gets the organization's behavior settings (`default_bundle_access`: the access new bundle repositories receive at creation). Requires the `organization:manageSettings` permission. |
| `update_organization_settings` | Updates the organization's behavior settings. `default_bundle_access` = `NONE` (new repos restricted until granted) or `ALL_PROJECTS` (auto-creates an org-wide `repo:pull` grant on each new repo); only affects repositories created afterwards. Requires the `organization:manageSettings` permission. |

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
| `get_group` | Gets a group by ID (id, name, description, role, timestamps). Members, service accounts, and policies are not included. |
| `create_group` | Creates a new group. Requires `name`. |
| `update_group` | Updates a group's name or description. |
| `delete_group` | Deletes a group. |
| `add_group_user` | Adds a user to a group by email (sends invitation if not yet a member). |
| `remove_group_user` | Removes a user from a group. |
| `revoke_group_invitation` | Revokes a pending group invitation. |
| `add_group_service_account` | Adds a service account to a group. |
| `remove_group_service_account` | Removes a service account from a group. |
| `list_group_members` | Lists the users who are members of a group, paginated. |
| `list_group_service_accounts` | Lists the service accounts that are members of a group, paginated. |
| `list_group_invitations` | Lists pending email invitations to a group (invited but not yet joined), paginated. |
| `list_group_policies` | Lists the ABAC policies attached to a group, paginated. |

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
| `list_oci_repos` | Lists OCI repositories. Optionally filter by `search`, `artifact_type`, `created_after`/`created_before` (RFC 3339 creation window), or `attributes` (repository-attribute filters, AND'd). |
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
| `create_policy` | Creates a policy on a group. Requires `group_id`, `effect` (`ALLOW`/`DENY`), and a non-empty `actions` list (an empty list grants nothing, not all actions). |
| `update_policy` | Updates a policy's effect, actions, or conditions. |
| `delete_policy` | Deletes a policy. |
| `list_policy_actions` | Lists all available policy actions. |
| `list_policy_entities` | Lists all entity kinds that policies can target. |
| `evaluate_policy` | Checks if the caller is allowed to perform an action on an entity. |
| `evaluate_policies_batch` | Checks multiple action/entity pairs in one request (max 10). |
| `explain_policy` | Returns a human-readable explanation of a policy configuration. Requires `effect` and a non-empty `actions` list. Conditions referencing attribute keys not declared in the org are silently dropped. |
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
