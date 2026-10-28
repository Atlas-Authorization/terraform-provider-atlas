# Define a resource type: a `key`, its `roles` (each a set of permissions, optionally `includes` other roles), an optional `parent_type` resources nest under, and `inherit` (default true) controlling whether assignments on a parent flow down.
resource "atlas_resource_type" "example" {

}
