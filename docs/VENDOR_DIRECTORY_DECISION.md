# Vendor Directory Decision Guide

## Current Situation

- **Vendor directory size**: 780+ KB
- **Dependencies managed by**: go.mod + go.sum (present)
- **Go version**: 1.25.0+
- **Dependency count**: 3 direct, ~15 indirect

## Decision Framework

### Option 1: Remove vendor/ (Recommended)

**When to choose this:**
- ✅ You have reliable internet access for builds
- ✅ You use standard Go toolchain (go mod)
- ✅ You want to follow modern Go practices (2024+)
- ✅ You want to reduce repository size
- ✅ Dependencies are stable and well-maintained

**Advantages:**
- Reduces repository size by ~780 KB
- Follows current Go best practices
- Simpler repository maintenance
- go.mod/go.sum provide same reproducibility
- Faster git operations

**How to Remove:**
```bash
# 1. Remove vendor directory
git rm -r vendor/

# 2. Update .gitignore
echo "vendor/" >> .gitignore

# 3. Verify builds still work
go mod download
go build ./cmd/github-mcp-server/

# 4. Commit
git add .gitignore
git commit -m "chore: Remove vendor directory, rely on go.mod"
```

**Build Changes:**
```bash
# Before (with vendor)
go build -mod=vendor ./cmd/github-mcp-server/

# After (without vendor)
go build ./cmd/github-mcp-server/
# OR
go build -mod=readonly ./cmd/github-mcp-server/
```

---

### Option 2: Keep vendor/ (Conservative)

**When to choose this:**
- ✅ You build in air-gapped/offline environments
- ✅ You need guaranteed reproducible builds without network
- ✅ You have corporate proxy/firewall restrictions
- ✅ You want absolute control over dependencies
- ✅ You need to audit all source code locally

**Advantages:**
- No network required for building
- Complete source code availability
- Guaranteed build reproducibility
- Useful for security audits
- Required for some enterprise environments

**How to Maintain:**
```bash
# Keep vendor up-to-date
go mod vendor

# Verify vendor is current
go mod verify

# Build using vendor
go build -mod=vendor ./cmd/github-mcp-server/
```

---

## Recommendation Matrix

| Use Case | Recommendation | Reason |
|----------|---------------|--------|
| Open source public repo | **Remove** | Standard practice, reduces size |
| Corporate/Enterprise | **Keep** | Offline builds, security audits |
| CI/CD pipeline | **Remove** | CI has internet access |
| Air-gapped deployment | **Keep** | No network available |
| Personal project | **Remove** | Simpler maintenance |
| Security-critical | **Keep** | Full source audit capability |

## Our Project's Context

**mcp-go-github characteristics:**
- ✅ Open source public repository
- ✅ Well-maintained dependencies (Google, Go team)
- ✅ Standard build environments (GitHub Actions ready)
- ✅ Modern Go project (1.25+)
- ❌ No air-gapped deployment requirement
- ❌ No special network restrictions

**Conclusion**: **REMOVE vendor/**

This project fits the "standard open source" profile where vendor/ provides minimal benefit and adds maintenance overhead.

## Migration Path (If Removing)

### Step 1: Preparation
```bash
# Verify current builds work
go build ./cmd/github-mcp-server/
go test ./...
```

### Step 2: Remove vendor
```bash
git rm -r vendor/
```

### Step 3: Update .gitignore
```bash
echo "vendor/" >> .gitignore
```

### Step 4: Update documentation
Update CLAUDE.md and README.md:
- Remove any references to vendor directory
- Update build instructions if they mention `-mod=vendor`

### Step 5: Test builds
```bash
# Clean build
go clean -cache
go mod download
go build ./cmd/github-mcp-server/

# Test
go test ./...
```

### Step 6: Update compile scripts
**compile.bat:**
```batch
REM Remove -mod=vendor flag if present
go build -ldflags="-s -w" -o mcp-go-github.exe ./cmd/github-mcp-server/
```

### Step 7: Commit
```bash
git add .gitignore
git commit -m "chore: Remove vendor directory, use go.mod exclusively

- Reduces repository size by 780KB
- Follows Go 1.25 best practices
- Dependencies managed via go.mod and go.sum
- No functional changes to build process"
```

## Post-Removal Verification

```bash
# Verify go.mod is complete
go mod verify

# Verify build works
go build ./cmd/github-mcp-server/

# Verify tests pass
go test ./...

# Check repository size reduction
git count-objects -vH
```

## Rollback Plan (If Issues Occur)

```bash
# Restore vendor directory
git revert <commit-hash>

# Or regenerate vendor
go mod vendor
git add vendor/
git commit -m "Restore vendor directory"
```

## Expected Results

### Before Removal
```
Repository size: ~1.2 MB
Clone time: ~3-5 seconds
Vendor directory: 780 KB
```

### After Removal
```
Repository size: ~420 KB
Clone time: ~1-2 seconds
Vendor directory: None (managed by Go)
Build time: Same (dependencies cached locally)
```

## Final Recommendation

**For mcp-go-github: REMOVE vendor/ directory**

Rationale:
1. Modern Go project (1.25+)
2. Public open source repository
3. Standard dependencies (Google, Go team)
4. No special build requirements
5. Following Go community best practices

The vendor/ directory adds 65% to repository size without providing meaningful benefits for this project's use case.
