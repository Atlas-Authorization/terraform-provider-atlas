# Connect a device source that Atlas pulls on a schedule: Microsoft Intune (Graph, `tenantId` + app client credentials with DeviceManagementManagedDevices.Read.All), Jamf Pro (`baseUrl` + API client), CrowdStrike Falcon (`cloud` + API client with Zero Trust Assessment and Hosts read; `minZtaScore`, `userEmailDomain`), Fleet (`baseUrl` + API token) or Google Chrome / ChromeOS (service account with domain-wide delegation).
resource "atlas_device_integration" "example" {

}
