# Contributing to Terraform Language Server

## Reporting Feedback

The hcl-lang library is an open source project and we appreciate contributions of various kinds, including bug reports and fixes, enhancement proposals, documentation updates, and user experience feedback.

To record a bug report, enhancement proposal, or give any other product feedback, please [open a GitHub issue](https://github.com/hashicorp/hcl-lang/issues/new/choose) using the most appropriate issue template. Please do fill in all of the information the issue templates request, because we've seen from experience that this will maximize the chance that we'll be able to act on your feedback.

**All communication on GitHub, the community forum, and other HashiCorp-provided communication channels is subject to [the HashiCorp community guidelines](https://www.hashicorp.com/community-guidelines).**

## Scope

This repository contains the source code only for hcl-lang, which is relied upon by other projects that have their own repositories.

This is primarily used by the [Terraform Language Server](https://github.com/hashicorp/terraform-ls), but it's still in early stage of development.

For that reason the API is not considered stable yet and should not be relied upon.

## Proposing a Change

If you'd like to contribute a code change, we'd love to review a GitHub pull request.

In order to be respectful of the time of community contributors, we prefer to discuss potential changes in GitHub issues prior to implementation.

That will allow us to give design feedback up front and set expectations about the scope of the change, and, for larger changes, how best to approach the work such that the maintainer team can review it and merge it along with other concurrent work.

If the bug you wish to fix or enhancement you wish to implement isn't already covered by a GitHub issue that contains feedback from the maintainer team, please do start a discussion (either in [a new GitHub issue](https://github.com/hashicorp/hcl-lang/issues/new/choose) or an existing one, as appropriate) before you invest significant development time.

If you mention your intent to implement the change described in your issue, the maintainer team can prioritize including implementation-related feedback in the subsequent discussion.

Most changes will involve updates to the test suite, and changes to the documentation.

The maintainer team can advise on different testing strategies for specific scenarios, and may ask you to revise the specific phrasing of your proposed documentation prose to match better with the standard "voice" of Terraform's documentation.

This repository is primarily maintained by a small team at HashiCorp along with their other responsibilities, so unfortunately we cannot always respond promptly to pull requests, particularly if they do not relate to an existing GitHub issue where the maintainer team has already participated. We _are_ grateful for all contributions however, and will give feedback on pull requests as soon as we're able to.
