# Test Suite for Bookmark Update Functionality

This test suite comprehensively covers the bookmark update functionality that intelligently routes between PATCH and PUT endpoints based on the type of update being performed.

## Test Files

### 1. `stores/__tests__/bookmarks.test.ts`
**Purpose:** Tests the Pinia store logic for bookmark updates

**Key Test Areas:**
- **Smart Routing Logic**: Verifies that the store correctly chooses between PATCH and PUT endpoints
  - PUT for content updates (title, URL, description)
  - PATCH for metadata updates (action, topic, shareTo)
  - Handles edge cases like `description: undefined`
- **Local State Management**: Ensures the store updates local state correctly after API calls
- **Error Handling**: Tests error scenarios and ensures state consistency
- **Integration with moveBookmarks**: Verifies bulk operations use correct endpoints

### 2. `services/__tests__/bookmarkService.test.ts`
**Purpose:** Tests the BookmarkService methods and data transformations

**Key Test Areas:**
- **API Endpoint Calls**: Verifies correct API calls to PATCH and PUT endpoints
- **Data Transformation**: Tests conversion between frontend and backend data formats
  - `toBackendUpdateRequest()` for PATCH data
  - `toBackendFullUpdateRequest()` for PUT data
  - `transformBackendBookmark()` for response parsing
- **Field Mapping**: Ensures proper handling of snake_case vs camelCase
- **Error Handling**: Tests network and validation error scenarios

### 3. `__tests__/bookmark-update-integration.test.ts`
**Purpose:** Integration tests covering the full update workflow

**Key Test Areas:**
- **Title Updates**: Full workflow for updating bookmark titles
- **Action Updates**: Workflow for moving bookmarks between states
- **Complex Scenarios**: Mixed content and metadata updates
- **URL Updates**: Handling URL changes correctly
- **Share Workflow**: Complete bookmark sharing process
- **Error Recovery**: Data consistency during failures
- **Performance**: Batching and efficiency considerations

## Test Scenarios Covered

### Content Updates (Use PUT Endpoint)
- ✅ Title changes
- ✅ URL changes  
- ✅ Description updates
- ✅ Description clearing (`undefined`)
- ✅ Mixed content + metadata updates

### Metadata Updates (Use PATCH Endpoint)
- ✅ Action changes (read-later → working → share → archived)
- ✅ Topic assignments
- ✅ Share destination changes
- ✅ Project ID assignments

### Edge Cases
- ✅ Empty update objects
- ✅ Null/undefined values
- ✅ Missing bookmark scenarios
- ✅ Bookmarks without optional fields
- ✅ No-op updates

### Error Scenarios
- ✅ Network timeouts
- ✅ Server errors
- ✅ Validation failures
- ✅ Data consistency during failures
- ✅ State rollback on errors

### Performance & Efficiency
- ✅ Single API call for multiple field updates
- ✅ Proper endpoint selection to minimize payload
- ✅ Batch operations for bulk updates

## Running Tests

```bash
# Run all tests once
npm run test:run

# Run tests in watch mode
npm run test

# Run tests with UI
npm run test:ui
```

## Test Architecture

The tests use:
- **Vitest** as the test runner
- **Vue Test Utils** for Vue component testing
- **Pinia** for state management testing
- **Mocked API client** to isolate backend dependencies
- **jsdom** environment for DOM simulation

## Key Testing Patterns

1. **Mock API Responses**: Each test mocks the expected backend response
2. **State Verification**: Tests verify both API calls and local state changes
3. **Error Isolation**: Failed API calls don't affect other tests
4. **Data Consistency**: All tests ensure data integrity throughout operations
5. **Real-world Scenarios**: Tests mirror actual user workflows

This comprehensive test suite ensures the bookmark update functionality works reliably across all supported scenarios and handles edge cases gracefully.