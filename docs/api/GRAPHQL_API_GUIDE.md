# GraphQL API Guide
## FFprobe API - Flexible Query Interface

## Overview

The FFprobe API provides a comprehensive GraphQL endpoint that offers flexible querying capabilities for video analysis, user management, and system administration. This guide covers everything you need to know about using the GraphQL API.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [Schema Overview](#schema-overview)
4. [Common Queries](#common-queries)
5. [Mutations](#mutations)
6. [Subscriptions](#subscriptions)
7. [Error Handling](#error-handling)
8. [Best Practices](#best-practices)
9. [Rate Limiting](#rate-limiting)
10. [Examples](#examples)

---

## Getting Started

### Endpoint

The GraphQL endpoint is available at:
```
POST /api/v1/graphql
GET  /api/v1/graphql  (for queries only)
```

### GraphQL Playground

In development mode, you can access the interactive GraphQL Playground at:
```
GET /api/v1/graphql/playground
```

### Schema Introspection

Get the complete schema definition:
```
GET /api/v1/graphql/schema
```

---

## Authentication

GraphQL endpoints support the same authentication methods as REST endpoints:

### API Key Authentication
```bash
curl -X POST https://api.example.com/api/v1/graphql \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { me { username email } }"}'
```

### JWT Authentication
```bash
curl -X POST https://api.example.com/api/v1/graphql \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { me { username email } }"}'
```

---

## Schema Overview

### Core Types

#### VideoAnalysis
```graphql
type VideoAnalysis {
  id: ID!
  filePath: String!
  fileName: String!
  duration: Float!
  bitrate: Int!
  format: VideoFormat!
  streams: [Stream!]!
  qualityMetrics: QualityMetrics
  contentAnalysis: ContentAnalysis
  status: AnalysisStatus!
  createdAt: DateTime!
}
```

#### QualityMetrics
```graphql
type QualityMetrics {
  id: ID!
  vmafScore: Float
  psnr: Float
  ssim: Float
  # ... 20+ quality metrics
}
```

#### User
```graphql
type User {
  id: ID!
  username: String!
  email: String!
  roles: [String!]!
  tenantId: String!
  apiKeys: [APIKey!]!
  rateLimits: UserRateLimit
}
```

---

## Common Queries

### Get Current User
```graphql
query GetCurrentUser {
  me {
    id
    username
    email
    roles
    tenantId
    rateLimits {
      perMinute
      perHour
      perDay
      monthlyQuota
      currentMonthUsage
    }
  }
}
```

### List Video Analyses
```graphql
query GetVideoAnalyses($filter: VideoAnalysisFilter, $pagination: PaginationInput) {
  videoAnalyses(filter: $filter, pagination: $pagination) {
    items {
      id
      fileName
      duration
      bitrate
      status
      createdAt
      qualityMetrics {
        vmafScore
        psnr
        ssim
      }
      contentAnalysis {
        blackFrames {
          timestamp
          duration
          percentage
        }
        freezeFrames {
          startTime
          endTime
          duration
        }
      }
    }
    totalCount
    hasNextPage
    hasPrevPage
  }
}
```

**Variables:**
```json
{
  "filter": {
    "status": ["COMPLETED"],
    "hasQualityMetrics": true,
    "createdAfter": "2024-01-01T00:00:00Z"
  },
  "pagination": {
    "page": 1,
    "limit": 20,
    "sortBy": "createdAt",
    "sortOrder": "DESC"
  }
}
```

### Get Specific Video Analysis
```graphql
query GetVideoAnalysis($id: ID!) {
  videoAnalysis(id: $id) {
    id
    filePath
    fileName
    fileSize
    duration
    bitrate
    format {
      name
      longName
      duration
      bitrate
      tags
    }
    streams {
      index
      codecName
      codecType
      width
      height
      bitrate
      framerate
      duration
    }
    qualityMetrics {
      vmafScore
      psnr
      ssim
      blockiness
      blur
      niqe
      brisque
    }
    contentAnalysis {
      blackFrames {
        timestamp
        duration
        percentage
      }
      freezeFrames {
        startTime
        endTime
        duration
      }
      silenceSegments {
        startTime
        endTime
        silenceThreshold
      }
      loudnessAnalysis {
        integratedLoudness
        loudnessRange
        maxTruePeak
        ebuR128Compliant
      }
    }
    hlsInfo {
      masterPlaylist
      variants {
        bandwidth
        resolution
        codecs
        frameRate
      }
      totalDuration
      segmentCount
    }
  }
}
```

### Search Video Analyses
```graphql
query SearchVideos($query: String!, $limit: Int) {
  searchVideoAnalyses(query: $query, limit: $limit) {
    id
    fileName
    duration
    format {
      name
    }
    status
    createdAt
  }
}
```

### Get Analytics Overview
```graphql
query GetAnalytics($tenantId: String, $startDate: DateTime, $endDate: DateTime) {
  analyticsOverview(
    tenantId: $tenantId
    startDate: $startDate
    endDate: $endDate
  ) {
    totalAnalyses
    analysesThisMonth
    averageProcessingTime
    totalUsersActive
    popularFormats {
      format
      count
      percentage
    }
    qualityDistribution {
      excellent
      good
      fair
      poor
    }
    usageByTenant {
      tenantId
      analysisCount
      storageUsed
      apiCallsThisMonth
    }
  }
}
```

---

## Mutations

### Create Video Analysis
```graphql
mutation CreateAnalysis($input: VideoAnalysisInput!) {
  createVideoAnalysis(input: $input) {
    id
    filePath
    fileName
    status
    createdAt
  }
}
```

**Variables:**
```json
{
  "input": {
    "filePath": "/path/to/video.mp4",
    "enableContentAnalysis": true,
    "enableQualityMetrics": true,
    "enableHLSAnalysis": false,
    "customParameters": {
      "vmaf_model": "4k",
      "quality_threshold": 0.8
    }
  }
}
```

### Create Video Comparison
```graphql
mutation CreateComparison($input: ComparisonInput!) {
  createVideoComparison(input: $input) {
    id
    referenceVideo {
      id
      fileName
    }
    testVideo {
      id
      fileName
    }
    status
    createdAt
  }
}
```

### Create API Key
```graphql
mutation CreateAPIKey($name: String!, $permissions: [String!]!) {
  createApiKey(name: $name, permissions: $permissions) {
    id
    name
    keyPrefix
    permissions
    expiresAt
    rotationDue
    rateLimits {
      perMinute
      perHour
      perDay
    }
  }
}
```

### Rotate API Key
```graphql
mutation RotateKey($id: ID!) {
  rotateApiKey(id: $id) {
    id
    keyPrefix
    expiresAt
    rotationDue
  }
}
```

### Generate Analysis Report
```graphql
mutation GenerateReport($input: ReportGenerationInput!) {
  generateAnalysisReport(input: $input) {
    id
    reportType
    format
    generatedAt
    llmGenerated
    filePath
  }
}
```

**Variables:**
```json
{
  "input": {
    "videoAnalysisId": "video-123",
    "reportType": "QUALITY_ASSESSMENT",
    "format": "PDF",
    "includeGraphs": true
  }
}
```

### Admin Mutations

#### Update Rate Limits
```graphql
mutation UpdateUserLimits($userId: ID!, $perMinute: Int!, $perHour: Int!, $perDay: Int!) {
  updateUserRateLimits(
    userId: $userId
    perMinute: $perMinute
    perHour: $perHour
    perDay: $perDay
  ) {
    userId
    perMinute
    perHour
    perDay
    burstMultiplier
  }
}
```

#### Rotate JWT Secret
```graphql
mutation RotateJWT {
  rotateJwtSecret
}
```

---

## Subscriptions

### Real-time Analysis Progress
```graphql
subscription AnalysisProgress($id: ID!) {
  videoAnalysisProgress(id: $id) {
    id
    status
    processingTime
    # ... other fields update in real-time
  }
}
```

### System Notifications
```graphql
subscription Notifications {
  userNotifications
}
```

### System Status Updates
```graphql
subscription SystemStatus {
  systemStatus
}
```

---

## Error Handling

GraphQL returns structured error responses:

```json
{
  "data": null,
  "errors": [
    {
      "message": "Authentication required",
      "locations": [{"line": 2, "column": 3}],
      "path": ["me"],
      "extensions": {
        "code": "UNAUTHENTICATED",
        "timestamp": "2024-02-09T10:00:00Z"
      }
    }
  ]
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `UNAUTHENTICATED` | Authentication required |
| `FORBIDDEN` | Insufficient permissions |
| `BAD_USER_INPUT` | Invalid input parameters |
| `INTERNAL_ERROR` | Server error |
| `NOT_FOUND` | Resource not found |
| `RATE_LIMITED` | Rate limit exceeded |

---

## Rate Limiting

GraphQL queries are subject to rate limiting based on:

1. **Query Complexity**: Expensive queries consume more quota
2. **Request Count**: Standard per-minute/hour/day limits
3. **User/Tenant Limits**: Custom limits per user or tenant

### Rate Limit Headers
```
X-RateLimit-Limit-Minute: 60
X-RateLimit-Remaining-Minute: 45
X-RateLimit-Reset-Minute: 1707475260
```

### Managing Query Complexity

Use pagination and field selection to reduce complexity:

```graphql
# ❌ High complexity
query ExpensiveQuery {
  videoAnalyses {
    items {
      streams {
        # All stream fields
      }
      qualityMetrics {
        # All quality metrics
      }
    }
  }
}

# ✅ Lower complexity
query OptimizedQuery {
  videoAnalyses(pagination: { limit: 10 }) {
    items {
      id
      fileName
      status
      # Only needed fields
    }
  }
}
```

---

## Best Practices

### 1. Use Field Selection
Only request fields you need:
```graphql
query GetVideos {
  videoAnalyses {
    items {
      id          # ✅ Always include ID
      fileName    # ✅ Only what you need
      status      # ✅ 
      # duration  # ❌ Skip unused fields
    }
  }
}
```

### 2. Implement Pagination
Always use pagination for lists:
```graphql
query GetVideosPaginated {
  videoAnalyses(
    pagination: { 
      page: 1, 
      limit: 20, 
      sortBy: "createdAt", 
      sortOrder: DESC 
    }
  ) {
    items { id fileName }
    totalCount
    hasNextPage
  }
}
```

### 3. Use Variables
Parameterize your queries:
```graphql
query GetUserVideos($userId: ID!, $limit: Int = 10) {
  videoAnalyses(
    filter: { userId: $userId }
    pagination: { limit: $limit }
  ) {
    items { id fileName }
  }
}
```

### 4. Handle Errors Gracefully
```javascript
const result = await client.query({
  query: GET_VIDEOS,
  errorPolicy: 'partial' // Return partial data on errors
});

if (result.errors) {
  console.error('GraphQL errors:', result.errors);
}
```

### 5. Use Fragments for Reusability
```graphql
fragment VideoBasics on VideoAnalysis {
  id
  fileName
  duration
  status
  createdAt
}

query GetVideos {
  videoAnalyses {
    items {
      ...VideoBasics
    }
  }
}
```

---

## Examples

### JavaScript/TypeScript Client

```typescript
import { ApolloClient, InMemoryCache, gql } from '@apollo/client';

const client = new ApolloClient({
  uri: 'https://api.example.com/api/v1/graphql',
  cache: new InMemoryCache(),
  headers: {
    'X-API-Key': 'your-api-key'
  }
});

// Query
const GET_VIDEOS = gql`
  query GetVideos($filter: VideoAnalysisFilter) {
    videoAnalyses(filter: $filter) {
      items {
        id
        fileName
        duration
        qualityMetrics {
          vmafScore
        }
      }
    }
  }
`;

const { data } = await client.query({
  query: GET_VIDEOS,
  variables: {
    filter: { status: ['COMPLETED'] }
  }
});

// Mutation
const CREATE_ANALYSIS = gql`
  mutation CreateAnalysis($input: VideoAnalysisInput!) {
    createVideoAnalysis(input: $input) {
      id
      status
    }
  }
`;

const { data } = await client.mutate({
  mutation: CREATE_ANALYSIS,
  variables: {
    input: {
      filePath: '/path/to/video.mp4',
      enableQualityMetrics: true
    }
  }
});

// Subscription
const ANALYSIS_PROGRESS = gql`
  subscription AnalysisProgress($id: ID!) {
    videoAnalysisProgress(id: $id) {
      id
      status
      processingTime
    }
  }
`;

client.subscribe({
  query: ANALYSIS_PROGRESS,
  variables: { id: 'analysis-123' }
}).subscribe({
  next: (result) => console.log('Progress:', result.data),
  error: (err) => console.error('Error:', err)
});
```

### Python Client

```python
import requests
import json

class FFprobeGraphQLClient:
    def __init__(self, url, api_key):
        self.url = url
        self.headers = {
            'Content-Type': 'application/json',
            'X-API-Key': api_key
        }
    
    def query(self, query, variables=None):
        payload = {
            'query': query,
            'variables': variables or {}
        }
        
        response = requests.post(
            self.url, 
            headers=self.headers,
            json=payload
        )
        
        return response.json()

# Usage
client = FFprobeGraphQLClient(
    'https://api.example.com/api/v1/graphql',
    'your-api-key'
)

# Get videos
query = '''
query GetVideos($limit: Int!) {
  videoAnalyses(pagination: { limit: $limit }) {
    items {
      id
      fileName
      duration
      qualityMetrics {
        vmafScore
        psnr
      }
    }
  }
}
'''

result = client.query(query, {'limit': 10})
videos = result['data']['videoAnalyses']['items']

# Create analysis
mutation = '''
mutation CreateAnalysis($input: VideoAnalysisInput!) {
  createVideoAnalysis(input: $input) {
    id
    status
  }
}
'''

result = client.query(mutation, {
    'input': {
        'filePath': '/path/to/video.mp4',
        'enableQualityMetrics': True,
        'enableContentAnalysis': True
    }
})
```

### curl Examples

```bash
# Get current user
curl -X POST https://api.example.com/api/v1/graphql \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { me { id username email roles } }"
  }'

# Get video analyses with filters
curl -X POST https://api.example.com/api/v1/graphql \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query GetVideos($filter: VideoAnalysisFilter) { videoAnalyses(filter: $filter) { items { id fileName status qualityMetrics { vmafScore } } } }",
    "variables": {
      "filter": {
        "status": ["COMPLETED"],
        "hasQualityMetrics": true
      }
    }
  }'

# Create new analysis
curl -X POST https://api.example.com/api/v1/graphql \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation CreateAnalysis($input: VideoAnalysisInput!) { createVideoAnalysis(input: $input) { id status } }",
    "variables": {
      "input": {
        "filePath": "/path/to/video.mp4",
        "enableQualityMetrics": true,
        "enableContentAnalysis": true
      }
    }
  }'
```

---

## Advanced Features

### Query Complexity Analysis

The GraphQL server analyzes query complexity to prevent expensive operations:

```graphql
# This query has high complexity due to nested relationships
query ExpensiveQuery {
  videoAnalyses(pagination: { limit: 100 }) {  # +200 complexity
    items {
      streams {                                 # +100 complexity per video
        # ... many fields
      }
      qualityMetrics {                          # +50 complexity per video
        # ... all metrics
      }
    }
  }
}
# Total complexity: 35,000 (likely rejected)
```

### Automatic Persisted Queries (APQ)

For better performance, use APQ to cache queries:

```javascript
// First request: send full query + hash
const query = gql`query GetVideos { ... }`;
const queryHash = sha256(print(query));

// Subsequent requests: send only hash
fetch('/api/v1/graphql', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    extensions: {
      persistedQuery: {
        version: 1,
        sha256Hash: queryHash
      }
    }
  })
});
```

### Custom Directives

Use custom directives for advanced features:

```graphql
query GetVideos {
  videoAnalyses @auth(roles: ["admin"]) {
    items @rateLimit(maxRequests: 100, window: "1h") {
      id
      fileName
    }
  }
}
```

---

## Monitoring & Debugging

### Enable Query Logging

Set `LOG_LEVEL=debug` to enable detailed query logging:

```json
{
  "level": "info",
  "timestamp": "2024-02-09T10:00:00Z",
  "event": "graphql_query",
  "operation": "GetVideos",
  "complexity": 150,
  "execution_time_ms": 45,
  "user_id": "user_123"
}
```

### Query Performance

Monitor query performance in GraphQL Playground or via logging:

- **Execution Time**: < 1s for simple queries
- **Complexity Score**: < 1000 recommended
- **Database Queries**: Minimize N+1 queries

---

## Migration from REST

### REST to GraphQL Equivalents

| REST Endpoint | GraphQL Query |
|--------------|---------------|
| `GET /api/v1/probe/:id` | `query { videoAnalysis(id: $id) { ... } }` |
| `GET /api/v1/probe?limit=10` | `query { videoAnalyses(pagination: { limit: 10 }) { ... } }` |
| `POST /api/v1/probe/file` | `mutation { createVideoAnalysis(input: $input) { ... } }` |
| `GET /api/v1/auth/profile` | `query { me { ... } }` |

### Benefits of GraphQL

1. **Single Endpoint**: No more endpoint sprawl
2. **Flexible Queries**: Get exactly what you need
3. **Strong Typing**: Built-in validation and introspection
4. **Real-time**: Native subscription support
5. **Tools**: Rich ecosystem of clients and tools

---

## Troubleshooting

### Common Issues

#### 1. Authentication Errors
```json
{
  "errors": [{"message": "Authentication required"}]
}
```
**Solution**: Ensure API key or JWT token is included in headers.

#### 2. Rate Limit Exceeded
```json
{
  "errors": [{"message": "Rate limit exceeded: 60 requests per minute"}]
}
```
**Solution**: Implement exponential backoff and reduce query complexity.

#### 3. Query Too Complex
```json
{
  "errors": [{"message": "Query complexity limit exceeded"}]
}
```
**Solution**: Use pagination, reduce field selection, or split into multiple queries.

#### 4. Field Not Found
```json
{
  "errors": [{"message": "Cannot query field 'unknownField' on type 'VideoAnalysis'"}]
}
```
**Solution**: Check schema documentation or use introspection.

### Debug Commands

```bash
# Get schema
curl https://api.example.com/api/v1/graphql/schema

# Introspection query
curl -X POST https://api.example.com/api/v1/graphql \
  -d '{"query": "{ __schema { types { name } } }"}'

# Check field availability
curl -X POST https://api.example.com/api/v1/graphql \
  -d '{"query": "{ __type(name: \"VideoAnalysis\") { fields { name type { name } } } }"}'
```

---

## Support

For GraphQL API issues:

1. Check the [Schema Documentation](#schema-overview)
2. Use GraphQL Playground for testing
3. Enable debug logging for detailed errors
4. Contact API support with query examples

---

*Last Updated: 2024-02-09*  
*Version: 1.0.0*  
*GraphQL Specification: [June 2018](https://spec.graphql.org/)*