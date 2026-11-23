# Put one directory on the allow or block list; it leaves the other list if it was there.
resource "atlas_web_bot_auth_directory" "example" {
  directory = "tf-web-bot-auth-directory"
  list = "allow"
}
