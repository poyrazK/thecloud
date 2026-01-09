window.BENCHMARK_DATA = {
  "lastUpdate": 1767990810167,
  "repoUrl": "https://github.com/PoyrazK/thecloud",
  "entries": {
    "Go Benchmarks": [
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "distinct": true,
          "id": "66677b5d976641a5e9eb591059f9944016e4a515",
          "message": "ci: add write permissions to benchmarks workflow",
          "timestamp": "2026-01-09T15:12:58+03:00",
          "tree_id": "fae73ef4a542441e13f73b3fe8b43bbb8a95e077",
          "url": "https://github.com/PoyrazK/thecloud/commit/66677b5d976641a5e9eb591059f9944016e4a515"
        },
        "date": 1767960802284,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceService_List",
            "value": 1.895,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639451438 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceService_List - ns/op",
            "value": 1.895,
            "unit": "ns/op",
            "extra": "639451438 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceService_List - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639451438 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceService_List - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639451438 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCService_Get",
            "value": 138.7,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9101067 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCService_Get - ns/op",
            "value": 138.7,
            "unit": "ns/op",
            "extra": "9101067 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCService_Get - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9101067 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCService_Get - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9101067 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "distinct": true,
          "id": "0bbf8279130caa7edf93beeb40843983c9a902fe",
          "message": "Merge branch 'fix/lint-issues': Major code quality improvements\n\n- Reduced cognitive complexity in InstanceService and LibvirtAdapter\n- Introduced parameter structs to reduce function parameter counts\n- Eliminated duplicate string literals across codebase\n- Improved security by addressing potential secret exposure in tests\n- All tests passing with no regressions",
          "timestamp": "2026-01-09T17:49:49+03:00",
          "tree_id": "26b7a92c928743ffa90c52dfabf4f1c14e7cb185",
          "url": "https://github.com/PoyrazK/thecloud/commit/0bbf8279130caa7edf93beeb40843983c9a902fe"
        },
        "date": 1767970227717,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.449,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "828889226 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.449,
            "unit": "ns/op",
            "extra": "828889226 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "828889226 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "828889226 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 126.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9403125 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 126.1,
            "unit": "ns/op",
            "extra": "9403125 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9403125 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9403125 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "distinct": true,
          "id": "ec4b81f55c22a27479bab40c5c666b8775badb93",
          "message": "fix(ci): start API server before running load tests\n\n- Add PostgreSQL service container for database\n- Build and start API server in background\n- Wait for health check to pass before running k6\n- Properly cleanup server process after tests\n- Fixes connection refused errors in load test workflow",
          "timestamp": "2026-01-09T17:52:58+03:00",
          "tree_id": "9ae91f74a73649766d838f3be980d9bb58f88f00",
          "url": "https://github.com/PoyrazK/thecloud/commit/ec4b81f55c22a27479bab40c5c666b8775badb93"
        },
        "date": 1767970399186,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642284600 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642284600 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642284600 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642284600 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.6,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8985362 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.6,
            "unit": "ns/op",
            "extra": "8985362 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8985362 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8985362 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "distinct": true,
          "id": "bca24ceea9ab3acff3ccb0b68b7b50c68d548622",
          "message": "fix(lint): resolve golangci-lint staticcheck and unused issues\n\n- Add package comment to setup package to satisfy ST1000\n- Remove unused constants bucketKeyRoute and roleIDRoute from cmd/api/main.go\n- All 5 lint issues resolved (3 staticcheck + 2 unused)",
          "timestamp": "2026-01-09T17:54:30+03:00",
          "tree_id": "64652a9094724a98ff93270a3880819d07948d1f",
          "url": "https://github.com/PoyrazK/thecloud/commit/bca24ceea9ab3acff3ccb0b68b7b50c68d548622"
        },
        "date": 1767970496850,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641201409 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641201409 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641201409 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641201409 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 131,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9077884 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 131,
            "unit": "ns/op",
            "extra": "9077884 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9077884 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9077884 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "distinct": true,
          "id": "05a197d3ebdbaccc097d43dbaf970e72e5a7516f",
          "message": "chore: remove sonarqube-mcp-server directory\n\n- Removed 297 files of SonarQube MCP server Java/Gradle project\n- This was a vendored/copied third-party project not needed for TheCloud\n- Reduces repository size and complexity\n- SonarQube analysis still works via GitHub Actions workflow",
          "timestamp": "2026-01-09T18:00:41+03:00",
          "tree_id": "e2b874edb40d58c70dffef16ab8c71fb799c9e20",
          "url": "https://github.com/PoyrazK/thecloud/commit/05a197d3ebdbaccc097d43dbaf970e72e5a7516f"
        },
        "date": 1767970885112,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642729123 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "642729123 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642729123 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642729123 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 132.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9271734 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 132.8,
            "unit": "ns/op",
            "extra": "9271734 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9271734 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9271734 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "name": "PoyrazK",
            "username": "PoyrazK"
          },
          "committer": {
            "name": "PoyrazK",
            "username": "PoyrazK"
          },
          "id": "73d9da0dde2e8533f17912305f72d7cd8fcb7ea4",
          "message": "chore(deps): bump actions/setup-go from 5 to 6",
          "timestamp": "2026-01-09T15:00:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/9/commits/73d9da0dde2e8533f17912305f72d7cd8fcb7ea4"
        },
        "date": 1767971360559,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640664216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "640664216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640664216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640664216 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 131,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9277029 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 131,
            "unit": "ns/op",
            "extra": "9277029 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9277029 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9277029 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "distinct": true,
          "id": "6cb460fa67f345c9d1e92324be7843f7c1aa930f",
          "message": "docs: update documentation with recent improvements\n\n- Add Code Quality, Load Testing, and SonarQube sections to CI/CD docs\n- Create comprehensive CHANGELOG.md documenting all recent changes\n- Add 'Recent Improvements' section to README highlighting code quality wins\n- Document refactoring efforts (parameter structs, complexity reduction)\n- Add references to new CI/CD workflows and tools\n- Update links and add changelog reference",
          "timestamp": "2026-01-09T18:10:24+03:00",
          "tree_id": "cb3523f1765b792c6f812d15aafcf4ee0b9557d7",
          "url": "https://github.com/PoyrazK/thecloud/commit/6cb460fa67f345c9d1e92324be7843f7c1aa930f"
        },
        "date": 1767971451597,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639050972 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "639050972 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639050972 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639050972 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 131.4,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9141504 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 131.4,
            "unit": "ns/op",
            "extra": "9141504 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9141504 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9141504 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "distinct": true,
          "id": "d8db6cf752c65474840994f4b17f000092bd1180",
          "message": "Merge feature/improve-test-coverage: Increase test coverage to 59.7%\n\nThis merge brings significant improvements to test coverage:\n\nTest Coverage Improvements:\n- Overall: 57.0% → 59.7% (+2.7%)\n- SDK: 44.2% → 80.1% (+35.9%)\n- Repositories: 65.0% → 70.1% (+5.1%)\n- Services: 71.5%\n- Handlers: 65.8%\n\nKey Achievements:\n✅ 16 new SDK test modules with comprehensive coverage\n✅ Fixed repository test issues and improved coverage\n✅ Added comprehensive TESTING.md documentation\n✅ Updated README.md and development.md with current stats\n✅ Fixed all golangci-lint issues (4 errors)\n✅ Resolved race condition in async function test\n✅ All tests passing with race detector\n\nCommits:\n- Repository refactoring and comprehensive unit tests\n- SDK tests for RBAC, Secret, Storage, Snapshot, Subnet, IAC\n- SDK tests for Security Group, Container, Database\n- SDK tests for Cron, Gateway, Events\n- Complete Autoscaling and LoadBalancer coverage\n- Documentation updates\n- Linting fixes (errcheck, ineffassign, staticcheck)\n- Race condition fix in async test",
          "timestamp": "2026-01-09T23:32:30+03:00",
          "tree_id": "c669cc51e38d9b429909997ac9a6a479d01770ca",
          "url": "https://github.com/PoyrazK/thecloud/commit/d8db6cf752c65474840994f4b17f000092bd1180"
        },
        "date": 1767990806014,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642333475 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642333475 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642333475 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642333475 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 128.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9194427 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 128.8,
            "unit": "ns/op",
            "extra": "9194427 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9194427 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9194427 times\n4 procs"
          }
        ]
      }
    ]
  }
}