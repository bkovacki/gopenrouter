<a name="unreleased"></a>
## [Unreleased]


<a name="v0.4.0"></a>
## [v0.4.0] - 2025-06-03
### Bug Fixes
- update completion request builder to return pointer
- remove unused buffer fields and lint errors

### Code Refactoring
- move streaming code to completion and chat modules

### Documentation
- update README and enhance streaming documentation
- consolidate examples to use main module structure

### Features
- add logprobs support to completion and chat APIs
- add streaming support for real-time completions
- add logprobs and stop parameters to completion requests


<a name="v0.3.0"></a>
## [v0.3.0] - 2025-05-29
### Documentation
- add examples of usage to README and repo

### Features
- add chat completion endpoint support

### Maintenance
- prepare release v0.3.0

### Tests
- align test and subtest names to pascal case


<a name="v0.2.0"></a>
## [v0.2.0] - 2025-05-28
### Bug Fixes
- change CreditsData fields from float32 to float64
- update ModelData fields for improved API compatibility

### Maintenance
- prepare release v0.2.0


<a name="v0.1.0"></a>
## v0.1.0 - 2025-05-23
### Bug Fixes
- encode list endpoints path params

### Features
- add completion endpoint support
- add querying cost and stats

### Maintenance
- prepare release v0.1.0

### Tests
- add tests for list endpoints


[Unreleased]: https://github.com/bkovacki/gopenrouter/compare/v0.4.0...HEAD
[v0.4.0]: https://github.com/bkovacki/gopenrouter/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/bkovacki/gopenrouter/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/bkovacki/gopenrouter/compare/v0.1.0...v0.2.0
