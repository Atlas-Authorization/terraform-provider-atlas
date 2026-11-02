# Create a pull source for an organization: provider (entra, google_workspace, bamboohr, hibob, personio, workday, successfactors, gusto, csv, rest_json), config, write-only credentials and settings (schedule, role mappings, disable on removal).
resource "atlas_directory_pull_source" "example" {
  organization_id = "tf-directory-pull-source"
  name = "tf-directory-pull-source"
  provider_value = "entra"
}
