# Releasing the Atlas Terraform provider

The provider publishes to the Terraform Registry under the **`Atlas-Authorization`**
namespace — customers write `source = "Atlas-Authorization/atlas"`. The release is a
GPG-signed GitHub Release built by `.goreleaser.yml` + `.github/workflows/release.yml`;
the registry ingests it automatically.

## Status

| Step | State |
|------|-------|
| Dedicated public repo `Atlas-Authorization/terraform-provider-atlas` | ✅ created, code pushed |
| GoReleaser config + release workflow + registry manifest | ✅ in the repo (validated: all 13 OS/arch targets build + checksum) |
| GPG signing key | ✅ generated — fingerprint `205CBD00CF2E0D70EE48270BBC3B43D3BAC81F1C` |
| Repo Actions secrets `GPG_PRIVATE_KEY` + `PASSPHRASE` | ✅ set |
| Provider address in `main.go` | ✅ `registry.terraform.io/Atlas-Authorization/atlas` |
| **Register the provider on the Terraform Registry** | ⬜ you (one-time web step, below) |
| **Tag `v0.1.0` to publish** | ⬜ after registration |

## 1. Register on the Terraform Registry (one-time, web — only you can)

1. Sign in at **https://registry.terraform.io** with GitHub, as a member of the
   **Atlas-Authorization** org.
2. **Publish → Provider** → authorize the Terraform Registry GitHub app for the org →
   select **`terraform-provider-atlas`**.
3. When asked for a **GPG signing key**, paste the **public** key block. Re-export it
   any time with:
   ```bash
   gpg --armor --export 205CBD00CF2E0D70EE48270BBC3B43D3BAC81F1C
   ```
   (The public key was also printed in the session that set this up.)

## 2. Cut the first release

```bash
git clone git@github.com:Atlas-Authorization/terraform-provider-atlas.git
cd terraform-provider-atlas
git tag v0.1.0
git push origin v0.1.0
```

The `release` workflow runs on the tag: builds every platform, writes
`terraform-provider-atlas_0.1.0_SHA256SUMS`, GPG-signs it, and creates the GitHub
Release. The registry picks it up within a few minutes. Then anyone can:

```hcl
terraform {
  required_providers {
    atlas = {
      source  = "Atlas-Authorization/atlas"
      version = "~> 0.1"
    }
  }
}
```

## Notes
- Re-releasing: push a new tag (`v0.1.1`, …); the registry ingests each one.
- The signing key's private material + passphrase are in
  `~/Desktop/atlas-terraform-signing-key.txt` (already loaded into the repo secrets —
  you can delete that file once you've confirmed a release works).
- Protocol **6.0** (terraform-plugin-framework), set in `terraform-registry-manifest.json`.
- `docs/` at the repo root is rendered as the provider's registry documentation.
- Optional cleanup: the `go.mod` module path is still `github.com/atlas/terraform-provider-atlas`;
  it builds fine as-is, but you can rename it to the real repo path later.
