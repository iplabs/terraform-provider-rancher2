# This example user is managed by an external Active Directory system
# and therefore requires no configuration other than the "distinguished name"
#data "rancher2_user" "existing_ad_user" {
#  activedirectory_user = "CN=Doe\\, John,OU=Software Development,OU=user"
#}

# A local user
resource "rancher2_user" "local_user" {
  username = "jane_doe"
  password = "secret"
  name = "Jane Doe"
  description = "Your friendly colleague."
}