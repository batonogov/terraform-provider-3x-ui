# Changelog

## [0.6.1](https://github.com/batonogov/terraform-provider-threexui/compare/v0.6.0...v0.6.1) (2026-04-24)


### Bug Fixes

* eliminate false drift on import for Optional+Computed inbound fields ([#100](https://github.com/batonogov/terraform-provider-threexui/issues/100)) ([2fcc66f](https://github.com/batonogov/terraform-provider-threexui/commit/2fcc66fa29fa311e16e5a8111158042ba7b54381))

## [0.6.0](https://github.com/batonogov/terraform-provider-threexui/compare/v0.5.3...v0.6.0) (2026-04-21)


### Features

* support 3x-ui 2.9.0 ([#91](https://github.com/batonogov/terraform-provider-threexui/issues/91)) ([db546eb](https://github.com/batonogov/terraform-provider-threexui/commit/db546eb5e160e6db428a52a551dbdc4c0204ed9f))

## [0.5.3](https://github.com/batonogov/terraform-provider-threexui/compare/v0.5.2...v0.5.3) (2026-03-20)


### Bug Fixes

* add markdownlint to pre-commit and fix MD060 violations ([#89](https://github.com/batonogov/terraform-provider-threexui/issues/89)) ([5fd6f6d](https://github.com/batonogov/terraform-provider-threexui/commit/5fd6f6d897dc6bbdc148578f25162b71dafb21a8))
* align tunnel inbound support with 3x-ui 2.8.11 ([#87](https://github.com/batonogov/terraform-provider-threexui/issues/87)) ([29adfce](https://github.com/batonogov/terraform-provider-threexui/commit/29adfced9a0138de42627b2f4132c213701fd584))

## [0.5.2](https://github.com/batonogov/terraform-provider-threexui/compare/v0.5.1...v0.5.2) (2026-03-14)


### Bug Fixes

* update client base_path after web_base_path change ([#81](https://github.com/batonogov/terraform-provider-threexui/issues/81)) ([2e890b0](https://github.com/batonogov/terraform-provider-threexui/commit/2e890b0ee3940765eade887ebd98a85a5f22182e))

## [0.5.1](https://github.com/batonogov/terraform-provider-threexui/compare/v0.5.0...v0.5.1) (2026-03-12)


### Bug Fixes

* fail fast when inbound settings JSON cannot be parsed ([#61](https://github.com/batonogov/terraform-provider-threexui/issues/61)) ([b764019](https://github.com/batonogov/terraform-provider-threexui/commit/b764019f9a85f75aa783e6308980d3a40ec6cefe))

## [0.5.0](https://github.com/batonogov/terraform-provider-threexui/compare/v0.4.2...v0.5.0) (2026-03-12)


### Features

* add threexui_client_traffics data source ([#56](https://github.com/batonogov/terraform-provider-threexui/issues/56)) ([4bb4d78](https://github.com/batonogov/terraform-provider-threexui/commit/4bb4d785584cbb682dc8926715cc7c0958c16f85))
* add threexui_xray_version resource ([#57](https://github.com/batonogov/terraform-provider-threexui/issues/57)) ([1dd9be6](https://github.com/batonogov/terraform-provider-threexui/commit/1dd9be6c8f217d01039094069859ec6cbb55c355))


### Bug Fixes

* normalize inbound listen and total defaults ([#53](https://github.com/batonogov/terraform-provider-threexui/issues/53)) ([f6e700e](https://github.com/batonogov/terraform-provider-threexui/commit/f6e700ec4d42d7c11e57eb31554a614310e9b2de))
* serialize panel_general updates with settingsMu and xrayTemplateMu ([#59](https://github.com/batonogov/terraform-provider-threexui/issues/59)) ([951b4cc](https://github.com/batonogov/terraform-provider-threexui/commit/951b4cc9c6e626dd12e7f96c6bb6c3c33b46b5d8))

## [0.4.2](https://github.com/batonogov/terraform-provider-threexui/compare/v0.4.1...v0.4.2) (2026-03-06)


### Bug Fixes

* remove email fallback in getClientIDFromModel ([#46](https://github.com/batonogov/terraform-provider-threexui/issues/46)) ([d496d93](https://github.com/batonogov/terraform-provider-threexui/commit/d496d9368bf30611b67bc284f0d9faafffcacc46))

## [0.4.1](https://github.com/batonogov/terraform-provider-threexui/compare/v0.4.0...v0.4.1) (2026-03-06)


### Bug Fixes

* add default empty string for inbound remark attribute ([#44](https://github.com/batonogov/terraform-provider-threexui/issues/44)) ([64cd3e3](https://github.com/batonogov/terraform-provider-threexui/commit/64cd3e300b0d41bd5660f03b75e8e6b28a7d84d9))

## [0.4.0](https://github.com/batonogov/terraform-provider-threexui/compare/v0.3.0...v0.4.0) (2026-03-06)


### Features

* add 3x-ui v2.8.10 support and xray_outbound_test_url attribute ([#24](https://github.com/batonogov/terraform-provider-threexui/issues/24)) ([4d624b2](https://github.com/batonogov/terraform-provider-threexui/commit/4d624b2f72f5b9a23e5c35c0c628c15dfe41b5df))
* add 3x-ui v2.8.11 support ([#34](https://github.com/batonogov/terraform-provider-threexui/issues/34)) ([36556eb](https://github.com/batonogov/terraform-provider-threexui/commit/36556eb4e44cb63ec498fd3ac7a58b7a7dfaf41f))
* add inbound tests and uuid helpers ([09fd75a](https://github.com/batonogov/terraform-provider-threexui/commit/09fd75ae42719ed86479b35e01c78d1de1417718))
* add inbound tests and uuid helpers ([e625755](https://github.com/batonogov/terraform-provider-threexui/commit/e6257552cb437a712e287cbf42a15eff47066749))
* add pre-commit hooks configuration ([99d5056](https://github.com/batonogov/terraform-provider-threexui/commit/99d5056724fc72ea24f7617f4716740e5cf64309))
* add pre-commit hooks configuration ([414aff5](https://github.com/batonogov/terraform-provider-threexui/commit/414aff52b837e9b64bdb876c3bf083d4b8e49f6b))
* add pre-commit hooks configuration ([433393e](https://github.com/batonogov/terraform-provider-threexui/commit/433393e83ad25d3a9c556231e46295c970c4387c))
* add provider settings resources and update examples ([705483f](https://github.com/batonogov/terraform-provider-threexui/commit/705483f7f37c60ee21dcfe87d07c615e68573aa3))
* add provider settings resources and update examples ([ba8fea7](https://github.com/batonogov/terraform-provider-threexui/commit/ba8fea7abed84b0ac3731e863ac7ac05b4fa835d))
* add Release Please for automated releases ([#29](https://github.com/batonogov/terraform-provider-threexui/issues/29)) ([02f4880](https://github.com/batonogov/terraform-provider-threexui/commit/02f48801e6204c71d5d584ac039806d746dac52d))
* add test plans, update CLAUDE.md and provider improvements ([ec6fcb2](https://github.com/batonogov/terraform-provider-threexui/commit/ec6fcb2003c1b2373cbb20b2469200664404849e))
* add threexui_online_clients data source ([#40](https://github.com/batonogov/terraform-provider-threexui/issues/40)) ([d7df4d2](https://github.com/batonogov/terraform-provider-threexui/commit/d7df4d2cf012c20968a97f689a4b202da9fc684a))
* add threexui_panel_user resource ([#18](https://github.com/batonogov/terraform-provider-threexui/issues/18)) ([e166e92](https://github.com/batonogov/terraform-provider-threexui/commit/e166e92a2601ce25bce44abf2199e92e07903317))
* implement initial provider skeleton ([2877b0f](https://github.com/batonogov/terraform-provider-threexui/commit/2877b0f7e5f4dfc4df303cee104e000bcb6d1c75))
* implement initial provider skeleton ([5a590f6](https://github.com/batonogov/terraform-provider-threexui/commit/5a590f628c4bb8bfb65666e061c0bf91e271a0bb))
* replace JSON string attributes with typed blocks in threexui_inbound ([#22](https://github.com/batonogov/terraform-provider-threexui/issues/22)) ([1a979a7](https://github.com/batonogov/terraform-provider-threexui/commit/1a979a7f6c624ef19c6292aab727ebf0fdbd5f37))
* switch from OpenTofu to Terraform Registry ([#28](https://github.com/batonogov/terraform-provider-threexui/issues/28)) ([fb26e72](https://github.com/batonogov/terraform-provider-threexui/commit/fb26e72d47deab62b5d8b1346bd5ed5b1320ffaf))


### Bug Fixes

* add mutex and subset diff suppress for xray settings ([906c4a7](https://github.com/batonogov/terraform-provider-threexui/commit/906c4a709474ac7156af544e20ef7cc4690dd494))
* add warnings for 2FA/base_path and fix sub_json_enable persistence ([0cf40e2](https://github.com/batonogov/terraform-provider-threexui/commit/0cf40e22912a5bb4e7e235dac58c2258288349b3))
* **ci:** use golangci-lint-action v7 for golangci-lint v2 config ([418de60](https://github.com/batonogov/terraform-provider-threexui/commit/418de60fc01139186e6014ff260f1a5f3386e587))
* correct TestBuildAndFlattenSettings assertions ([4825e90](https://github.com/batonogov/terraform-provider-threexui/commit/4825e90bada76aa587920b11f584af5ee75ccea9))
* fix remaining golangci-lint v2 config issues ([f1b10c4](https://github.com/batonogov/terraform-provider-threexui/commit/f1b10c405080b77076cbcb0e29418e414e87ca7f))
* make client email required to prevent 3x-ui SQL errors ([4bc4543](https://github.com/batonogov/terraform-provider-threexui/commit/4bc454376dad4f02a87521f388bd6e72ef52803c))
* migrate .golangci.yml to v2 format ([e57f5b7](https://github.com/batonogov/terraform-provider-threexui/commit/e57f5b73d0d16b9089b2fa5b22e3ba48144b00a4))
* preserve testseed in state by adding flattenIntList to flattenSettings ([c6e6c59](https://github.com/batonogov/terraform-provider-threexui/commit/c6e6c592c26283cbca92ec844bef3922788d1821))
* remove unsupported max-issues options from golangci-lint v2 config ([b471654](https://github.com/batonogov/terraform-provider-threexui/commit/b4716543cf10436ac715f539d9c1152ddc6a9445))
* reset CHANGELOG.md so Release Please recreates v0.3.0 ([#39](https://github.com/batonogov/terraform-provider-threexui/issues/39)) ([9207002](https://github.com/batonogov/terraform-provider-threexui/commit/920700217cf1cecded95070bfb35636c45f257a8))
* resolve all golangci-lint v2 warnings ([45015cc](https://github.com/batonogov/terraform-provider-threexui/commit/45015ccfaf315509cbddaebe53afac95ba64dd8f))
* run GoReleaser inside release-please workflow ([#35](https://github.com/batonogov/terraform-provider-threexui/issues/35)) ([c7599b4](https://github.com/batonogov/terraform-provider-threexui/commit/c7599b4eca9515783f978cf6852596620096458d))

## Changelog
