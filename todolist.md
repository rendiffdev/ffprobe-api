# FFprobe API - Action Items

Generated from validated Gemini audit findings on 2025-07-31

## 🚨 High Priority

### 1. Fix User Role System Discrepancy
**Issue**: Database schema defines roles as `('admin', 'user', 'viewer', 'guest')` but middleware and documentation reference different roles.

**Actions**:
- [ ] Decide on final role set: Consider `('admin', 'user', 'pro', 'premium', 'viewer', 'guest')`
- [ ] Create migration to update `user_role` enum in database schema
- [ ] Update middleware auth checks to match database roles
- [ ] Implement role-based rate limiting for pro/premium tiers
- [ ] Add API endpoints for role management:
  - `GET /api/v1/admin/users/:id/role`
  - `PUT /api/v1/admin/users/:id/role`
- [ ] Update documentation to reflect correct roles

### 2. Implement Custom VMAF Models Support
**Issue**: Database table `vmaf_models` exists but no business logic to use custom models.

**Actions**:
- [ ] Implement VMAF model repository in `internal/repositories/vmaf_models.go`
- [ ] Add service layer for VMAF model management
- [ ] Update `internal/quality/analyzer.go` to:
  - Load custom models from database
  - Allow model selection in quality analysis requests
  - Validate model files before storage
- [ ] Create API endpoints:
  - `POST /api/v1/quality/models` - Upload custom VMAF model
  - `GET /api/v1/quality/models` - List available models
  - `DELETE /api/v1/quality/models/:id` - Remove custom model
- [ ] Add model validation and security checks
- [ ] Update quality analysis endpoints to accept `model_id` parameter

## 📈 Medium Priority

### 3. Enhance Real-time Progress Implementation
**Issue**: Current SSE implementation uses simulated progress data.

**Actions**:
- [ ] Replace mock progress data in `StreamAnalysis` with real ffprobe progress
- [ ] Implement progress callback in ffprobe command execution
- [ ] Add progress reporting to:
  - Video quality analysis
  - HLS manifest parsing
  - Batch processing operations
- [ ] Create progress channel in AnalysisService for real-time updates
- [ ] Add WebSocket reconnection logic for resilience

### 4. Complete Enterprise Worker Features
**Issue**: Worker services use some mock data for demonstration.

**Actions**:
- [ ] Replace mock responses in ffprobe-worker with actual ffprobe execution
- [ ] Implement distributed task queue using RabbitMQ
- [ ] Add worker health monitoring and auto-scaling logic
- [ ] Implement circuit breaker pattern for worker failures
- [ ] Add worker performance metrics to Prometheus

## 🔧 Low Priority

### 5. Testing Improvements
**From audit recommendations**:
- [ ] Add edge case tests for ffmpeg package
- [ ] Add edge case tests for quality analyzer
- [ ] Add edge case tests for HLS parser
- [ ] Expand E2E test coverage in `internal/workflows/e2e_tester.go`
- [ ] Add integration tests for worker services

### 6. Caching Strategy Enhancement
**From audit recommendations**:
- [ ] Implement tiered caching (in-memory + Redis)
- [ ] Add cache invalidation mechanisms
- [ ] Implement cache warming for frequently accessed data
- [ ] Add cache hit/miss metrics

### 7. Documentation Enhancements
- [ ] Create detailed tutorial for custom VMAF models
- [ ] Add enterprise deployment guide
- [ ] Document worker service architecture
- [ ] Create performance tuning guide
- [ ] Add troubleshooting guide for common issues

## ✅ Completed/No Action Needed

- ~~OpenAPI Specification~~ - Already exists and is comprehensive
- ~~Worker Services Implementation~~ - Already implemented
- ~~WebSocket/SSE Implementation~~ - Already functional
- ~~Enterprise Architecture~~ - Already properly structured

## 📊 Success Metrics

- All database roles match middleware implementation
- Custom VMAF models can be uploaded and used in analysis
- Real-time progress shows actual ffprobe execution progress
- Zero mock data in production worker services
- 90%+ test coverage for edge cases
- Cache hit rate > 80% for repeated analyses

## 🗓️ Suggested Timeline

- **Week 1-2**: Fix user role discrepancy (High priority)
- **Week 3-4**: Implement custom VMAF models support
- **Week 5-6**: Enhance real-time progress and worker features
- **Week 7-8**: Testing improvements and documentation
- **Ongoing**: Performance optimization and caching improvements

---

*Note: This todolist was generated after validating the Gemini audit findings. Several audit concerns were found to be incorrect after examining the actual codebase.*