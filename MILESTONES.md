# Skyscale Restructuring Milestones

This document tracks the milestones for restructuring Skyscale to support both Lambda-style execution and E2B-style long-lived sandboxes.

---

## Milestone 1: Foundation and Documentation
**Target:** Week 1
**Status:** In Progress

### Goals
- Document the restructuring plan
- Create directory structure
- Set up tracking infrastructure

### Issues
1. **Create restructuring documentation** - Document the overall plan and architecture
2. **Create milestone tracking** - Set up milestone and issue tracking
3. **Create directory structure placeholders** - Create new directories with README files
4. **Document current architecture** - Document the existing architecture for reference

### Success Criteria
- [ ] All documentation files created
- [ ] Directory structure initialized
- [ ] Team aligned on restructuring plan

---

## Milestone 2: VM Module Extraction
**Target:** Weeks 2-3
**Status:** Not Started

### Goals
- Extract all Firecracker-related code into `vm/` module
- Establish VM module as the single source of truth for VM operations
- No functional changes, pure refactoring

### Issues
5. **Create vm/manager module** - Move control-plane/vm/ to vm/manager/
6. **Create vm/pool module** - Extract VM pool logic into dedicated module
7. **Create vm/snapshot module** - Create snapshot management module
8. **Create vm/network module** - Extract network configuration logic
9. **Create vm/jailer module** - Create security isolation module
10. **Update all imports** - Update imports across codebase to use new vm/ module
11. **Add vm/ module tests** - Ensure test coverage for vm/ module
12. **Update documentation** - Document vm/ module API

### Success Criteria
- [ ] All Firecracker code in vm/ module
- [ ] No direct Firecracker dependencies outside vm/
- [ ] All tests passing
- [ ] Documentation updated

---

## Milestone 3: Runtime Formalization
**Target:** Weeks 4-5
**Status:** Not Started

### Goals
- Formalize the guest runtime as a first-class module
- Define clear host ↔ guest protocol
- Separate agent, Python runtime, and protocol

### Issues
13. **Create runtime/agent module** - Formalize Go agent as a module
14. **Create runtime/python module** - Extract Python runtime bootstrap
15. **Create runtime/protocol module** - Define host-guest communication protocol
16. **Document protocol specification** - Write comprehensive protocol docs
17. **Implement protocol versioning** - Add version negotiation support
18. **Add runtime tests** - Test host-guest communication
19. **Update agent deployment** - Update how agent is deployed to VMs

### Success Criteria
- [ ] Clear protocol definition
- [ ] Agent is a standalone module
- [ ] Protocol versioning in place
- [ ] All tests passing

---

## Milestone 4: Control Plane Reorganization
**Target:** Weeks 6-7
**Status:** Not Started

### Goals
- Move control plane to be orchestration-only
- Remove all VM and runtime logic
- Create thin entry points in cmd/

### Issues
20. **Create cmd/control-plane** - Move main.go to cmd/control-plane/main.go
21. **Remove VM logic from control-plane** - Control plane uses vm/ module only
22. **Refactor control-plane/api** - Clean up API handlers
23. **Update control-plane/scheduler** - Scheduler uses vm/ module
24. **Clean up control-plane imports** - Remove direct Firecracker dependencies
25. **Add control-plane tests** - Test orchestration logic
26. **Update build scripts** - Update Makefile and build scripts

### Success Criteria
- [ ] Control plane has no Firecracker dependencies
- [ ] Entry points in cmd/ are thin
- [ ] All tests passing
- [ ] Build scripts updated

---

## Milestone 5: Sandbox API Implementation
**Target:** Weeks 8-10
**Status:** Not Started

### Goals
- Implement E2B-style sandbox API
- Support long-lived VM sessions
- Implement behind feature flag

### Issues
27. **Design sandbox API** - Design REST API for sandbox operations
28. **Create sandbox/api module** - Implement API handlers
29. **Create sandbox/lifecycle module** - Implement create/suspend/resume/destroy
30. **Create sandbox/exec module** - Implement code execution in sandboxes
31. **Create sandbox/fs module** - Implement filesystem operations
32. **Add feature flag** - Implement feature flag for sandbox mode
33. **Implement sandbox state management** - Track sandbox lifecycle state
34. **Add sandbox tests** - Comprehensive testing for sandbox operations
35. **Document sandbox API** - API documentation and examples

### Success Criteria
- [ ] Sandbox API functional behind feature flag
- [ ] Can create and manage long-lived sandboxes
- [ ] File upload/download working
- [ ] Comprehensive tests
- [ ] API documented

---

## Milestone 6: SDK Development
**Target:** Weeks 11-12
**Status:** Not Started

### Goals
- Create Python SDK for both Lambda and Sandbox styles
- Provide excellent developer experience
- Include examples and documentation

### Issues
36. **Design SDK API** - Design Python SDK interface
37. **Create sdk/python module** - Implement Python SDK
38. **Implement Lambda-style API** - Support function invocation
39. **Implement Sandbox-style API** - Support sandbox operations
40. **Add SDK tests** - Test SDK functionality
41. **Create SDK examples** - Example code for both styles
42. **Document SDK** - Comprehensive SDK documentation
43. **Publish SDK** - Prepare for PyPI publication

### Success Criteria
- [ ] Python SDK supports both Lambda and Sandbox
- [ ] Excellent developer experience
- [ ] Examples and documentation complete
- [ ] Ready for release

---

## Milestone 7: Internal Utilities and Assets
**Target:** Weeks 13-14
**Status:** Not Started

### Goals
- Extract common utilities to internal/
- Organize assets properly
- Improve code reuse

### Issues
44. **Create internal/logging module** - Centralized logging utilities
45. **Create internal/config module** - Configuration management
46. **Create internal/errors module** - Error handling utilities
47. **Organize assets/ directory** - Move kernels, rootfs, snapshots
48. **Update asset references** - Update all references to assets
49. **Add internal tests** - Test internal utilities
50. **Document internal modules** - Document internal API

### Success Criteria
- [ ] Common utilities extracted
- [ ] Assets organized
- [ ] Code duplication reduced
- [ ] All tests passing

---

## Milestone 8: Testing and Documentation
**Target:** Weeks 15-16
**Status:** Not Started

### Goals
- Comprehensive testing of new architecture
- Update all documentation
- Performance benchmarking

### Issues
51. **Integration testing** - End-to-end tests for both Lambda and Sandbox
52. **Performance benchmarking** - Compare performance before/after
53. **Update README** - Update project README
54. **Update architecture docs** - Update architecture documentation
55. **Create migration guide** - Guide for users to migrate
56. **Update CONTRIBUTING** - Update contribution guidelines
57. **Code review** - Comprehensive code review
58. **Security audit** - Security review of new architecture

### Success Criteria
- [ ] All tests passing
- [ ] Performance meets or exceeds baseline
- [ ] Documentation complete
- [ ] Security reviewed

---

## Milestone 9: Release Preparation
**Target:** Week 17
**Status:** Not Started

### Goals
- Prepare for release
- Final testing and validation
- Communication and rollout

### Issues
59. **Final testing** - Comprehensive testing in production-like environment
60. **Update changelog** - Document all changes
61. **Create release notes** - Prepare release notes
62. **Update version numbers** - Bump version numbers
63. **Tag release** - Create git tags
64. **Deploy to staging** - Deploy to staging environment
65. **User acceptance testing** - Testing with real users
66. **Deploy to production** - Production deployment

### Success Criteria
- [ ] All testing complete
- [ ] Release notes prepared
- [ ] Deployed successfully
- [ ] No critical issues

---

## Timeline Overview

```
Week 1-1:    M1 - Foundation and Documentation
Week 2-3:    M2 - VM Module Extraction
Week 4-5:    M3 - Runtime Formalization
Week 6-7:    M4 - Control Plane Reorganization
Week 8-10:   M5 - Sandbox API Implementation
Week 11-12:  M6 - SDK Development
Week 13-14:  M7 - Internal Utilities and Assets
Week 15-16:  M8 - Testing and Documentation
Week 17:     M9 - Release Preparation
```

**Total Duration:** ~4 months

---

## Notes

- Milestones can overlap when dependencies allow
- Each milestone includes testing and documentation
- Feature flags protect experimental features
- Rollback plan exists for each milestone

---

## Progress Tracking

Track progress using:
- This document for milestone-level tracking
- [ISSUES.md](ISSUES.md) for issue-level tracking
- GitHub Issues for detailed task management
- GitHub Projects for visual progress tracking
