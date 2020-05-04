path "dctna/data/dev" {
    capabilities = [ "read" ]
}

# Create read and update secrets
path "dctna/data/dev/*" {
  capabilities = ["create", "read", "update"]
}

# Delete last version of secret
path "dctna/data/dev/*" {
  capabilities = ["delete"]
}

# Delete any version of secret
path "dctna/delete/dev/*" {
  capabilities = ["update"]
}

# Destroy version
path "dctna/destroy/dev/*" {
  capabilities = ["update"]
}

# Un-delete any version of secret
path "dctna/undelete/dev/*" {
  capabilities = ["update"]
}

# list keys, view metadata and permanently remove all versions and destroy metadata for a key
path "dctna/metadata/dev/*" {
  capabilities = ["list", "read", "delete"]
}

# password generation
path "gen/password" {
  capabilities = ["create", "update"]
}
