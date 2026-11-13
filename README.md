# intoto-github-demo

A GitHub Actions pipeline builds a small container image, attests it with [`actions/attest`](https://github.com/actions/attest), and records those attestations in [Kosli](https://www.kosli.com) with `kosli attest in-toto --source github`.

The build already produces signed SLSA provenance and an SBOM attestation, and GitHub stores them. Kosli reads them back from GitHub by the image's digest, verifies the signature, and records them. No bundle file moves between steps, and the pipeline doesn't regenerate the same evidence with `kosli attest artifact` and `kosli attest sbom`.

## What the pipeline does

The workflow is [`.github/workflows/build.yml`](.github/workflows/build.yml). It runs on every push to `main` and on pull requests from this repository.

1. **Build and push the image.** [`main.go`](main.go) is a one-route web server. The image goes to `ghcr.io/kosli-labs/intoto-github-demo`, tagged with the commit SHA. Everything after this step names the image by the digest the push returned.
2. **Attest the provenance.** `actions/attest` writes a SLSA provenance statement for that digest, signs it with a short-lived Sigstore certificate that names this workflow, logs the signature in the Rekor transparency log, and stores the attestation with GitHub.
3. **Attest the SBOM.** [`anchore/sbom-action`](https://github.com/anchore/sbom-action) generates an SPDX SBOM for the image, and `actions/attest` signs and stores it in the same way.
4. **Build the Kosli CLI.** `--source github` is on the CLI's `main` branch but not yet in a release, so the workflow builds the CLI from a pinned commit.
5. **Create the flow and begin a trail.** The flow is `intoto-github-demo`, with the template in [`kosli-flow.yml`](kosli-flow.yml): an artifact named `app` with an attestation named `sbom`. The trail is named after the commit SHA.
6. **Record the provenance.** `kosli attest in-toto --predicate-type provenance --name app` finds the provenance GitHub holds for the digest, verifies it, and reports the image as the artifact `app`. The commit, repository and build URL come from the provenance statement.
7. **Record the SBOM.** The same command with `--predicate-type sbom --name app.sbom` records the SBOM attestation against that artifact.

## Where it records

This repository records to the Kosli org `in-toto-demo`. Two settings under **Settings › Secrets and variables › Actions** connect the workflow to it:

| Kind | Name | Value |
| --- | --- | --- |
| Repository secret | `KOSLI_API_TOKEN` | An API token that can write to the org |
| Repository variable | `KOSLI_ORG` | The name of the Kosli org |

The recording steps need no other configuration: in GitHub Actions the CLI looks for attestations in this repository, and only accepts one signed by a workflow in this repository.

In a copy of this repository without those two settings, steps 5 to 7 are skipped and the run still passes. Set both to record to an org of your own.

## What to look at in a demo

1. **The workflow run**, under [Actions](https://github.com/kosli-labs/intoto-github-demo/actions). The log of **Build and push the image** shows the digest that everything else refers to. The log of each **Attest** step ends with two links: the attestation on GitHub, and its entry in the Rekor transparency log.
2. **The attestations GitHub holds**, on the repository's [Attestations page](https://github.com/kosli-labs/intoto-github-demo/attestations). Each run adds two for its image: one with the predicate type `https://slsa.dev/provenance/v1`, one with `https://spdx.dev/Document/v2.3`. Open one to see the subject digest, the workflow that signed it, and the commit it was built from.
3. **The transparency log entry**, at the `search.sigstore.dev` link from the **Attest** step's log. This is a public record, outside both GitHub and Kosli, that this workflow signed an attestation for this digest at this time.
4. **The trail in Kosli**, at `https://app.kosli.com/in-toto-demo/flows/intoto-github-demo/trails/<commit sha>`. The **Begin the Kosli trail** step's log prints the trail's name. The artifact `app` has the image's digest as its fingerprint and an `sbom` attestation. Both carry what verification established as annotations: `verified`, `signer_identity`, `signer_issuer`, `log_index` and `integrated_time`. The provenance adds `builder_id`, `build_type` and `workflow`; the SBOM adds `sbom_format` and `sbom_sha256`. The `attestation` link on each goes to the same Rekor entry as step 3.

On a pull request, GitHub builds the merge of the branch into `main`, so the commit SHA in the image tag, the provenance and the trail name is that merge commit's, and the signer is `build.yml@refs/pull/<number>/merge`.
