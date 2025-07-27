# Pull Request

## 📝 Description
Brief description of the changes and motivation behind them.

Fixes #(issue_number)

## 🔄 Type of Change
Please delete options that are not relevant.

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📚 Documentation update
- [ ] 🔧 Code refactoring (no functional changes)
- [ ] ⚡ Performance improvement
- [ ] 🧪 Test improvements
- [ ] 🔒 Security enhancement
- [ ] 🎨 UI/UX improvement
- [ ] 🚀 Build/CI improvement

## 🧪 Testing
Please describe the tests that you ran to verify your changes.

### Test Environment
- [ ] Docker environment
- [ ] Local development environment
- [ ] Production-like environment

### Test Types Completed
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] End-to-end tests added/updated
- [ ] Manual testing completed
- [ ] Performance testing (if applicable)
- [ ] Security testing (if applicable)

### Test Results
```bash
# Example test commands and results
go test ./...
# All tests passed

docker compose -f compose.test.yml up --abort-on-container-exit
# Integration tests passed
```

## 📋 Checklist
Please check all that apply:

### Code Quality
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings
- [ ] Code is properly formatted (`go fmt`, `goimports`)
- [ ] Linting passes (`golangci-lint run`)

### Testing & Validation
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Integration tests pass with my changes
- [ ] I have tested edge cases and error conditions

### Documentation
- [ ] I have made corresponding changes to the documentation
- [ ] API documentation updated (if applicable)
- [ ] README updated (if applicable)
- [ ] Code comments added where necessary
- [ ] Examples updated (if applicable)

### Security & Performance
- [ ] I have reviewed security implications of my changes
- [ ] Input validation added where necessary
- [ ] No sensitive information is exposed
- [ ] Performance impact evaluated
- [ ] Resource usage considered (memory, CPU, disk)

### Dependencies & Compatibility
- [ ] Any dependent changes have been merged and published
- [ ] No new dependencies added without justification
- [ ] Breaking changes documented and justified
- [ ] Backward compatibility maintained (or breaking changes documented)

## 🔍 Changes Made
### Files Modified
- `path/to/file1.go` - Description of changes
- `path/to/file2.go` - Description of changes
- `docs/README.md` - Updated documentation

### Key Changes
1. **Change 1**: Detailed description of what was changed and why
2. **Change 2**: Detailed description of what was changed and why
3. **Change 3**: Detailed description of what was changed and why

### Code Examples
**Before:**
```go
// Old implementation
func oldFunction() {
    // old code
}
```

**After:**
```go
// New implementation
func newFunction() error {
    // new improved code with error handling
    return nil
}
```

## 🚀 Deployment Considerations
### Environment Variables
```bash
# New environment variables (if any)
NEW_CONFIG_OPTION=value
FEATURE_ENABLED=true
```

### Database Changes
- [ ] No database changes
- [ ] New migrations added
- [ ] Backward compatible migrations
- [ ] Data migration required

### Configuration Changes
- [ ] No configuration changes
- [ ] New optional configuration
- [ ] New required configuration
- [ ] Breaking configuration changes

## 📊 Performance Impact
### Benchmarks (if applicable)
```bash
# Before changes
BenchmarkFunction-8    1000000    1000 ns/op

# After changes  
BenchmarkFunction-8    2000000     500 ns/op
```

### Resource Usage
- [ ] No significant impact
- [ ] Improved performance
- [ ] Increased memory usage (justified)
- [ ] Increased CPU usage (justified)

## 🔒 Security Considerations
- [ ] No security implications
- [ ] Input validation added
- [ ] Authentication/authorization updated
- [ ] Potential security improvements
- [ ] Security review required

### Security Checklist
- [ ] No hardcoded secrets or credentials
- [ ] Input sanitization implemented
- [ ] SQL injection prevention
- [ ] XSS prevention
- [ ] CSRF protection (if applicable)

## 📱 User Experience Impact
### API Changes
- [ ] No API changes
- [ ] New API endpoints
- [ ] Modified existing endpoints (backward compatible)
- [ ] Breaking API changes (documented)

### Error Handling
- [ ] Improved error messages
- [ ] Better error codes
- [ ] Enhanced logging
- [ ] User-friendly error responses

## 🔗 Related Issues and PRs
- Closes #(issue_number)
- Related to #(issue_number)
- Depends on #(pr_number)
- Blocks #(issue_number)

## 📸 Screenshots (if applicable)
**Before:**
[Add screenshot of before state]

**After:**
[Add screenshot of after state]

## 📚 Additional Documentation
- [Link to design document](example.com)
- [Link to API documentation](example.com)
- [Link to user guide](example.com)

## 🤔 Questions for Reviewers
1. Is the approach taken for X the best solution?
2. Should we consider Y alternative?
3. Any concerns about the performance impact?

## 🚨 Breaking Changes
If this is a breaking change, please describe:

### What breaks?
Describe what functionality will no longer work.

### Migration Guide
```bash
# How users should update their code/config
# Old way
curl -X POST "/old/endpoint"

# New way  
curl -X POST "/new/endpoint"
```

### Deprecation Timeline
- Version X.Y: Feature deprecated with warnings
- Version X.Z: Feature removed

---

## 📝 Review Notes
**For Reviewers:**
- [ ] Code review completed
- [ ] Architecture review completed
- [ ] Security review completed (if needed)
- [ ] Performance review completed (if needed)
- [ ] Documentation review completed

**Reviewer Checklist:**
- [ ] Code quality meets standards
- [ ] Tests are comprehensive
- [ ] Documentation is accurate
- [ ] No security vulnerabilities introduced
- [ ] Performance impact is acceptable

---

## 🎉 Post-Merge Actions
- [ ] Update changelog
- [ ] Announce in discussions (if significant feature)
- [ ] Update project roadmap
- [ ] Create follow-up issues (if needed)

**Thank you for contributing to FFprobe API! 🚀**