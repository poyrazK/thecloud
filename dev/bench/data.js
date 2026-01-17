window.BENCHMARK_DATA = {
  "lastUpdate": 1768661022489,
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
          "id": "21eb901e1068616b85240461218fc5b3b06921de",
          "message": "chore(deps): bump schemathesis/action from 1 to 2",
          "timestamp": "2026-01-09T20:32:47Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/17/commits/21eb901e1068616b85240461218fc5b3b06921de"
        },
        "date": 1768123629363,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642471482 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "642471482 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642471482 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642471482 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 127.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9371042 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 127.1,
            "unit": "ns/op",
            "extra": "9371042 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9371042 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9371042 times\n4 procs"
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
          "id": "2d040c444f8a9840506a3983e2af6ae659c58568",
          "message": "chore(deps): bump sonarsource/sonarqube-scan-action from 6 to 7",
          "timestamp": "2026-01-09T20:32:47Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/16/commits/2d040c444f8a9840506a3983e2af6ae659c58568"
        },
        "date": 1768123632166,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.877,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641778633 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.877,
            "unit": "ns/op",
            "extra": "641778633 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641778633 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641778633 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 128.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9262960 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 128.1,
            "unit": "ns/op",
            "extra": "9262960 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9262960 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9262960 times\n4 procs"
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
          "id": "2295f3bae9804e164217e092877de415cbfc55df",
          "message": "fix(test): update service tests to match new backend architecture",
          "timestamp": "2026-01-11T14:38:11+03:00",
          "tree_id": "6beaecf197ff08b6971a8f20967186e5d1a28b0c",
          "url": "https://github.com/PoyrazK/thecloud/commit/2295f3bae9804e164217e092877de415cbfc55df"
        },
        "date": 1768131561145,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642167769 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "642167769 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642167769 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642167769 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 129,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9332886 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 129,
            "unit": "ns/op",
            "extra": "9332886 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9332886 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9332886 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 397.6,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "3038193 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 397.6,
            "unit": "ns/op",
            "extra": "3038193 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "3038193 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3038193 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54847,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "30759 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54847,
            "unit": "ns/op",
            "extra": "30759 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "30759 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "30759 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 219.1,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5882216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 219.1,
            "unit": "ns/op",
            "extra": "5882216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5882216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5882216 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 183.5,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5462811 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 183.5,
            "unit": "ns/op",
            "extra": "5462811 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5462811 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5462811 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.885,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642431590 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.885,
            "unit": "ns/op",
            "extra": "642431590 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642431590 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642431590 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 103.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11750553 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 103.7,
            "unit": "ns/op",
            "extra": "11750553 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11750553 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11750553 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.831,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "207980216 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.831,
            "unit": "ns/op",
            "extra": "207980216 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "207980216 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "207980216 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "638782935 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "638782935 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "638782935 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "638782935 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.49,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28846948 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.49,
            "unit": "ns/op",
            "extra": "28846948 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28846948 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28846948 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66368756,
            "unit": "ns/op\t    6420 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66368756,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6420,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66368860,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66368860,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "b17ed7c689feece314f504809dc2fe67c84ce126",
          "message": "fix(test): update setup_test to match InitNetworkBackend signature",
          "timestamp": "2026-01-11T14:41:13+03:00",
          "tree_id": "9c610093dade9f32893e719a9b377c634658cc6b",
          "url": "https://github.com/PoyrazK/thecloud/commit/b17ed7c689feece314f504809dc2fe67c84ce126"
        },
        "date": 1768131744381,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641976818 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641976818 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641976818 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641976818 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 132.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9459202 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 132.3,
            "unit": "ns/op",
            "extra": "9459202 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9459202 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9459202 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 392.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "3031705 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 392.8,
            "unit": "ns/op",
            "extra": "3031705 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "3031705 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3031705 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 49829,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "32306 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 49829,
            "unit": "ns/op",
            "extra": "32306 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "32306 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "32306 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 201.3,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5928009 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 201.3,
            "unit": "ns/op",
            "extra": "5928009 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5928009 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5928009 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 176.8,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "6486400 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 176.8,
            "unit": "ns/op",
            "extra": "6486400 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "6486400 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "6486400 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642786879 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "642786879 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642786879 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642786879 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 99.43,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12212928 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 99.43,
            "unit": "ns/op",
            "extra": "12212928 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12212928 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12212928 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.997,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203024440 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.997,
            "unit": "ns/op",
            "extra": "203024440 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203024440 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203024440 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "628360490 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "628360490 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "628360490 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "628360490 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 39.94,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28843850 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 39.94,
            "unit": "ns/op",
            "extra": "28843850 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28843850 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28843850 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66393337,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66393337,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66441947,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66441947,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "de5b0c2c581823aafd690c039f315abbca7ded20",
          "message": "fix(test): resolve data race in snapshot and fix test expectations",
          "timestamp": "2026-01-11T14:50:30+03:00",
          "tree_id": "8df39c729be8ded756eaead9461fb20209a47580",
          "url": "https://github.com/PoyrazK/thecloud/commit/de5b0c2c581823aafd690c039f315abbca7ded20"
        },
        "date": 1768132299833,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "638916325 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "638916325 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "638916325 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "638916325 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 127.6,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9319579 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 127.6,
            "unit": "ns/op",
            "extra": "9319579 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9319579 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9319579 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 398.9,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "3024810 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 398.9,
            "unit": "ns/op",
            "extra": "3024810 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "3024810 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3024810 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 47740,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "32527 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 47740,
            "unit": "ns/op",
            "extra": "32527 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "32527 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "32527 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 199.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "6044517 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 199.8,
            "unit": "ns/op",
            "extra": "6044517 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "6044517 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "6044517 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 180.2,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5947496 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 180.2,
            "unit": "ns/op",
            "extra": "5947496 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5947496 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5947496 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642875046 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "642875046 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642875046 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642875046 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 99.99,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12031560 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 99.99,
            "unit": "ns/op",
            "extra": "12031560 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12031560 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12031560 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.906,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203313243 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.906,
            "unit": "ns/op",
            "extra": "203313243 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203313243 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203313243 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.878,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641056248 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.878,
            "unit": "ns/op",
            "extra": "641056248 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641056248 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641056248 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.07,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "29045673 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.07,
            "unit": "ns/op",
            "extra": "29045673 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "29045673 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "29045673 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66316793,
            "unit": "ns/op\t    6420 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66316793,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6420,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66678801,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66678801,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "b6f97dc30bcb8b7604835fa8a94b00da24169533",
          "message": "fix(lint): address staticcheck and errcheck issues",
          "timestamp": "2026-01-11T14:52:43+03:00",
          "tree_id": "759ec4e36a1103ca02d8d9cc001593fac53a8727",
          "url": "https://github.com/PoyrazK/thecloud/commit/b6f97dc30bcb8b7604835fa8a94b00da24169533"
        },
        "date": 1768132438516,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.873,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641730309 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.873,
            "unit": "ns/op",
            "extra": "641730309 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641730309 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641730309 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 138.5,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9156618 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 138.5,
            "unit": "ns/op",
            "extra": "9156618 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9156618 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9156618 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 397,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "3010710 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 397,
            "unit": "ns/op",
            "extra": "3010710 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "3010710 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3010710 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 65566,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "32625 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 65566,
            "unit": "ns/op",
            "extra": "32625 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "32625 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "32625 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 211,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5691513 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 211,
            "unit": "ns/op",
            "extra": "5691513 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5691513 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5691513 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 180.3,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5580210 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 180.3,
            "unit": "ns/op",
            "extra": "5580210 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5580210 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5580210 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.887,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "636478821 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.887,
            "unit": "ns/op",
            "extra": "636478821 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "636478821 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "636478821 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 102.1,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11611756 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 102.1,
            "unit": "ns/op",
            "extra": "11611756 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11611756 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11611756 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.901,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "205398918 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.901,
            "unit": "ns/op",
            "extra": "205398918 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "205398918 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "205398918 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.88,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "637744560 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.88,
            "unit": "ns/op",
            "extra": "637744560 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "637744560 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "637744560 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.84,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28529727 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.84,
            "unit": "ns/op",
            "extra": "28529727 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28529727 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28529727 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66432288,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66432288,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66407535,
            "unit": "ns/op\t    5592 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66407535,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5592,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "d9f14ae1d24c4dd72f29c33d148ebca57990aa15",
          "message": "fix(lint): add package comment to lvm adapter",
          "timestamp": "2026-01-11T14:55:37+03:00",
          "tree_id": "60d3a6a79e737a3c7d7773e5abd08bda25be7984",
          "url": "https://github.com/PoyrazK/thecloud/commit/d9f14ae1d24c4dd72f29c33d148ebca57990aa15"
        },
        "date": 1768132607418,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642342631 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "642342631 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642342631 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642342631 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 127.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9351871 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 127.3,
            "unit": "ns/op",
            "extra": "9351871 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9351871 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9351871 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 400.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "3038205 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 400.2,
            "unit": "ns/op",
            "extra": "3038205 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "3038205 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3038205 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 48271,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "32329 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 48271,
            "unit": "ns/op",
            "extra": "32329 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "32329 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "32329 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 202.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5923608 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 202.7,
            "unit": "ns/op",
            "extra": "5923608 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5923608 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5923608 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 177.8,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5928133 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 177.8,
            "unit": "ns/op",
            "extra": "5928133 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5928133 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5928133 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643367636 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "643367636 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643367636 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643367636 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 99.64,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11961228 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 99.64,
            "unit": "ns/op",
            "extra": "11961228 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11961228 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11961228 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.912,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202600941 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.912,
            "unit": "ns/op",
            "extra": "202600941 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202600941 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202600941 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.884,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642654842 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.884,
            "unit": "ns/op",
            "extra": "642654842 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642654842 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642654842 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.07,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28953231 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.07,
            "unit": "ns/op",
            "extra": "28953231 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28953231 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28953231 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66835811,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66835811,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66411914,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66411914,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "f2a1d820aea47f48fc77154bd3420bd00d21a88c",
          "message": "fix(services): inject logger into AccountingService",
          "timestamp": "2026-01-11T14:59:36+03:00",
          "tree_id": "8c4f5509c7f386068a3ff88e2c7170746af5d8ba",
          "url": "https://github.com/PoyrazK/thecloud/commit/f2a1d820aea47f48fc77154bd3420bd00d21a88c"
        },
        "date": 1768132848761,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641388217 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641388217 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641388217 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641388217 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 133.4,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9146529 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 133.4,
            "unit": "ns/op",
            "extra": "9146529 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9146529 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9146529 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 406.1,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2681796 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 406.1,
            "unit": "ns/op",
            "extra": "2681796 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2681796 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2681796 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 45793,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "32096 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 45793,
            "unit": "ns/op",
            "extra": "32096 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "32096 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "32096 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 236.4,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "4677058 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 236.4,
            "unit": "ns/op",
            "extra": "4677058 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "4677058 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "4677058 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 178.2,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5814385 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 178.2,
            "unit": "ns/op",
            "extra": "5814385 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5814385 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5814385 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641435806 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "641435806 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641435806 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641435806 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 99.87,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12332937 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 99.87,
            "unit": "ns/op",
            "extra": "12332937 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12332937 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12332937 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.919,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "201858585 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.919,
            "unit": "ns/op",
            "extra": "201858585 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "201858585 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "201858585 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640895000 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "640895000 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640895000 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640895000 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.47,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27812930 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.47,
            "unit": "ns/op",
            "extra": "27812930 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27812930 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27812930 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66324369,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66324369,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66358230,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66358230,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "a1889081082dfb38c9925e0b26416c1a5f1d951d",
          "message": "refactor(lint): address SonarQube issues - reduce complexity and duplication\n\n- Refactor InitServices to use ServiceConfig struct (reduces parameter count)\n- Extract worker execution to runWorkers helper function\n- Add constants for duplicate string literals in instance_repo and snapshot service\n- Refactor dashboard_test mock to reduce code duplication\n- All changes maintain backward compatibility and pass tests",
          "timestamp": "2026-01-11T15:05:00+03:00",
          "tree_id": "455762fea987fa306560c48e5af14dcdbb795a55",
          "url": "https://github.com/PoyrazK/thecloud/commit/a1889081082dfb38c9925e0b26416c1a5f1d951d"
        },
        "date": 1768133173853,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.885,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "635501880 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.885,
            "unit": "ns/op",
            "extra": "635501880 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "635501880 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "635501880 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 137.4,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9220750 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 137.4,
            "unit": "ns/op",
            "extra": "9220750 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9220750 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9220750 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 409.4,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2904661 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 409.4,
            "unit": "ns/op",
            "extra": "2904661 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2904661 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2904661 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 45865,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "34044 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 45865,
            "unit": "ns/op",
            "extra": "34044 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "34044 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "34044 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 234.6,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5886157 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 234.6,
            "unit": "ns/op",
            "extra": "5886157 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5886157 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5886157 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 181.3,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5785465 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 181.3,
            "unit": "ns/op",
            "extra": "5785465 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5785465 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5785465 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "638085165 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "638085165 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "638085165 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "638085165 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 101.8,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12252620 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 101.8,
            "unit": "ns/op",
            "extra": "12252620 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12252620 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12252620 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.911,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202980092 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.911,
            "unit": "ns/op",
            "extra": "202980092 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202980092 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202980092 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641601440 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641601440 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641601440 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641601440 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.09,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28984898 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.09,
            "unit": "ns/op",
            "extra": "28984898 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28984898 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28984898 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66549763,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66549763,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66351744,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66351744,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "9c2dc134ddfe3e50e902ffa79d5e7a2def70b410",
          "message": "fix(cmd): allow migrations to run without redis connection",
          "timestamp": "2026-01-11T20:16:12+03:00",
          "tree_id": "73002e70aa14adc590a9e47b557c78e212205d3c",
          "url": "https://github.com/PoyrazK/thecloud/commit/9c2dc134ddfe3e50e902ffa79d5e7a2def70b410"
        },
        "date": 1768151842984,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641512054 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641512054 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641512054 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641512054 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 137.7,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8443005 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 137.7,
            "unit": "ns/op",
            "extra": "8443005 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8443005 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8443005 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 419.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2865872 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 419.2,
            "unit": "ns/op",
            "extra": "2865872 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2865872 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2865872 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 47832,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "32703 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 47832,
            "unit": "ns/op",
            "extra": "32703 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "32703 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "32703 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 234.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5882920 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 234.7,
            "unit": "ns/op",
            "extra": "5882920 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5882920 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5882920 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 184.4,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5704941 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 184.4,
            "unit": "ns/op",
            "extra": "5704941 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5704941 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5704941 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.885,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "633551325 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.885,
            "unit": "ns/op",
            "extra": "633551325 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "633551325 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "633551325 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.1,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11272430 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.1,
            "unit": "ns/op",
            "extra": "11272430 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11272430 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11272430 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.91,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203056077 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.91,
            "unit": "ns/op",
            "extra": "203056077 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203056077 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203056077 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640939987 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "640939987 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640939987 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640939987 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.26,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "26214807 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.26,
            "unit": "ns/op",
            "extra": "26214807 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "26214807 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "26214807 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66431681,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66431681,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66371338,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66371338,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "1b112c21b1199f079c5678026655f01bc0250cc6",
          "message": "fix(services): update event_test to match NewEventService signature",
          "timestamp": "2026-01-11T20:36:03+03:00",
          "tree_id": "6ef5055c9782118fd3b2e46d245de73b057d2822",
          "url": "https://github.com/PoyrazK/thecloud/commit/1b112c21b1199f079c5678026655f01bc0250cc6"
        },
        "date": 1768153048048,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641055440 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641055440 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641055440 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641055440 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.2,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8863710 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.2,
            "unit": "ns/op",
            "extra": "8863710 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8863710 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8863710 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 422.9,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2801844 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 422.9,
            "unit": "ns/op",
            "extra": "2801844 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2801844 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2801844 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 51730,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "28138 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 51730,
            "unit": "ns/op",
            "extra": "28138 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "28138 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "28138 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 266,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5494704 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 266,
            "unit": "ns/op",
            "extra": "5494704 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5494704 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5494704 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 186.2,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5613174 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 186.2,
            "unit": "ns/op",
            "extra": "5613174 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5613174 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5613174 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642600757 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642600757 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642600757 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642600757 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "10837627 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104.7,
            "unit": "ns/op",
            "extra": "10837627 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "10837627 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "10837627 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.301,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226268992 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.301,
            "unit": "ns/op",
            "extra": "226268992 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226268992 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226268992 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "635240223 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "635240223 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "635240223 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "635240223 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.9,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28327101 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.9,
            "unit": "ns/op",
            "extra": "28327101 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28327101 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28327101 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66924616,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66924616,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66680509,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66680509,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "c3b170af8616e350487f0e8bb507fa3af4f3c460",
          "message": "fix(setup): update setup_test to pass logger to InitHandlers",
          "timestamp": "2026-01-11T20:39:15+03:00",
          "tree_id": "852af758a66aa7913315e99922007795d057f7ca",
          "url": "https://github.com/PoyrazK/thecloud/commit/c3b170af8616e350487f0e8bb507fa3af4f3c460"
        },
        "date": 1768153238093,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.891,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642106063 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.891,
            "unit": "ns/op",
            "extra": "642106063 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642106063 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642106063 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.9,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8894575 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.9,
            "unit": "ns/op",
            "extra": "8894575 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8894575 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8894575 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 424.4,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2842561 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 424.4,
            "unit": "ns/op",
            "extra": "2842561 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2842561 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2842561 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 49393,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "30624 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 49393,
            "unit": "ns/op",
            "extra": "30624 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "30624 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "30624 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 217.9,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5528611 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 217.9,
            "unit": "ns/op",
            "extra": "5528611 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5528611 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5528611 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 188.3,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5648288 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 188.3,
            "unit": "ns/op",
            "extra": "5648288 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5648288 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5648288 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.892,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641630445 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.892,
            "unit": "ns/op",
            "extra": "641630445 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641630445 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641630445 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 105.8,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11624383 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 105.8,
            "unit": "ns/op",
            "extra": "11624383 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11624383 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11624383 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.302,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226644241 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.302,
            "unit": "ns/op",
            "extra": "226644241 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226644241 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226644241 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.874,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "637051382 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.874,
            "unit": "ns/op",
            "extra": "637051382 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "637051382 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "637051382 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.54,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28724667 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.54,
            "unit": "ns/op",
            "extra": "28724667 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28724667 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28724667 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66318451,
            "unit": "ns/op\t    6424 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66318451,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6424,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66329949,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66329949,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "5c0736dfc0812e3d14a4ac204257585a88b7be85",
          "message": "refactor(api): modularize router setup and fix golangci-lint configuration",
          "timestamp": "2026-01-11T20:55:30+03:00",
          "tree_id": "56203db6dfb9e10a875fe01f4cfdf0da52fde6bf",
          "url": "https://github.com/PoyrazK/thecloud/commit/5c0736dfc0812e3d14a4ac204257585a88b7be85"
        },
        "date": 1768154218688,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643354880 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "643354880 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643354880 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643354880 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8991643 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.3,
            "unit": "ns/op",
            "extra": "8991643 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8991643 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8991643 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 426.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2810347 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 426.2,
            "unit": "ns/op",
            "extra": "2810347 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2810347 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2810347 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 50474,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "31736 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 50474,
            "unit": "ns/op",
            "extra": "31736 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "31736 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "31736 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 214.5,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5536545 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 214.5,
            "unit": "ns/op",
            "extra": "5536545 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5536545 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5536545 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 184.5,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "6222003 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 184.5,
            "unit": "ns/op",
            "extra": "6222003 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "6222003 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "6222003 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.876,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "633220280 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.876,
            "unit": "ns/op",
            "extra": "633220280 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "633220280 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "633220280 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 103.3,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11668683 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 103.3,
            "unit": "ns/op",
            "extra": "11668683 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11668683 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11668683 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.296,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226646744 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.296,
            "unit": "ns/op",
            "extra": "226646744 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226646744 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226646744 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641886254 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "641886254 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641886254 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641886254 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.84,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28944267 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.84,
            "unit": "ns/op",
            "extra": "28944267 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28944267 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28944267 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66393762,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66393762,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66303494,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66303494,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "1cb05bafa64351110eddd7fa314e937d3bd1a247",
          "message": "fix(ci): restore golangci-lint config version for CI compatibility",
          "timestamp": "2026-01-11T20:58:43+03:00",
          "tree_id": "880f178c95278aa12330b11de28b008358fc53e6",
          "url": "https://github.com/PoyrazK/thecloud/commit/1cb05bafa64351110eddd7fa314e937d3bd1a247"
        },
        "date": 1768154410006,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641075402 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641075402 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641075402 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641075402 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8926612 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.1,
            "unit": "ns/op",
            "extra": "8926612 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8926612 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8926612 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 462,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2846942 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 462,
            "unit": "ns/op",
            "extra": "2846942 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2846942 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2846942 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 48744,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "32376 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 48744,
            "unit": "ns/op",
            "extra": "32376 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "32376 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "32376 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 219,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5634109 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 219,
            "unit": "ns/op",
            "extra": "5634109 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5634109 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5634109 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 186.4,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5596291 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 186.4,
            "unit": "ns/op",
            "extra": "5596291 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5596291 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5596291 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.884,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641947984 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.884,
            "unit": "ns/op",
            "extra": "641947984 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641947984 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641947984 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 106.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11279724 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 106.7,
            "unit": "ns/op",
            "extra": "11279724 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11279724 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11279724 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.309,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226056178 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.309,
            "unit": "ns/op",
            "extra": "226056178 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226056178 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226056178 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642490501 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "642490501 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642490501 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642490501 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.11,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27646794 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.11,
            "unit": "ns/op",
            "extra": "27646794 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27646794 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27646794 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66419883,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66419883,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66382276,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66382276,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "58ee054230fa0bcf47c413005552d45b181b1154",
          "message": "fix(ci): migrate golangci-lint config to v2 schema",
          "timestamp": "2026-01-11T21:03:02+03:00",
          "tree_id": "e74d9262b949814768d22c16a6aa82a091ef089f",
          "url": "https://github.com/PoyrazK/thecloud/commit/58ee054230fa0bcf47c413005552d45b181b1154"
        },
        "date": 1768154668375,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.904,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641971642 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.904,
            "unit": "ns/op",
            "extra": "641971642 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641971642 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641971642 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 164.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8278938 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 164.3,
            "unit": "ns/op",
            "extra": "8278938 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8278938 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8278938 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 440.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2711200 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 440.7,
            "unit": "ns/op",
            "extra": "2711200 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2711200 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2711200 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 52669,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "27459 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 52669,
            "unit": "ns/op",
            "extra": "27459 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "27459 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "27459 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 216.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5495031 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 216.7,
            "unit": "ns/op",
            "extra": "5495031 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5495031 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5495031 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 190.6,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5369488 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 190.6,
            "unit": "ns/op",
            "extra": "5369488 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5369488 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5369488 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.874,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639348712 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.874,
            "unit": "ns/op",
            "extra": "639348712 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639348712 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639348712 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 108,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "10865960 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 108,
            "unit": "ns/op",
            "extra": "10865960 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "10865960 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "10865960 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.301,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "225268004 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.301,
            "unit": "ns/op",
            "extra": "225268004 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "225268004 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "225268004 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641138548 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641138548 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641138548 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641138548 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 46.18,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "24386287 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 46.18,
            "unit": "ns/op",
            "extra": "24386287 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "24386287 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "24386287 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66477986,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66477986,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66468310,
            "unit": "ns/op\t    5605 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66468310,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5605,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "ee4ee7855825004bf240feb4d54c165c34bebba0",
          "message": "fix(ws): fix initialism mockId -> mockID to satisfy staticcheck",
          "timestamp": "2026-01-11T21:06:08+03:00",
          "tree_id": "4e4478bb9534c68914276aefc565537b2ced7019",
          "url": "https://github.com/PoyrazK/thecloud/commit/ee4ee7855825004bf240feb4d54c165c34bebba0"
        },
        "date": 1768154852389,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "623448777 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "623448777 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "623448777 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "623448777 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 136.9,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8794748 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 136.9,
            "unit": "ns/op",
            "extra": "8794748 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8794748 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8794748 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 422.9,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2845568 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 422.9,
            "unit": "ns/op",
            "extra": "2845568 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2845568 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2845568 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 49946,
            "unit": "ns/op\t     992 B/op\t      18 allocs/op",
            "extra": "27576 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 49946,
            "unit": "ns/op",
            "extra": "27576 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 992,
            "unit": "B/op",
            "extra": "27576 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "27576 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 217.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5360590 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 217.7,
            "unit": "ns/op",
            "extra": "5360590 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5360590 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5360590 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 187.2,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5585461 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 187.2,
            "unit": "ns/op",
            "extra": "5585461 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5585461 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5585461 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.888,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "607057952 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.888,
            "unit": "ns/op",
            "extra": "607057952 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "607057952 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "607057952 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 106.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "10675102 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 106.7,
            "unit": "ns/op",
            "extra": "10675102 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "10675102 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "10675102 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.305,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226659650 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.305,
            "unit": "ns/op",
            "extra": "226659650 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226659650 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226659650 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.892,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642284751 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.892,
            "unit": "ns/op",
            "extra": "642284751 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642284751 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642284751 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.28,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28594448 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.28,
            "unit": "ns/op",
            "extra": "28594448 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28594448 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28594448 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66470795,
            "unit": "ns/op\t    6420 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66470795,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6420,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66289742,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66289742,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "c00406ed28cbe1db0c7ff91a024b3711e1041f9f",
          "message": "Merge fix/sonarqube-issues: Comprehensive SonarQube warnings cleanup",
          "timestamp": "2026-01-11T23:27:14+03:00",
          "tree_id": "c0f6de908911a77f36e0224d86d78e0526bbe28d",
          "url": "https://github.com/PoyrazK/thecloud/commit/c00406ed28cbe1db0c7ff91a024b3711e1041f9f"
        },
        "date": 1768163334998,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641416784 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "641416784 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641416784 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641416784 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8865446 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.8,
            "unit": "ns/op",
            "extra": "8865446 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8865446 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8865446 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 417.6,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2893137 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 417.6,
            "unit": "ns/op",
            "extra": "2893137 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2893137 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2893137 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54254,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22460 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54254,
            "unit": "ns/op",
            "extra": "22460 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22460 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22460 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 214.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5586549 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 214.8,
            "unit": "ns/op",
            "extra": "5586549 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5586549 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5586549 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 183.4,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "6119604 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 183.4,
            "unit": "ns/op",
            "extra": "6119604 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "6119604 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "6119604 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641743062 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "641743062 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641743062 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641743062 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104.2,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11902179 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104.2,
            "unit": "ns/op",
            "extra": "11902179 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11902179 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11902179 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.296,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226490274 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.296,
            "unit": "ns/op",
            "extra": "226490274 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226490274 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226490274 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640298901 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "640298901 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640298901 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640298901 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28362064 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42,
            "unit": "ns/op",
            "extra": "28362064 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28362064 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28362064 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66403358,
            "unit": "ns/op\t    6420 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66403358,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6420,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66341734,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66341734,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "2d90ee55e830768ec420365630eb7a53f9e4d7af",
          "message": "Fix CI issues: Add Redis service to load tests and regenerate Swagger docs",
          "timestamp": "2026-01-11T23:46:28+03:00",
          "tree_id": "038e40aab81e77232e69e7f495c64ad772589af7",
          "url": "https://github.com/PoyrazK/thecloud/commit/2d90ee55e830768ec420365630eb7a53f9e4d7af"
        },
        "date": 1768164480004,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641351866 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "641351866 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641351866 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641351866 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8939367 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134,
            "unit": "ns/op",
            "extra": "8939367 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8939367 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8939367 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 413.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2887933 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 413.8,
            "unit": "ns/op",
            "extra": "2887933 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2887933 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2887933 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54947,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21897 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54947,
            "unit": "ns/op",
            "extra": "21897 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21897 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21897 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 221.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5293515 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 221.2,
            "unit": "ns/op",
            "extra": "5293515 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5293515 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5293515 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 187,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5705281 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 187,
            "unit": "ns/op",
            "extra": "5705281 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5705281 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5705281 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "635753798 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "635753798 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "635753798 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "635753798 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 119.6,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "10171939 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 119.6,
            "unit": "ns/op",
            "extra": "10171939 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "10171939 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "10171939 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.295,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226457995 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.295,
            "unit": "ns/op",
            "extra": "226457995 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226457995 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226457995 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.872,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "634521943 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.872,
            "unit": "ns/op",
            "extra": "634521943 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "634521943 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "634521943 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.3,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27602413 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.3,
            "unit": "ns/op",
            "extra": "27602413 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27602413 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27602413 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66356831,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66356831,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66330557,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66330557,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "71cdbebfc8c4746706af3f1815f694c1bf428358",
          "message": "Fix staticcheck warning: remove empty if branch in autoscaling worker",
          "timestamp": "2026-01-11T23:50:21+03:00",
          "tree_id": "74b33ed956a446f081610690e67cf2668b6664c4",
          "url": "https://github.com/PoyrazK/thecloud/commit/71cdbebfc8c4746706af3f1815f694c1bf428358"
        },
        "date": 1768164702626,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643042309 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "643042309 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643042309 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643042309 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.5,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8821776 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.5,
            "unit": "ns/op",
            "extra": "8821776 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8821776 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8821776 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 415.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2858676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 415.2,
            "unit": "ns/op",
            "extra": "2858676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2858676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2858676 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55431,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22290 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55431,
            "unit": "ns/op",
            "extra": "22290 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22290 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22290 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 223,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5559553 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 223,
            "unit": "ns/op",
            "extra": "5559553 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5559553 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5559553 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 183.8,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5746470 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 183.8,
            "unit": "ns/op",
            "extra": "5746470 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5746470 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5746470 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643154454 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "643154454 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643154454 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643154454 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 106.4,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11204385 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 106.4,
            "unit": "ns/op",
            "extra": "11204385 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11204385 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11204385 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.304,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226561699 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.304,
            "unit": "ns/op",
            "extra": "226561699 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226561699 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226561699 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640971373 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "640971373 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640971373 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640971373 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.69,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28667052 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.69,
            "unit": "ns/op",
            "extra": "28667052 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28667052 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28667052 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66348790,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66348790,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66378234,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66378234,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "592527359bccea5bb026a988148f5f682f2971b9",
          "message": "Security: Fix CodeQL path traversal warnings in libvirt adapter\n\n- Add filepath.Base() to ensure safeName contains no path separators\n- Add validation to prevent '.' and '..' directory references\n- Use filepath.Clean() when constructing ISO path\n- Addresses CodeQL warnings about uncontrolled data in path expressions",
          "timestamp": "2026-01-11T23:56:02+03:00",
          "tree_id": "1e1079cd0246a094ffd7635de7e05cae72950f9d",
          "url": "https://github.com/PoyrazK/thecloud/commit/592527359bccea5bb026a988148f5f682f2971b9"
        },
        "date": 1768165047504,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "632962676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "632962676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "632962676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "632962676 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9002646 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134,
            "unit": "ns/op",
            "extra": "9002646 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9002646 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9002646 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 412.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2889535 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 412.8,
            "unit": "ns/op",
            "extra": "2889535 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2889535 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2889535 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55966,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22417 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55966,
            "unit": "ns/op",
            "extra": "22417 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22417 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22417 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 214.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5623245 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 214.2,
            "unit": "ns/op",
            "extra": "5623245 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5623245 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5623245 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 184.7,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5693670 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 184.7,
            "unit": "ns/op",
            "extra": "5693670 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5693670 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5693670 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639521661 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "639521661 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639521661 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639521661 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 102.5,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11482134 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 102.5,
            "unit": "ns/op",
            "extra": "11482134 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11482134 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11482134 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.308,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226148787 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.308,
            "unit": "ns/op",
            "extra": "226148787 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226148787 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226148787 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "637840476 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "637840476 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "637840476 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "637840476 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.62,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28307356 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.62,
            "unit": "ns/op",
            "extra": "28307356 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28307356 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28307356 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66391140,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66391140,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66699386,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66699386,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "fa96f89fb1efe5d7e616fdc5402cf6763d904c4e",
          "message": "Security: Improve password display security in CLI\n\n- Add explicit security warning when displaying database password\n- Add nosemgrep comment to indicate intentional password display\n- Clarify that password is shown only once at creation time\n- Addresses CodeQL warning about clear-text logging of sensitive information\n\nThe password display is intentional and necessary for CLI usability,\nbut now includes clear warnings about the security implications.",
          "timestamp": "2026-01-11T23:58:57+03:00",
          "tree_id": "c2ef2f45de3c58c9b0973f96c3ef0693bd81f644",
          "url": "https://github.com/PoyrazK/thecloud/commit/fa96f89fb1efe5d7e616fdc5402cf6763d904c4e"
        },
        "date": 1768165221775,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643154788 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "643154788 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643154788 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643154788 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8842137 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.8,
            "unit": "ns/op",
            "extra": "8842137 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8842137 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8842137 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 413.9,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2895812 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 413.9,
            "unit": "ns/op",
            "extra": "2895812 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2895812 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2895812 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55078,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22411 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55078,
            "unit": "ns/op",
            "extra": "22411 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22411 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22411 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 212.1,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5539500 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 212.1,
            "unit": "ns/op",
            "extra": "5539500 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5539500 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5539500 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 188.5,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5809506 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 188.5,
            "unit": "ns/op",
            "extra": "5809506 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5809506 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5809506 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.873,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642766197 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.873,
            "unit": "ns/op",
            "extra": "642766197 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642766197 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642766197 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 103,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12204434 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 103,
            "unit": "ns/op",
            "extra": "12204434 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12204434 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12204434 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.297,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226523140 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.297,
            "unit": "ns/op",
            "extra": "226523140 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226523140 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226523140 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.885,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639271530 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.885,
            "unit": "ns/op",
            "extra": "639271530 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639271530 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639271530 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.96,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28892630 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.96,
            "unit": "ns/op",
            "extra": "28892630 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28892630 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28892630 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66360578,
            "unit": "ns/op\t    6420 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66360578,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6420,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66371353,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66371353,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "8754b04a89172e79fa6843652bbd6a2a0367a705",
          "message": "chore(deps): Update Go dependencies to latest versions\n\n- Updated OpenTelemetry from 1.38.0 to 1.39.0\n- Updated Prometheus libraries (common 0.66.1→0.67.5, procfs 0.16.1→0.19.2)\n- Updated go-openapi packages (spec, swag, jsonpointer, jsonreference)\n- Updated validator/v10 from 10.27.0 to 10.30.1\n- Updated protobuf from 1.36.10 to 1.36.11\n- Updated various golang.org/x packages (sys, text, mod, arch)\n- All tests passing ✓",
          "timestamp": "2026-01-12T00:03:04+03:00",
          "tree_id": "f59e1599c84a2131cc378b5c96d45aab39e34615",
          "url": "https://github.com/PoyrazK/thecloud/commit/8754b04a89172e79fa6843652bbd6a2a0367a705"
        },
        "date": 1768165483263,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.878,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641437143 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.878,
            "unit": "ns/op",
            "extra": "641437143 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641437143 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641437143 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 143.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9207712 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 143.3,
            "unit": "ns/op",
            "extra": "9207712 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9207712 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9207712 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 415.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2878669 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 415.7,
            "unit": "ns/op",
            "extra": "2878669 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2878669 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2878669 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55757,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22292 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55757,
            "unit": "ns/op",
            "extra": "22292 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22292 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22292 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 216.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5351085 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 216.8,
            "unit": "ns/op",
            "extra": "5351085 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5351085 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5351085 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 189.2,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5801950 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 189.2,
            "unit": "ns/op",
            "extra": "5801950 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5801950 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5801950 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.886,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642771180 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.886,
            "unit": "ns/op",
            "extra": "642771180 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642771180 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642771180 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11496640 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104,
            "unit": "ns/op",
            "extra": "11496640 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11496640 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11496640 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.301,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226197806 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.301,
            "unit": "ns/op",
            "extra": "226197806 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226197806 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226197806 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "638468943 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "638468943 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "638468943 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "638468943 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.72,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28434374 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.72,
            "unit": "ns/op",
            "extra": "28434374 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28434374 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28434374 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66347767,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66347767,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66389195,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66389195,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "3d56133018972ba4b09bd063ab0b5fbe58f9cfa7",
          "message": "docs: Add comprehensive package documentation\n\nAdded detailed package-level documentation for:\n- Main package (doc.go): Project overview, architecture, quick start\n- domain package: Business entities, lifecycle, multi-tenancy\n- ports package: Hexagonal architecture, port categories, DI pattern\n- services package: Business logic layer, service patterns, workers\n\nBenefits:\n- Improved code discoverability via godoc\n- Clear architecture documentation\n- Better onboarding for new developers\n- Examples and usage patterns",
          "timestamp": "2026-01-12T00:04:59+03:00",
          "tree_id": "0a80d19e43b3c6df022e222bb59c0e5fcac745a8",
          "url": "https://github.com/PoyrazK/thecloud/commit/3d56133018972ba4b09bd063ab0b5fbe58f9cfa7"
        },
        "date": 1768165587786,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640508624 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "640508624 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640508624 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640508624 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 131.4,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9268629 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 131.4,
            "unit": "ns/op",
            "extra": "9268629 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9268629 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9268629 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 418.1,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2883753 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 418.1,
            "unit": "ns/op",
            "extra": "2883753 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2883753 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2883753 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55012,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22333 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55012,
            "unit": "ns/op",
            "extra": "22333 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22333 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22333 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 218.5,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5397188 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 218.5,
            "unit": "ns/op",
            "extra": "5397188 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5397188 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5397188 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 185.1,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5667651 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 185.1,
            "unit": "ns/op",
            "extra": "5667651 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5667651 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5667651 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "622683525 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "622683525 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "622683525 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "622683525 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 102.6,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12044020 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 102.6,
            "unit": "ns/op",
            "extra": "12044020 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12044020 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12044020 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.296,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226468928 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.296,
            "unit": "ns/op",
            "extra": "226468928 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226468928 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226468928 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642666816 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "642666816 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642666816 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642666816 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.8,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28461976 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.8,
            "unit": "ns/op",
            "extra": "28461976 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28461976 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28461976 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66917214,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66917214,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66384879,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66384879,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "dabd6f316c39bc979e7a4aef3fe73a42c67d46c6",
          "message": "docs: Improve and simplify godoc comments\n\n- Made documentation more concise while keeping essential information\n- Added clear, brief comments to Instance and InstanceService\n- Fixed doc.go package declaration (thecloud not main)\n- Focused on practical information over verbose explanations\n\nImproved godoc readability for:\n- domain/instance.go: Instance struct and related types\n- services/instance.go: InstanceService and dependencies\n- Package-level documentation",
          "timestamp": "2026-01-12T00:10:52+03:00",
          "tree_id": "6881cc7f6097633cbf2ee20d4860f089fab93a92",
          "url": "https://github.com/PoyrazK/thecloud/commit/dabd6f316c39bc979e7a4aef3fe73a42c67d46c6"
        },
        "date": 1768165937512,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.458,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "820181522 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.458,
            "unit": "ns/op",
            "extra": "820181522 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "820181522 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "820181522 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 127.4,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9303222 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 127.4,
            "unit": "ns/op",
            "extra": "9303222 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9303222 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9303222 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 427.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "3092040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 427.7,
            "unit": "ns/op",
            "extra": "3092040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "3092040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3092040 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 26318,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "49326 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 26318,
            "unit": "ns/op",
            "extra": "49326 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "49326 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "49326 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 241.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5045793 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 241.8,
            "unit": "ns/op",
            "extra": "5045793 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5045793 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5045793 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 187.1,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5564152 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 187.1,
            "unit": "ns/op",
            "extra": "5564152 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5564152 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5564152 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.483,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "820175877 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.483,
            "unit": "ns/op",
            "extra": "820175877 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "820175877 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "820175877 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 113.9,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "10324122 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 113.9,
            "unit": "ns/op",
            "extra": "10324122 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "10324122 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "10324122 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 3.811,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "315415005 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 3.811,
            "unit": "ns/op",
            "extra": "315415005 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "315415005 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "315415005 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.546,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "783534016 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.546,
            "unit": "ns/op",
            "extra": "783534016 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "783534016 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "783534016 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 39.16,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "29409043 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 39.16,
            "unit": "ns/op",
            "extra": "29409043 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "29409043 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "29409043 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 63658630,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 63658630,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 63560232,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 63560232,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "8918be361e01f7eca1d0e8de8da4f4e359fe84af",
          "message": "chore: Regenerate Swagger docs after godoc improvements",
          "timestamp": "2026-01-12T00:13:07+03:00",
          "tree_id": "11dd073eeee0896cb3deabab04646a2448a4fb7b",
          "url": "https://github.com/PoyrazK/thecloud/commit/8918be361e01f7eca1d0e8de8da4f4e359fe84af"
        },
        "date": 1768166069419,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "625156422 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "625156422 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "625156422 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "625156422 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 136.6,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8928295 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 136.6,
            "unit": "ns/op",
            "extra": "8928295 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8928295 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8928295 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 430.1,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2784512 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 430.1,
            "unit": "ns/op",
            "extra": "2784512 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2784512 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2784512 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54689,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22249 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54689,
            "unit": "ns/op",
            "extra": "22249 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22249 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22249 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 265,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "4520040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 265,
            "unit": "ns/op",
            "extra": "4520040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "4520040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "4520040 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 184.1,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5648900 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 184.1,
            "unit": "ns/op",
            "extra": "5648900 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5648900 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5648900 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "636139298 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "636139298 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "636139298 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "636139298 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11547966 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104,
            "unit": "ns/op",
            "extra": "11547966 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11547966 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11547966 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.295,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226416480 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.295,
            "unit": "ns/op",
            "extra": "226416480 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226416480 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226416480 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642319440 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "642319440 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642319440 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642319440 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.03,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28289247 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.03,
            "unit": "ns/op",
            "extra": "28289247 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28289247 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28289247 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66377599,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66377599,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66344366,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66344366,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "2be63c83acb377e056d22b879fd2825b96f14785",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/2be63c83acb377e056d22b879fd2825b96f14785"
        },
        "date": 1768331783621,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.873,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640445838 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.873,
            "unit": "ns/op",
            "extra": "640445838 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640445838 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640445838 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 138.9,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8662687 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 138.9,
            "unit": "ns/op",
            "extra": "8662687 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8662687 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8662687 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 423.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2429546 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 423.7,
            "unit": "ns/op",
            "extra": "2429546 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2429546 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2429546 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 56837,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22012 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 56837,
            "unit": "ns/op",
            "extra": "22012 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22012 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22012 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 224.1,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5086159 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 224.1,
            "unit": "ns/op",
            "extra": "5086159 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5086159 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5086159 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 187.8,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5515552 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 187.8,
            "unit": "ns/op",
            "extra": "5515552 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5515552 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5515552 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641666299 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641666299 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641666299 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641666299 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 105.9,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11046003 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 105.9,
            "unit": "ns/op",
            "extra": "11046003 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11046003 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11046003 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.916,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202926606 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.916,
            "unit": "ns/op",
            "extra": "202926606 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202926606 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202926606 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642270873 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642270873 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642270873 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642270873 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.26,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "29002221 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.26,
            "unit": "ns/op",
            "extra": "29002221 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "29002221 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "29002221 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66756075,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66756075,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66696430,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66696430,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "ce9dedd94b5744c68842e0a66a86d989babac366",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/ce9dedd94b5744c68842e0a66a86d989babac366"
        },
        "date": 1768380756428,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641983416 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "641983416 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641983416 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641983416 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 139.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8639413 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 139.8,
            "unit": "ns/op",
            "extra": "8639413 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8639413 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8639413 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 417.4,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2870617 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 417.4,
            "unit": "ns/op",
            "extra": "2870617 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2870617 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2870617 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 57470,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22071 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 57470,
            "unit": "ns/op",
            "extra": "22071 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22071 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22071 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 221.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5010818 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 221.2,
            "unit": "ns/op",
            "extra": "5010818 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5010818 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5010818 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 186.9,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5612698 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 186.9,
            "unit": "ns/op",
            "extra": "5612698 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5612698 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5612698 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641094478 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "641094478 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641094478 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641094478 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.2,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11438887 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.2,
            "unit": "ns/op",
            "extra": "11438887 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11438887 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11438887 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.923,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203351788 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.923,
            "unit": "ns/op",
            "extra": "203351788 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203351788 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203351788 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.874,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639103315 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.874,
            "unit": "ns/op",
            "extra": "639103315 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639103315 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639103315 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.03,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28190668 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.03,
            "unit": "ns/op",
            "extra": "28190668 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28190668 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28190668 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66783496,
            "unit": "ns/op\t    6424 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66783496,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6424,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66634996,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66634996,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "f0644c655d89e06b2d61035027228f73fe74f98b",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/f0644c655d89e06b2d61035027228f73fe74f98b"
        },
        "date": 1768380930853,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641731040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641731040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641731040 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641731040 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 157.9,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8632892 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 157.9,
            "unit": "ns/op",
            "extra": "8632892 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8632892 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8632892 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 411.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2926026 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 411.8,
            "unit": "ns/op",
            "extra": "2926026 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2926026 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2926026 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54830,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22394 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54830,
            "unit": "ns/op",
            "extra": "22394 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22394 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22394 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 221.3,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5455222 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 221.3,
            "unit": "ns/op",
            "extra": "5455222 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5455222 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5455222 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 191.1,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5537793 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 191.1,
            "unit": "ns/op",
            "extra": "5537793 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5537793 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5537793 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "627941548 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "627941548 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "627941548 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "627941548 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 108.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11152944 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 108.7,
            "unit": "ns/op",
            "extra": "11152944 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11152944 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11152944 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.925,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202859482 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.925,
            "unit": "ns/op",
            "extra": "202859482 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202859482 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202859482 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "632185975 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "632185975 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "632185975 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "632185975 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.31,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28320595 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.31,
            "unit": "ns/op",
            "extra": "28320595 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28320595 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28320595 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66698173,
            "unit": "ns/op\t    6452 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66698173,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6452,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66676482,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66676482,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "cebfdcc69dbf43148b191434919cca761b589577",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/cebfdcc69dbf43148b191434919cca761b589577"
        },
        "date": 1768380979286,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.448,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "823532676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.448,
            "unit": "ns/op",
            "extra": "823532676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "823532676 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "823532676 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 128.7,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9108788 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 128.7,
            "unit": "ns/op",
            "extra": "9108788 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9108788 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9108788 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 397.1,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "3021657 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 397.1,
            "unit": "ns/op",
            "extra": "3021657 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "3021657 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3021657 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 29804,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "48192 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 29804,
            "unit": "ns/op",
            "extra": "48192 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "48192 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "48192 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 237.9,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5090618 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 237.9,
            "unit": "ns/op",
            "extra": "5090618 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5090618 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5090618 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 203.9,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5509060 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 203.9,
            "unit": "ns/op",
            "extra": "5509060 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5509060 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5509060 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.46,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "787149030 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.46,
            "unit": "ns/op",
            "extra": "787149030 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "787149030 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "787149030 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 123.9,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "9663926 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 123.9,
            "unit": "ns/op",
            "extra": "9663926 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "9663926 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "9663926 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 4.057,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "295881214 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 4.057,
            "unit": "ns/op",
            "extra": "295881214 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "295881214 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "295881214 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.607,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "743985370 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.607,
            "unit": "ns/op",
            "extra": "743985370 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "743985370 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "743985370 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.13,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "29659273 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.13,
            "unit": "ns/op",
            "extra": "29659273 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "29659273 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "29659273 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 63556712,
            "unit": "ns/op\t    6433 B/op\t      36 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 63556712,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6433,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 63520260,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 63520260,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "1fdec6f3b9994e0585d6a95bbbd1ebc4582b57ce",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/1fdec6f3b9994e0585d6a95bbbd1ebc4582b57ce"
        },
        "date": 1768384899738,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640936234 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "640936234 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640936234 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640936234 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 140.6,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8704807 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 140.6,
            "unit": "ns/op",
            "extra": "8704807 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8704807 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8704807 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 411.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2924077 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 411.8,
            "unit": "ns/op",
            "extra": "2924077 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2924077 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2924077 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 53736,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22894 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 53736,
            "unit": "ns/op",
            "extra": "22894 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22894 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22894 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 220,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5379408 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 220,
            "unit": "ns/op",
            "extra": "5379408 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5379408 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5379408 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 198.5,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5581852 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 198.5,
            "unit": "ns/op",
            "extra": "5581852 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5581852 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5581852 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.898,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "631562925 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.898,
            "unit": "ns/op",
            "extra": "631562925 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "631562925 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "631562925 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 109,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11586090 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 109,
            "unit": "ns/op",
            "extra": "11586090 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11586090 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11586090 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.913,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202399616 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.913,
            "unit": "ns/op",
            "extra": "202399616 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202399616 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202399616 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.875,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641138172 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.875,
            "unit": "ns/op",
            "extra": "641138172 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641138172 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641138172 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.21,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28248832 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.21,
            "unit": "ns/op",
            "extra": "28248832 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28248832 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28248832 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66688942,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66688942,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66595204,
            "unit": "ns/op\t    5595 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66595204,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5595,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "383393a2acb48faf3e79ea28653011d955b59b67",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/383393a2acb48faf3e79ea28653011d955b59b67"
        },
        "date": 1768385344688,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641507124 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641507124 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641507124 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641507124 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 138.4,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8398039 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 138.4,
            "unit": "ns/op",
            "extra": "8398039 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8398039 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8398039 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 411.4,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2930930 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 411.4,
            "unit": "ns/op",
            "extra": "2930930 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2930930 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2930930 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55111,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21976 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55111,
            "unit": "ns/op",
            "extra": "21976 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21976 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21976 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 221.3,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5531991 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 221.3,
            "unit": "ns/op",
            "extra": "5531991 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5531991 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5531991 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 185.6,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5658312 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 185.6,
            "unit": "ns/op",
            "extra": "5658312 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5658312 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5658312 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.879,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "626966456 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.879,
            "unit": "ns/op",
            "extra": "626966456 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "626966456 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "626966456 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.1,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11736405 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.1,
            "unit": "ns/op",
            "extra": "11736405 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11736405 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11736405 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.916,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202421936 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.916,
            "unit": "ns/op",
            "extra": "202421936 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202421936 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202421936 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642185114 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642185114 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642185114 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642185114 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.35,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28276939 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.35,
            "unit": "ns/op",
            "extra": "28276939 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28276939 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28276939 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66662691,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66662691,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66637368,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66637368,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "8b785ebef4a364153c6214c6377b0eaef3d5ae17",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/8b785ebef4a364153c6214c6377b0eaef3d5ae17"
        },
        "date": 1768385835296,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641672588 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641672588 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641672588 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641672588 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 156.7,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8633116 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 156.7,
            "unit": "ns/op",
            "extra": "8633116 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8633116 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8633116 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 413.4,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2907478 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 413.4,
            "unit": "ns/op",
            "extra": "2907478 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2907478 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2907478 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54155,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22776 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54155,
            "unit": "ns/op",
            "extra": "22776 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22776 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22776 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 218.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5453714 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 218.2,
            "unit": "ns/op",
            "extra": "5453714 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5453714 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5453714 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 188.8,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5636660 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 188.8,
            "unit": "ns/op",
            "extra": "5636660 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5636660 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5636660 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "629568218 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "629568218 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "629568218 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "629568218 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104.6,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11460247 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104.6,
            "unit": "ns/op",
            "extra": "11460247 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11460247 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11460247 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.921,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203130440 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.921,
            "unit": "ns/op",
            "extra": "203130440 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203130440 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203130440 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642260031 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "642260031 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642260031 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642260031 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.71,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27904675 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.71,
            "unit": "ns/op",
            "extra": "27904675 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27904675 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27904675 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66654511,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66654511,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66692877,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66692877,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "617b13538f6b87193b8c420d3d60aef5ceef9615",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/617b13538f6b87193b8c420d3d60aef5ceef9615"
        },
        "date": 1768385895051,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641288706 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641288706 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641288706 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641288706 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 138.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8501062 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 138.3,
            "unit": "ns/op",
            "extra": "8501062 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8501062 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8501062 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 453.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2913296 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 453.8,
            "unit": "ns/op",
            "extra": "2913296 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2913296 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2913296 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55609,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22195 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55609,
            "unit": "ns/op",
            "extra": "22195 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22195 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22195 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 220.6,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5317424 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 220.6,
            "unit": "ns/op",
            "extra": "5317424 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5317424 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5317424 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 189.6,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5597070 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 189.6,
            "unit": "ns/op",
            "extra": "5597070 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5597070 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5597070 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.88,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642174160 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.88,
            "unit": "ns/op",
            "extra": "642174160 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642174160 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642174160 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11064157 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.7,
            "unit": "ns/op",
            "extra": "11064157 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11064157 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11064157 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.918,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "201887280 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.918,
            "unit": "ns/op",
            "extra": "201887280 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "201887280 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "201887280 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.887,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642785636 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.887,
            "unit": "ns/op",
            "extra": "642785636 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642785636 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642785636 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.38,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28003843 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.38,
            "unit": "ns/op",
            "extra": "28003843 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28003843 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28003843 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66644679,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66644679,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66669258,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66669258,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "8160b80619ee09f23b447f736f56a8c70dc706e2",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/8160b80619ee09f23b447f736f56a8c70dc706e2"
        },
        "date": 1768385919199,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.9,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "624756321 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.9,
            "unit": "ns/op",
            "extra": "624756321 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "624756321 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "624756321 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 139.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8678940 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 139.1,
            "unit": "ns/op",
            "extra": "8678940 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8678940 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8678940 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 411.9,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2904255 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 411.9,
            "unit": "ns/op",
            "extra": "2904255 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2904255 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2904255 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55432,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22190 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55432,
            "unit": "ns/op",
            "extra": "22190 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22190 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22190 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 218.5,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5416779 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 218.5,
            "unit": "ns/op",
            "extra": "5416779 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5416779 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5416779 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 187.1,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5624992 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 187.1,
            "unit": "ns/op",
            "extra": "5624992 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5624992 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5624992 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641793945 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "641793945 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641793945 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641793945 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.4,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11630155 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.4,
            "unit": "ns/op",
            "extra": "11630155 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11630155 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11630155 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.917,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203385409 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.917,
            "unit": "ns/op",
            "extra": "203385409 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203385409 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203385409 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "636424154 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "636424154 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "636424154 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "636424154 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.48,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28050127 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.48,
            "unit": "ns/op",
            "extra": "28050127 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28050127 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28050127 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66744283,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66744283,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 67010197,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 67010197,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "7ebf7ae037b496b87e40a13d7db10465af0b81e5",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/7ebf7ae037b496b87e40a13d7db10465af0b81e5"
        },
        "date": 1768386016489,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.895,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640538413 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.895,
            "unit": "ns/op",
            "extra": "640538413 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640538413 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640538413 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 140,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8640978 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 140,
            "unit": "ns/op",
            "extra": "8640978 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8640978 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8640978 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 416.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2827707 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 416.2,
            "unit": "ns/op",
            "extra": "2827707 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2827707 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2827707 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 56064,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21732 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 56064,
            "unit": "ns/op",
            "extra": "21732 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21732 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21732 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 221.3,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5470722 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 221.3,
            "unit": "ns/op",
            "extra": "5470722 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5470722 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5470722 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 191.3,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5428596 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 191.3,
            "unit": "ns/op",
            "extra": "5428596 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5428596 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5428596 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.884,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "606512650 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.884,
            "unit": "ns/op",
            "extra": "606512650 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "606512650 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "606512650 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 109.3,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11201547 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 109.3,
            "unit": "ns/op",
            "extra": "11201547 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11201547 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11201547 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.915,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "198072940 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.915,
            "unit": "ns/op",
            "extra": "198072940 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "198072940 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "198072940 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642492194 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "642492194 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642492194 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642492194 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.48,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28244428 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.48,
            "unit": "ns/op",
            "extra": "28244428 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28244428 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28244428 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66734266,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66734266,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66662881,
            "unit": "ns/op\t    5594 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66662881,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5594,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "1b514c152ec274e47d73ccc72c410dbb89afc605",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/1b514c152ec274e47d73ccc72c410dbb89afc605"
        },
        "date": 1768386404222,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642945216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642945216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642945216 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642945216 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 157.4,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8444640 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 157.4,
            "unit": "ns/op",
            "extra": "8444640 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8444640 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8444640 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 410.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2917712 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 410.2,
            "unit": "ns/op",
            "extra": "2917712 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2917712 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2917712 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54693,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22561 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54693,
            "unit": "ns/op",
            "extra": "22561 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22561 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22561 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 218.6,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5435394 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 218.6,
            "unit": "ns/op",
            "extra": "5435394 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5435394 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5435394 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 186.2,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5577577 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 186.2,
            "unit": "ns/op",
            "extra": "5577577 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5577577 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5577577 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642328507 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "642328507 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642328507 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642328507 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.1,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11103142 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.1,
            "unit": "ns/op",
            "extra": "11103142 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11103142 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11103142 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.922,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "199900857 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.922,
            "unit": "ns/op",
            "extra": "199900857 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "199900857 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "199900857 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.885,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642160666 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.885,
            "unit": "ns/op",
            "extra": "642160666 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642160666 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642160666 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.85,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28188376 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.85,
            "unit": "ns/op",
            "extra": "28188376 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28188376 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28188376 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66715013,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66715013,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66659354,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66659354,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "91cfe3aa1fd715006e5c1a33a9bde137e4101242",
          "message": "test: increase project coverage to 70%",
          "timestamp": "2026-01-11T21:13:12Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/19/commits/91cfe3aa1fd715006e5c1a33a9bde137e4101242"
        },
        "date": 1768386594706,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641191623 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "641191623 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641191623 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641191623 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 158.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8627661 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 158.8,
            "unit": "ns/op",
            "extra": "8627661 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8627661 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8627661 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 411.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2846162 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 411.7,
            "unit": "ns/op",
            "extra": "2846162 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2846162 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2846162 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 56817,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21980 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 56817,
            "unit": "ns/op",
            "extra": "21980 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21980 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21980 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 224.1,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5443938 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 224.1,
            "unit": "ns/op",
            "extra": "5443938 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5443938 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5443938 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 188.5,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5540106 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 188.5,
            "unit": "ns/op",
            "extra": "5540106 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5540106 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5540106 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641221692 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "641221692 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641221692 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641221692 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.4,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11367400 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.4,
            "unit": "ns/op",
            "extra": "11367400 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11367400 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11367400 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.919,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "199915639 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.919,
            "unit": "ns/op",
            "extra": "199915639 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "199915639 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "199915639 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641394736 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641394736 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641394736 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641394736 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.66,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27333463 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.66,
            "unit": "ns/op",
            "extra": "27333463 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27333463 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27333463 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66739211,
            "unit": "ns/op\t    6424 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66739211,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6424,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66693290,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66693290,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "distinct": false,
          "id": "91cfe3aa1fd715006e5c1a33a9bde137e4101242",
          "message": "fix: address staticcheck SA1006 warning in instance command",
          "timestamp": "2026-01-14T13:28:25+03:00",
          "tree_id": "585d0a74475745a5a44740a1f6e5a94b57818808",
          "url": "https://github.com/PoyrazK/thecloud/commit/91cfe3aa1fd715006e5c1a33a9bde137e4101242"
        },
        "date": 1768388701512,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "635659233 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "635659233 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "635659233 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "635659233 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 139.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8592680 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 139.3,
            "unit": "ns/op",
            "extra": "8592680 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8592680 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8592680 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 416.4,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2908010 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 416.4,
            "unit": "ns/op",
            "extra": "2908010 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2908010 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2908010 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55943,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22244 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55943,
            "unit": "ns/op",
            "extra": "22244 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22244 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22244 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 227.7,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5088373 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 227.7,
            "unit": "ns/op",
            "extra": "5088373 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5088373 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5088373 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 185.7,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5599152 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 185.7,
            "unit": "ns/op",
            "extra": "5599152 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5599152 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5599152 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.865,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640781325 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.865,
            "unit": "ns/op",
            "extra": "640781325 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640781325 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640781325 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 105.2,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11788336 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 105.2,
            "unit": "ns/op",
            "extra": "11788336 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11788336 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11788336 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.915,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202719386 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.915,
            "unit": "ns/op",
            "extra": "202719386 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202719386 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202719386 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.877,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641592205 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.877,
            "unit": "ns/op",
            "extra": "641592205 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641592205 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641592205 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.09,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28490918 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.09,
            "unit": "ns/op",
            "extra": "28490918 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28490918 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28490918 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66806583,
            "unit": "ns/op\t    6438 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66806583,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6438,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66663201,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66663201,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "10b3e1dc1f44dc857397434b4bfbdf117b783ae8",
          "message": "feat: add comprehensive GEMINI.md guidelines and automated workflows",
          "timestamp": "2026-01-14T14:05:32+03:00",
          "tree_id": "575a4aca82c324ca3ae281c6bc734c0b22548974",
          "url": "https://github.com/PoyrazK/thecloud/commit/10b3e1dc1f44dc857397434b4bfbdf117b783ae8"
        },
        "date": 1768388820482,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.88,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642440702 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.88,
            "unit": "ns/op",
            "extra": "642440702 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642440702 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642440702 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 138.2,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8563795 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 138.2,
            "unit": "ns/op",
            "extra": "8563795 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8563795 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8563795 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 412.2,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2915588 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 412.2,
            "unit": "ns/op",
            "extra": "2915588 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2915588 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2915588 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55507,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22260 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55507,
            "unit": "ns/op",
            "extra": "22260 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22260 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22260 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 219.9,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5419852 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 219.9,
            "unit": "ns/op",
            "extra": "5419852 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5419852 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5419852 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 186.6,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "5624466 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 186.6,
            "unit": "ns/op",
            "extra": "5624466 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "5624466 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5624466 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.882,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "634948068 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.882,
            "unit": "ns/op",
            "extra": "634948068 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "634948068 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "634948068 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11761112 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104.7,
            "unit": "ns/op",
            "extra": "11761112 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11761112 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11761112 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.918,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202787119 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.918,
            "unit": "ns/op",
            "extra": "202787119 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202787119 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202787119 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.873,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642111518 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.873,
            "unit": "ns/op",
            "extra": "642111518 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642111518 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642111518 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.77,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28479643 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.77,
            "unit": "ns/op",
            "extra": "28479643 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28479643 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28479643 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66718781,
            "unit": "ns/op\t    6424 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66718781,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6424,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66648398,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66648398,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "0e96d6eee05d7c60fbc2aeb480c1fa0f01464f46",
          "message": "Readme update && and some ci/cd fix",
          "timestamp": "2026-01-14T14:08:52+03:00",
          "tree_id": "365c6de376de6f49a69125d498f765bfd43f9578",
          "url": "https://github.com/PoyrazK/thecloud/commit/0e96d6eee05d7c60fbc2aeb480c1fa0f01464f46"
        },
        "date": 1768389022801,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.872,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643264424 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.872,
            "unit": "ns/op",
            "extra": "643264424 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643264424 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643264424 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 139.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8503860 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 139.8,
            "unit": "ns/op",
            "extra": "8503860 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8503860 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8503860 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 412.8,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "2920239 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 412.8,
            "unit": "ns/op",
            "extra": "2920239 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2920239 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "2920239 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54505,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22383 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54505,
            "unit": "ns/op",
            "extra": "22383 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22383 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22383 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 221.4,
            "unit": "ns/op\t     320 B/op\t       4 allocs/op",
            "extra": "5428267 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 221.4,
            "unit": "ns/op",
            "extra": "5428267 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5428267 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5428267 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 184.2,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "6110270 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 184.2,
            "unit": "ns/op",
            "extra": "6110270 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "6110270 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "6110270 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.872,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642277147 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.872,
            "unit": "ns/op",
            "extra": "642277147 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642277147 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642277147 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.9,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11489322 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.9,
            "unit": "ns/op",
            "extra": "11489322 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11489322 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11489322 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.922,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203034052 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.922,
            "unit": "ns/op",
            "extra": "203034052 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203034052 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203034052 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643130818 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "643130818 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643130818 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643130818 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.87,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28148588 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.87,
            "unit": "ns/op",
            "extra": "28148588 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28148588 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28148588 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66692689,
            "unit": "ns/op\t    6452 B/op\t      36 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66692689,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6452,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66650187,
            "unit": "ns/op\t    5580 B/op\t      18 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66650187,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5580,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "8b2741e9555853b652dc3c6f8fb0b8cafc57bc97",
          "message": "docs: add distributed tracing documentation",
          "timestamp": "2026-01-14T15:02:35+03:00",
          "tree_id": "6d6dc4b8b25970bce98f15ae4315ae35971c6585",
          "url": "https://github.com/PoyrazK/thecloud/commit/8b2741e9555853b652dc3c6f8fb0b8cafc57bc97"
        },
        "date": 1768392359419,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.877,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "621228466 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.877,
            "unit": "ns/op",
            "extra": "621228466 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "621228466 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "621228466 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.4,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8903179 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.4,
            "unit": "ns/op",
            "extra": "8903179 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8903179 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8903179 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 752.5,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1610607 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 752.5,
            "unit": "ns/op",
            "extra": "1610607 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1610607 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1610607 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55875,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22112 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55875,
            "unit": "ns/op",
            "extra": "22112 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22112 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22112 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 549.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2212874 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 549.2,
            "unit": "ns/op",
            "extra": "2212874 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2212874 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2212874 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 407.6,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2740900 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 407.6,
            "unit": "ns/op",
            "extra": "2740900 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2740900 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2740900 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "638357714 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "638357714 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "638357714 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "638357714 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.9,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11457904 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.9,
            "unit": "ns/op",
            "extra": "11457904 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11457904 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11457904 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.311,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "219603164 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.311,
            "unit": "ns/op",
            "extra": "219603164 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "219603164 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "219603164 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "635407611 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "635407611 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "635407611 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "635407611 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.18,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28180231 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.18,
            "unit": "ns/op",
            "extra": "28180231 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28180231 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28180231 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66426886,
            "unit": "ns/op\t    6737 B/op\t      40 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66426886,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6737,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 40,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66471056,
            "unit": "ns/op\t    5884 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66471056,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5884,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "3c9c572587ddecad7c2a89ad13231d4b33591f32",
          "message": "style: add package comment to tracing package",
          "timestamp": "2026-01-14T15:09:03+03:00",
          "tree_id": "ca226106c41b326295f9ad9c4afd29a3e5e1f3fd",
          "url": "https://github.com/PoyrazK/thecloud/commit/3c9c572587ddecad7c2a89ad13231d4b33591f32"
        },
        "date": 1768392642019,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "600225164 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "600225164 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "600225164 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "600225164 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 133.5,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8986514 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 133.5,
            "unit": "ns/op",
            "extra": "8986514 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8986514 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8986514 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 743.7,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1618098 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 743.7,
            "unit": "ns/op",
            "extra": "1618098 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1618098 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1618098 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54767,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22431 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54767,
            "unit": "ns/op",
            "extra": "22431 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22431 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22431 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 618.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2325867 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 618.2,
            "unit": "ns/op",
            "extra": "2325867 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2325867 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2325867 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 386.3,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2847129 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 386.3,
            "unit": "ns/op",
            "extra": "2847129 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2847129 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2847129 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.889,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "638697846 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.889,
            "unit": "ns/op",
            "extra": "638697846 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "638697846 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "638697846 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 102,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11944017 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 102,
            "unit": "ns/op",
            "extra": "11944017 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11944017 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11944017 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.357,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226562956 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.357,
            "unit": "ns/op",
            "extra": "226562956 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226562956 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226562956 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642860588 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642860588 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642860588 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642860588 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.25,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27221367 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.25,
            "unit": "ns/op",
            "extra": "27221367 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27221367 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27221367 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66409091,
            "unit": "ns/op\t    6724 B/op\t      40 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66409091,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6724,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 40,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66334271,
            "unit": "ns/op\t    5884 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66334271,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5884,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "1e5d80c3d1ab9b8a2096e8071e9c723aeedef3f7",
          "message": "chore: disable tracing in benchmarks to avoid performance regression noise",
          "timestamp": "2026-01-14T15:12:39+03:00",
          "tree_id": "b6c76a16d2f59ae8047f5b4376506e40037e4b44",
          "url": "https://github.com/PoyrazK/thecloud/commit/1e5d80c3d1ab9b8a2096e8071e9c723aeedef3f7"
        },
        "date": 1768392852108,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.879,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "637589607 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.879,
            "unit": "ns/op",
            "extra": "637589607 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "637589607 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "637589607 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.2,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8245783 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.2,
            "unit": "ns/op",
            "extra": "8245783 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8245783 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8245783 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 743.1,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1616797 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 743.1,
            "unit": "ns/op",
            "extra": "1616797 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1616797 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1616797 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 56753,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22364 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 56753,
            "unit": "ns/op",
            "extra": "22364 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22364 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22364 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 521.7,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2280888 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 521.7,
            "unit": "ns/op",
            "extra": "2280888 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2280888 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2280888 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 391.7,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2922753 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 391.7,
            "unit": "ns/op",
            "extra": "2922753 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2922753 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2922753 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640039149 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "640039149 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640039149 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640039149 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 105.1,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11592256 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 105.1,
            "unit": "ns/op",
            "extra": "11592256 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11592256 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11592256 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.306,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226424546 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.306,
            "unit": "ns/op",
            "extra": "226424546 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226424546 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226424546 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640125453 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "640125453 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640125453 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640125453 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.61,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28147627 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.61,
            "unit": "ns/op",
            "extra": "28147627 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28147627 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28147627 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66388487,
            "unit": "ns/op\t    6749 B/op\t      40 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66388487,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6749,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 40,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66359358,
            "unit": "ns/op\t    5896 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66359358,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5896,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "20aeedec76039fcf30c090f03e3aafa5660e587c",
          "message": "build: fix docker build by compiling entire cmd/api package",
          "timestamp": "2026-01-14T15:18:59+03:00",
          "tree_id": "d3ec06655d4ec707bf3d1104ac892188dc068fc9",
          "url": "https://github.com/PoyrazK/thecloud/commit/20aeedec76039fcf30c090f03e3aafa5660e587c"
        },
        "date": 1768393229107,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.479,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "790516166 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.479,
            "unit": "ns/op",
            "extra": "790516166 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "790516166 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "790516166 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 143.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8438470 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 143.8,
            "unit": "ns/op",
            "extra": "8438470 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8438470 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8438470 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 716.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1669454 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 716.2,
            "unit": "ns/op",
            "extra": "1669454 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1669454 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1669454 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 28668,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "45852 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 28668,
            "unit": "ns/op",
            "extra": "45852 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "45852 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "45852 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 523.9,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2384478 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 523.9,
            "unit": "ns/op",
            "extra": "2384478 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2384478 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2384478 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 429.7,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2696926 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 429.7,
            "unit": "ns/op",
            "extra": "2696926 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2696926 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2696926 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.497,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "736359650 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.497,
            "unit": "ns/op",
            "extra": "736359650 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "736359650 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "736359650 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 123.2,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "10138562 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 123.2,
            "unit": "ns/op",
            "extra": "10138562 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "10138562 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "10138562 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 3.807,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "307538065 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 3.807,
            "unit": "ns/op",
            "extra": "307538065 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "307538065 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "307538065 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.608,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "756342379 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.608,
            "unit": "ns/op",
            "extra": "756342379 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "756342379 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "756342379 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.08,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "26325861 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.08,
            "unit": "ns/op",
            "extra": "26325861 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "26325861 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "26325861 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 63553612,
            "unit": "ns/op\t    6724 B/op\t      40 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 63553612,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 6724,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 40,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 63613712,
            "unit": "ns/op\t    5884 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 63613712,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5884,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "be7f1df9e35667d45312e5b047608fdf3ebe68e8",
          "message": "feat: implement security groups with full CLI, tracing, and auditing",
          "timestamp": "2026-01-14T12:19:05Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/20/commits/be7f1df9e35667d45312e5b047608fdf3ebe68e8"
        },
        "date": 1768409817411,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641663461 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641663461 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641663461 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641663461 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.5,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8782324 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.5,
            "unit": "ns/op",
            "extra": "8782324 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8782324 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8782324 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 753.1,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1585701 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 753.1,
            "unit": "ns/op",
            "extra": "1585701 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1585701 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1585701 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54666,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21709 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54666,
            "unit": "ns/op",
            "extra": "21709 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21709 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21709 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 560.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2337598 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 560.2,
            "unit": "ns/op",
            "extra": "2337598 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2337598 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2337598 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 386.4,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2710261 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 386.4,
            "unit": "ns/op",
            "extra": "2710261 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2710261 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2710261 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "631622355 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "631622355 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "631622355 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "631622355 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 108.1,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11449443 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 108.1,
            "unit": "ns/op",
            "extra": "11449443 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11449443 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11449443 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.295,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "225989094 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.295,
            "unit": "ns/op",
            "extra": "225989094 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "225989094 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "225989094 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642454411 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "642454411 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642454411 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642454411 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.28,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27408553 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.28,
            "unit": "ns/op",
            "extra": "27408553 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27408553 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27408553 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66359785,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66359785,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66374362,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66374362,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "c6854f1bac60f38e6ef2d4893e1b6361ef59668e",
          "message": "feat: implement security groups with full CLI, tracing, and auditing",
          "timestamp": "2026-01-14T12:19:05Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/20/commits/c6854f1bac60f38e6ef2d4893e1b6361ef59668e"
        },
        "date": 1768412252478,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.857,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "651069351 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.857,
            "unit": "ns/op",
            "extra": "651069351 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "651069351 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "651069351 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 146.5,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8015066 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 146.5,
            "unit": "ns/op",
            "extra": "8015066 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8015066 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8015066 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 820.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1617493 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 820.2,
            "unit": "ns/op",
            "extra": "1617493 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1617493 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1617493 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54965,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22159 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54965,
            "unit": "ns/op",
            "extra": "22159 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22159 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22159 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 518.9,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2328918 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 518.9,
            "unit": "ns/op",
            "extra": "2328918 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2328918 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2328918 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 378.8,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2943529 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 378.8,
            "unit": "ns/op",
            "extra": "2943529 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2943529 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2943529 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.862,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640715886 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.862,
            "unit": "ns/op",
            "extra": "640715886 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640715886 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640715886 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 102.1,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12104859 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 102.1,
            "unit": "ns/op",
            "extra": "12104859 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12104859 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12104859 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.294,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226612580 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.294,
            "unit": "ns/op",
            "extra": "226612580 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226612580 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226612580 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.864,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641826607 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.864,
            "unit": "ns/op",
            "extra": "641826607 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641826607 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641826607 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.87,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "29593526 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.87,
            "unit": "ns/op",
            "extra": "29593526 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "29593526 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "29593526 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66244402,
            "unit": "ns/op\t    7280 B/op\t      42 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66244402,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7280,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66100799,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66100799,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "21dd617a815d523218f81bce9b10b0368dee78ff",
          "message": "feat: implement security groups with full CLI, tracing, and auditing",
          "timestamp": "2026-01-14T12:19:05Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/20/commits/21dd617a815d523218f81bce9b10b0368dee78ff"
        },
        "date": 1768412747419,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641033012 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641033012 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641033012 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641033012 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 148.7,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8181444 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 148.7,
            "unit": "ns/op",
            "extra": "8181444 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8181444 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8181444 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 756.9,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1583419 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 756.9,
            "unit": "ns/op",
            "extra": "1583419 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1583419 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1583419 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55152,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22107 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55152,
            "unit": "ns/op",
            "extra": "22107 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22107 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22107 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 626.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2314844 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 626.2,
            "unit": "ns/op",
            "extra": "2314844 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2314844 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2314844 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 391.3,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2832345 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 391.3,
            "unit": "ns/op",
            "extra": "2832345 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2832345 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2832345 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641191339 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "641191339 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641191339 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641191339 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104.6,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11807146 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104.6,
            "unit": "ns/op",
            "extra": "11807146 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11807146 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11807146 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.294,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "225120096 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.294,
            "unit": "ns/op",
            "extra": "225120096 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "225120096 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "225120096 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.865,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642385090 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.865,
            "unit": "ns/op",
            "extra": "642385090 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642385090 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642385090 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.59,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27907639 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.59,
            "unit": "ns/op",
            "extra": "27907639 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27907639 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27907639 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66324687,
            "unit": "ns/op\t    7263 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66324687,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7263,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66328359,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66328359,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "96b579ea625ebda293ea970a4e49778be0e87bc0",
          "message": "feat: implement security groups with full CLI, tracing, and auditing",
          "timestamp": "2026-01-14T12:19:05Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/20/commits/96b579ea625ebda293ea970a4e49778be0e87bc0"
        },
        "date": 1768413201537,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641932366 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "641932366 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641932366 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641932366 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 149.9,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8080378 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 149.9,
            "unit": "ns/op",
            "extra": "8080378 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8080378 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8080378 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 744.6,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1609080 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 744.6,
            "unit": "ns/op",
            "extra": "1609080 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1609080 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1609080 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55383,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22560 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55383,
            "unit": "ns/op",
            "extra": "22560 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22560 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22560 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 539.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2221767 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 539.2,
            "unit": "ns/op",
            "extra": "2221767 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2221767 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2221767 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 397.1,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2806665 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 397.1,
            "unit": "ns/op",
            "extra": "2806665 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2806665 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2806665 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.868,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "637867935 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.868,
            "unit": "ns/op",
            "extra": "637867935 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "637867935 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "637867935 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12028321 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104.7,
            "unit": "ns/op",
            "extra": "12028321 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12028321 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12028321 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.294,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226171646 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.294,
            "unit": "ns/op",
            "extra": "226171646 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226171646 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226171646 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641712526 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "641712526 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641712526 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641712526 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.14,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28560248 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.14,
            "unit": "ns/op",
            "extra": "28560248 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28560248 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28560248 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66348428,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66348428,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66759586,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66759586,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "cee90921a085aa6de7671005179f6eb2a8a4bfd3",
          "message": "feat: implement security groups with full CLI, tracing, and auditing",
          "timestamp": "2026-01-14T12:19:05Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/20/commits/cee90921a085aa6de7671005179f6eb2a8a4bfd3"
        },
        "date": 1768413855601,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.876,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639293523 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.876,
            "unit": "ns/op",
            "extra": "639293523 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639293523 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639293523 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 149.9,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8035855 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 149.9,
            "unit": "ns/op",
            "extra": "8035855 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8035855 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8035855 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 751.4,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1593921 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 751.4,
            "unit": "ns/op",
            "extra": "1593921 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1593921 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1593921 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55568,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21622 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55568,
            "unit": "ns/op",
            "extra": "21622 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21622 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21622 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 599.7,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1876441 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 599.7,
            "unit": "ns/op",
            "extra": "1876441 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1876441 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1876441 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 398.5,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2848464 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 398.5,
            "unit": "ns/op",
            "extra": "2848464 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2848464 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2848464 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.873,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640670328 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.873,
            "unit": "ns/op",
            "extra": "640670328 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640670328 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640670328 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 106.9,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11663822 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 106.9,
            "unit": "ns/op",
            "extra": "11663822 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11663822 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11663822 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.297,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226355983 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.297,
            "unit": "ns/op",
            "extra": "226355983 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226355983 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226355983 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.88,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641068710 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.88,
            "unit": "ns/op",
            "extra": "641068710 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641068710 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641068710 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.6,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28383962 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.6,
            "unit": "ns/op",
            "extra": "28383962 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28383962 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28383962 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66338662,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66338662,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66333153,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66333153,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "e573dc9ad1d0fe2a2cc8eeeaeebc3cf0609e7cf7",
          "message": "feat: implement security groups with full CLI, tracing, and auditing",
          "timestamp": "2026-01-14T12:19:05Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/20/commits/e573dc9ad1d0fe2a2cc8eeeaeebc3cf0609e7cf7"
        },
        "date": 1768414144077,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642337993 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "642337993 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642337993 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642337993 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 165.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8061337 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 165.3,
            "unit": "ns/op",
            "extra": "8061337 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8061337 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8061337 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 751.1,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1597176 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 751.1,
            "unit": "ns/op",
            "extra": "1597176 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1597176 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1597176 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55686,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21762 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55686,
            "unit": "ns/op",
            "extra": "21762 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21762 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21762 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 540.7,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2227640 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 540.7,
            "unit": "ns/op",
            "extra": "2227640 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2227640 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2227640 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 403.4,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2798652 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 403.4,
            "unit": "ns/op",
            "extra": "2798652 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2798652 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2798652 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639675405 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "639675405 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639675405 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639675405 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.8,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "10352839 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.8,
            "unit": "ns/op",
            "extra": "10352839 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "10352839 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "10352839 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.294,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "225188523 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.294,
            "unit": "ns/op",
            "extra": "225188523 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "225188523 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "225188523 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640745960 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "640745960 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640745960 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640745960 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.46,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27677964 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.46,
            "unit": "ns/op",
            "extra": "27677964 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27677964 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27677964 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66475389,
            "unit": "ns/op\t    7266 B/op\t      42 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66475389,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7266,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66397926,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66397926,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Hüseyin Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "013aed8e7ede9ed654c5aa2d2ca9e563ec82c420",
          "message": "Merge pull request #20 from PoyrazK/feature/security-groups\n\nfeat: implement security groups with full CLI, tracing, and auditing",
          "timestamp": "2026-01-14T21:14:48+03:00",
          "tree_id": "95550ae71190c1e05b65539da595db5cfcd92b2b",
          "url": "https://github.com/PoyrazK/thecloud/commit/013aed8e7ede9ed654c5aa2d2ca9e563ec82c420"
        },
        "date": 1768414577918,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.877,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642283191 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.877,
            "unit": "ns/op",
            "extra": "642283191 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642283191 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642283191 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 148.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "7980336 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 148.3,
            "unit": "ns/op",
            "extra": "7980336 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "7980336 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7980336 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 819.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1615827 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 819.2,
            "unit": "ns/op",
            "extra": "1615827 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1615827 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1615827 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55143,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22251 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55143,
            "unit": "ns/op",
            "extra": "22251 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22251 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22251 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 509.9,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2380623 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 509.9,
            "unit": "ns/op",
            "extra": "2380623 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2380623 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2380623 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 396.3,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2929934 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 396.3,
            "unit": "ns/op",
            "extra": "2929934 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2929934 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2929934 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642770839 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "642770839 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642770839 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642770839 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 102.2,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12246570 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 102.2,
            "unit": "ns/op",
            "extra": "12246570 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12246570 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12246570 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.294,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226264683 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.294,
            "unit": "ns/op",
            "extra": "226264683 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226264683 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226264683 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.872,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641878996 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.872,
            "unit": "ns/op",
            "extra": "641878996 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641878996 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641878996 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.36,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27785899 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.36,
            "unit": "ns/op",
            "extra": "27785899 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27785899 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27785899 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66345143,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66345143,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66664760,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66664760,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "47e80bb4ae3f119dec23c7503c37f266284c6dfd",
          "message": "docs: improve docstring coverage",
          "timestamp": "2026-01-14T18:15:38Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/21/commits/47e80bb4ae3f119dec23c7503c37f266284c6dfd"
        },
        "date": 1768419343919,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.88,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642183217 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.88,
            "unit": "ns/op",
            "extra": "642183217 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642183217 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642183217 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 147.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8061924 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 147.8,
            "unit": "ns/op",
            "extra": "8061924 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8061924 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8061924 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 742.4,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1609654 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 742.4,
            "unit": "ns/op",
            "extra": "1609654 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1609654 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1609654 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54725,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22366 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54725,
            "unit": "ns/op",
            "extra": "22366 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22366 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22366 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 532.5,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2278430 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 532.5,
            "unit": "ns/op",
            "extra": "2278430 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2278430 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2278430 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 395.8,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2941281 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 395.8,
            "unit": "ns/op",
            "extra": "2941281 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2941281 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2941281 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.879,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642404965 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.879,
            "unit": "ns/op",
            "extra": "642404965 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642404965 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642404965 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 103.3,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11610859 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 103.3,
            "unit": "ns/op",
            "extra": "11610859 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11610859 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11610859 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.323,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "226668764 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.323,
            "unit": "ns/op",
            "extra": "226668764 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "226668764 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "226668764 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.873,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642339086 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.873,
            "unit": "ns/op",
            "extra": "642339086 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642339086 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642339086 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.25,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27555766 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.25,
            "unit": "ns/op",
            "extra": "27555766 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27555766 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27555766 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66339092,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66339092,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66325129,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66325129,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Hüseyin Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "494d5408d2f798706295d6a7df3c29f89fdae89b",
          "message": "Merge pull request #21 from PoyrazK/feature/docstrings-coverage\n\ndocs: improve docstring coverage",
          "timestamp": "2026-01-14T22:49:29+03:00",
          "tree_id": "ce294ed7be435400d43a55d026adcfb379c4971c",
          "url": "https://github.com/PoyrazK/thecloud/commit/494d5408d2f798706295d6a7df3c29f89fdae89b"
        },
        "date": 1768420259818,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.478,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "825043706 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.478,
            "unit": "ns/op",
            "extra": "825043706 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "825043706 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "825043706 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 133.6,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "9002124 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 133.6,
            "unit": "ns/op",
            "extra": "9002124 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "9002124 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9002124 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 719.7,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1668280 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 719.7,
            "unit": "ns/op",
            "extra": "1668280 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1668280 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1668280 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 28412,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "46990 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 28412,
            "unit": "ns/op",
            "extra": "46990 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "46990 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "46990 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 509.5,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2377851 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 509.5,
            "unit": "ns/op",
            "extra": "2377851 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2377851 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2377851 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 411,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2843474 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 411,
            "unit": "ns/op",
            "extra": "2843474 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2843474 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2843474 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.992,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "594994371 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.992,
            "unit": "ns/op",
            "extra": "594994371 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "594994371 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "594994371 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 122.4,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "9703719 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 122.4,
            "unit": "ns/op",
            "extra": "9703719 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "9703719 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "9703719 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 3.809,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "309002712 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 3.809,
            "unit": "ns/op",
            "extra": "309002712 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "309002712 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "309002712 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.67,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "676351363 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.67,
            "unit": "ns/op",
            "extra": "676351363 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "676351363 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "676351363 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 40.78,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "29118956 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 40.78,
            "unit": "ns/op",
            "extra": "29118956 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "29118956 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "29118956 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 63649750,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 63649750,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 63591039,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 63591039,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "52ed87e47b45d59c6a9a7d905461993920e4ee88",
          "message": "feat: Improve test coverage to 75.2% with libvirt adapter refactor",
          "timestamp": "2026-01-14T19:49:38Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/22/commits/52ed87e47b45d59c6a9a7d905461993920e4ee88"
        },
        "date": 1768583978543,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.887,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639927655 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.887,
            "unit": "ns/op",
            "extra": "639927655 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639927655 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639927655 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 133.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8885502 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 133.3,
            "unit": "ns/op",
            "extra": "8885502 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8885502 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8885502 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 743.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1613294 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 743.2,
            "unit": "ns/op",
            "extra": "1613294 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1613294 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1613294 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55720,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22248 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55720,
            "unit": "ns/op",
            "extra": "22248 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22248 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22248 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 524,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2304854 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 524,
            "unit": "ns/op",
            "extra": "2304854 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2304854 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2304854 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 427.1,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2437752 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 427.1,
            "unit": "ns/op",
            "extra": "2437752 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2437752 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2437752 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641614173 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641614173 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641614173 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641614173 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 103.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11547488 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 103.7,
            "unit": "ns/op",
            "extra": "11547488 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11547488 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11547488 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.956,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203044713 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.956,
            "unit": "ns/op",
            "extra": "203044713 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203044713 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203044713 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641875525 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "641875525 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641875525 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641875525 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.07,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28477280 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.07,
            "unit": "ns/op",
            "extra": "28477280 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28477280 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28477280 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66319784,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66319784,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66369871,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66369871,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "d8a60354a887c1bb8e205e0157e103ce808b455c",
          "message": "feat: Improve test coverage to 75.2% with libvirt adapter refactor",
          "timestamp": "2026-01-14T19:49:38Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/22/commits/d8a60354a887c1bb8e205e0157e103ce808b455c"
        },
        "date": 1768584251267,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639376816 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "639376816 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639376816 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639376816 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 136.3,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8920716 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 136.3,
            "unit": "ns/op",
            "extra": "8920716 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8920716 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8920716 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 765.1,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1583580 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 765.1,
            "unit": "ns/op",
            "extra": "1583580 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1583580 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1583580 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55882,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22077 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55882,
            "unit": "ns/op",
            "extra": "22077 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22077 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22077 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 539.3,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2256122 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 539.3,
            "unit": "ns/op",
            "extra": "2256122 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2256122 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2256122 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 400.1,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2782233 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 400.1,
            "unit": "ns/op",
            "extra": "2782233 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2782233 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2782233 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641739609 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641739609 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641739609 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641739609 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 109.5,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11391732 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 109.5,
            "unit": "ns/op",
            "extra": "11391732 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11391732 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11391732 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.923,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202868976 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.923,
            "unit": "ns/op",
            "extra": "202868976 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202868976 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202868976 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639033950 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "639033950 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639033950 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639033950 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.73,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28574287 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.73,
            "unit": "ns/op",
            "extra": "28574287 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28574287 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28574287 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66477213,
            "unit": "ns/op\t    7263 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66477213,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7263,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66384131,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66384131,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "4a10c7bdaaa4c00d869d40f4bcc6d77bced660fe",
          "message": "feat: Improve test coverage to 75.2% with libvirt adapter refactor",
          "timestamp": "2026-01-14T19:49:38Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/22/commits/4a10c7bdaaa4c00d869d40f4bcc6d77bced660fe"
        },
        "date": 1768585129978,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640150904 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "640150904 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640150904 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640150904 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8928410 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.1,
            "unit": "ns/op",
            "extra": "8928410 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8928410 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8928410 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 817.3,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1453594 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 817.3,
            "unit": "ns/op",
            "extra": "1453594 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1453594 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1453594 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55193,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22413 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55193,
            "unit": "ns/op",
            "extra": "22413 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22413 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22413 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 539.6,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2278334 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 539.6,
            "unit": "ns/op",
            "extra": "2278334 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2278334 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2278334 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 386.2,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2875334 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 386.2,
            "unit": "ns/op",
            "extra": "2875334 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2875334 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2875334 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.886,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640529736 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.886,
            "unit": "ns/op",
            "extra": "640529736 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640529736 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640529736 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 102.4,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12183003 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 102.4,
            "unit": "ns/op",
            "extra": "12183003 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12183003 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12183003 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.915,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202879760 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.915,
            "unit": "ns/op",
            "extra": "202879760 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202879760 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202879760 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642530823 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642530823 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642530823 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642530823 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.15,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28092739 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.15,
            "unit": "ns/op",
            "extra": "28092739 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28092739 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28092739 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66395700,
            "unit": "ns/op\t    7263 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66395700,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7263,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66429413,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66429413,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "01d7fc81535c4f4d2f022fcf072a284a2f0a608a",
          "message": "feat: Improve test coverage to 75.2% with libvirt adapter refactor",
          "timestamp": "2026-01-14T19:49:38Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/22/commits/01d7fc81535c4f4d2f022fcf072a284a2f0a608a"
        },
        "date": 1768585190219,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.881,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "636981096 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.881,
            "unit": "ns/op",
            "extra": "636981096 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "636981096 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "636981096 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 146.7,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8817543 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 146.7,
            "unit": "ns/op",
            "extra": "8817543 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8817543 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8817543 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 746.6,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1438429 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 746.6,
            "unit": "ns/op",
            "extra": "1438429 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1438429 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1438429 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55579,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21961 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55579,
            "unit": "ns/op",
            "extra": "21961 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21961 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21961 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 542.4,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2191995 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 542.4,
            "unit": "ns/op",
            "extra": "2191995 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2191995 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2191995 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 399.7,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2792782 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 399.7,
            "unit": "ns/op",
            "extra": "2792782 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2792782 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2792782 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.885,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "636069608 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.885,
            "unit": "ns/op",
            "extra": "636069608 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "636069608 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "636069608 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11537114 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.7,
            "unit": "ns/op",
            "extra": "11537114 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11537114 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11537114 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.919,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203168775 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.919,
            "unit": "ns/op",
            "extra": "203168775 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203168775 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203168775 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.878,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642121596 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.878,
            "unit": "ns/op",
            "extra": "642121596 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642121596 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642121596 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.77,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28193790 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.77,
            "unit": "ns/op",
            "extra": "28193790 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28193790 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28193790 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66354348,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66354348,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66366173,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66366173,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "83272398+PoyrazK@users.noreply.github.com",
            "name": "Hüseyin Poyraz Küçükarslan",
            "username": "PoyrazK"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2d203af6b40e5b55cc377db04fc41020a0677d3c",
          "message": "Merge pull request #22 from PoyrazK/feature/coverage-80\n\nfeat: Improve test coverage to 75.2% with libvirt adapter refactor",
          "timestamp": "2026-01-16T20:43:50+03:00",
          "tree_id": "4d1b2b970a83391904d58f92cfebb5f6bc33db10",
          "url": "https://github.com/PoyrazK/thecloud/commit/2d203af6b40e5b55cc377db04fc41020a0677d3c"
        },
        "date": 1768585520953,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642743089 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642743089 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642743089 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642743089 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8858239 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.1,
            "unit": "ns/op",
            "extra": "8858239 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8858239 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8858239 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 748.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1599828 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 748.2,
            "unit": "ns/op",
            "extra": "1599828 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1599828 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1599828 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55093,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22272 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55093,
            "unit": "ns/op",
            "extra": "22272 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22272 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22272 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 617.6,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2334907 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 617.6,
            "unit": "ns/op",
            "extra": "2334907 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2334907 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2334907 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 392.3,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2755796 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 392.3,
            "unit": "ns/op",
            "extra": "2755796 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2755796 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2755796 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643138970 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "643138970 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643138970 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643138970 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 104.4,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11840431 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 104.4,
            "unit": "ns/op",
            "extra": "11840431 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11840431 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11840431 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.918,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202888202 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.918,
            "unit": "ns/op",
            "extra": "202888202 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202888202 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202888202 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641789346 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641789346 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641789346 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641789346 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.54,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27615828 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.54,
            "unit": "ns/op",
            "extra": "27615828 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27615828 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27615828 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66363636,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66363636,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 67018127,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 67018127,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "b39231d36aca6b67cdd92f3a58237f1611860eca",
          "message": "feat: improve test coverage to 80% and fix CI/CD issues",
          "timestamp": "2026-01-16T17:43:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/24/commits/b39231d36aca6b67cdd92f3a58237f1611860eca"
        },
        "date": 1768657818755,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640994967 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "640994967 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640994967 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640994967 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8770881 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134,
            "unit": "ns/op",
            "extra": "8770881 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8770881 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8770881 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 760.8,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1610610 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 760.8,
            "unit": "ns/op",
            "extra": "1610610 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1610610 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1610610 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55850,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "21860 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55850,
            "unit": "ns/op",
            "extra": "21860 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "21860 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "21860 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 559.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2172824 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 559.2,
            "unit": "ns/op",
            "extra": "2172824 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2172824 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2172824 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 401.4,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2712378 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 401.4,
            "unit": "ns/op",
            "extra": "2712378 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2712378 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2712378 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.878,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639414844 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.878,
            "unit": "ns/op",
            "extra": "639414844 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639414844 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639414844 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 105.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11869928 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 105.7,
            "unit": "ns/op",
            "extra": "11869928 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11869928 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11869928 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.947,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203205867 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.947,
            "unit": "ns/op",
            "extra": "203205867 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203205867 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203205867 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641137041 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "641137041 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641137041 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641137041 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.54,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28278234 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.54,
            "unit": "ns/op",
            "extra": "28278234 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28278234 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28278234 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66380895,
            "unit": "ns/op\t    7287 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66380895,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7287,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66344550,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66344550,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "9a70d2bfe08f4bf095b2f30f8daae6fb60a43cb8",
          "message": "feat: improve test coverage to 80% and fix CI/CD issues",
          "timestamp": "2026-01-16T17:43:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/24/commits/9a70d2bfe08f4bf095b2f30f8daae6fb60a43cb8"
        },
        "date": 1768658289506,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.882,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641585454 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.882,
            "unit": "ns/op",
            "extra": "641585454 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641585454 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641585454 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.9,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8804306 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.9,
            "unit": "ns/op",
            "extra": "8804306 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8804306 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8804306 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 743.6,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1607203 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 743.6,
            "unit": "ns/op",
            "extra": "1607203 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1607203 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1607203 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55106,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22508 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55106,
            "unit": "ns/op",
            "extra": "22508 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22508 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22508 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 578.6,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2215963 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 578.6,
            "unit": "ns/op",
            "extra": "2215963 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2215963 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2215963 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 397.6,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2639737 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 397.6,
            "unit": "ns/op",
            "extra": "2639737 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2639737 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2639737 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.896,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "640833907 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.896,
            "unit": "ns/op",
            "extra": "640833907 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "640833907 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "640833907 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.9,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11260342 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.9,
            "unit": "ns/op",
            "extra": "11260342 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11260342 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11260342 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.915,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202800898 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.915,
            "unit": "ns/op",
            "extra": "202800898 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202800898 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202800898 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.874,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642893565 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.874,
            "unit": "ns/op",
            "extra": "642893565 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642893565 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642893565 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.5,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28028391 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.5,
            "unit": "ns/op",
            "extra": "28028391 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28028391 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28028391 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66393989,
            "unit": "ns/op\t    7263 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66393989,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7263,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 67021275,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 67021275,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "a44a406f49afaca2e9aa18e90aa735e65add837b",
          "message": "feat: improve test coverage to 80% and fix CI/CD issues",
          "timestamp": "2026-01-16T17:43:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/24/commits/a44a406f49afaca2e9aa18e90aa735e65add837b"
        },
        "date": 1768659211014,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.881,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "619092151 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.881,
            "unit": "ns/op",
            "extra": "619092151 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "619092151 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "619092151 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8898478 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.1,
            "unit": "ns/op",
            "extra": "8898478 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8898478 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8898478 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 780,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1597621 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 780,
            "unit": "ns/op",
            "extra": "1597621 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1597621 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1597621 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55484,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22195 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55484,
            "unit": "ns/op",
            "extra": "22195 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22195 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22195 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 531.6,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2287112 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 531.6,
            "unit": "ns/op",
            "extra": "2287112 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2287112 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2287112 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 406.1,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2831923 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 406.1,
            "unit": "ns/op",
            "extra": "2831923 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2831923 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2831923 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642447632 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "642447632 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642447632 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642447632 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 106,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "12035826 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 106,
            "unit": "ns/op",
            "extra": "12035826 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12035826 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12035826 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.914,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202796814 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.914,
            "unit": "ns/op",
            "extra": "202796814 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202796814 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202796814 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.883,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641620636 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.883,
            "unit": "ns/op",
            "extra": "641620636 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641620636 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641620636 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 43.5,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28381401 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 43.5,
            "unit": "ns/op",
            "extra": "28381401 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28381401 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28381401 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66339901,
            "unit": "ns/op\t    7280 B/op\t      42 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66339901,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7280,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66327506,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66327506,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "97d9f43a388c9958c2104a84759ef0ce2c72ec65",
          "message": "feat: improve test coverage to 80% and fix CI/CD issues",
          "timestamp": "2026-01-16T17:43:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/24/commits/97d9f43a388c9958c2104a84759ef0ce2c72ec65"
        },
        "date": 1768659282121,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.879,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "614287711 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.879,
            "unit": "ns/op",
            "extra": "614287711 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "614287711 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "614287711 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 134.8,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8772016 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 134.8,
            "unit": "ns/op",
            "extra": "8772016 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8772016 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8772016 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 749.1,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1603573 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 749.1,
            "unit": "ns/op",
            "extra": "1603573 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1603573 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1603573 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55117,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22262 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55117,
            "unit": "ns/op",
            "extra": "22262 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22262 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22262 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 519,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2314438 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 519,
            "unit": "ns/op",
            "extra": "2314438 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2314438 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2314438 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 381.8,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2826007 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 381.8,
            "unit": "ns/op",
            "extra": "2826007 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2826007 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2826007 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.873,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "632838748 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.873,
            "unit": "ns/op",
            "extra": "632838748 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "632838748 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "632838748 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 103.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11487519 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 103.7,
            "unit": "ns/op",
            "extra": "11487519 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11487519 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11487519 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.918,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202646661 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.918,
            "unit": "ns/op",
            "extra": "202646661 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202646661 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202646661 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642830858 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "642830858 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642830858 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642830858 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.12,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27759024 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.12,
            "unit": "ns/op",
            "extra": "27759024 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27759024 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27759024 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66351416,
            "unit": "ns/op\t    7266 B/op\t      42 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66351416,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7266,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66382931,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66382931,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
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
          "id": "8149fcab65df70f27702371b9e2540847af72e3b",
          "message": "feat: improve test coverage to 80% and fix CI/CD issues",
          "timestamp": "2026-01-16T17:43:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/24/commits/8149fcab65df70f27702371b9e2540847af72e3b"
        },
        "date": 1768660592157,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.871,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "643536021 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.871,
            "unit": "ns/op",
            "extra": "643536021 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "643536021 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "643536021 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "7761681 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.1,
            "unit": "ns/op",
            "extra": "7761681 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "7761681 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7761681 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 748.7,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1589554 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 748.7,
            "unit": "ns/op",
            "extra": "1589554 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1589554 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1589554 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 56155,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22236 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 56155,
            "unit": "ns/op",
            "extra": "22236 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22236 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22236 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 534.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2228941 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 534.2,
            "unit": "ns/op",
            "extra": "2228941 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2228941 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2228941 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 431.6,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2466076 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 431.6,
            "unit": "ns/op",
            "extra": "2466076 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2466076 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2466076 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "639407376 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "639407376 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "639407376 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "639407376 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107.4,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11416648 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107.4,
            "unit": "ns/op",
            "extra": "11416648 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11416648 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11416648 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.945,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "201357669 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.945,
            "unit": "ns/op",
            "extra": "201357669 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "201357669 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "201357669 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642494416 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "642494416 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642494416 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642494416 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 43.1,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28416703 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 43.1,
            "unit": "ns/op",
            "extra": "28416703 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28416703 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28416703 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66416094,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66416094,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66795394,
            "unit": "ns/op\t    5944 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66795394,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5944,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "53134913731a7314c7a7236a383e064ea7338e48",
          "message": "feat: improve test coverage to 80% and fix CI/CD issues",
          "timestamp": "2026-01-16T17:43:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/24/commits/53134913731a7314c7a7236a383e064ea7338e48"
        },
        "date": 1768660685979,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.873,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "626967924 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.873,
            "unit": "ns/op",
            "extra": "626967924 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "626967924 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "626967924 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8926467 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135,
            "unit": "ns/op",
            "extra": "8926467 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8926467 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8926467 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 820.5,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1596846 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 820.5,
            "unit": "ns/op",
            "extra": "1596846 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1596846 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1596846 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 54922,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22270 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 54922,
            "unit": "ns/op",
            "extra": "22270 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22270 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22270 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 529,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2276784 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 529,
            "unit": "ns/op",
            "extra": "2276784 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2276784 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2276784 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 406.2,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2824468 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 406.2,
            "unit": "ns/op",
            "extra": "2824468 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2824468 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2824468 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642826178 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "642826178 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642826178 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642826178 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 106.7,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11085204 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 106.7,
            "unit": "ns/op",
            "extra": "11085204 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11085204 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11085204 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.92,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "202636311 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.92,
            "unit": "ns/op",
            "extra": "202636311 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "202636311 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "202636311 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641290936 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "641290936 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641290936 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641290936 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 41.36,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28581186 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 41.36,
            "unit": "ns/op",
            "extra": "28581186 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28581186 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28581186 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66376463,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66376463,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66336285,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66336285,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "0d9a85395cc133d2ca7561fe6edf94a5b978b9a8",
          "message": "feat: improve test coverage to 80% and fix CI/CD issues",
          "timestamp": "2026-01-16T17:43:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/24/commits/0d9a85395cc133d2ca7561fe6edf94a5b978b9a8"
        },
        "date": 1768660716087,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.869,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642757105 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.869,
            "unit": "ns/op",
            "extra": "642757105 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642757105 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642757105 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.7,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8667762 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.7,
            "unit": "ns/op",
            "extra": "8667762 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8667762 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8667762 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 751.3,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1591700 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 751.3,
            "unit": "ns/op",
            "extra": "1591700 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1591700 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1591700 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 55040,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22107 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 55040,
            "unit": "ns/op",
            "extra": "22107 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22107 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22107 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 564.7,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2207206 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 564.7,
            "unit": "ns/op",
            "extra": "2207206 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2207206 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2207206 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 399.5,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2817302 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 399.5,
            "unit": "ns/op",
            "extra": "2817302 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2817302 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2817302 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.896,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "638940234 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.896,
            "unit": "ns/op",
            "extra": "638940234 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "638940234 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "638940234 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 107,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11722458 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 107,
            "unit": "ns/op",
            "extra": "11722458 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11722458 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11722458 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.921,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203335911 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.921,
            "unit": "ns/op",
            "extra": "203335911 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203335911 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203335911 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642090055 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "642090055 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642090055 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642090055 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.63,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "27378600 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.63,
            "unit": "ns/op",
            "extra": "27378600 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "27378600 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "27378600 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66418274,
            "unit": "ns/op\t    7275 B/op\t      42 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66418274,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7275,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66346244,
            "unit": "ns/op\t    5932 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66346244,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5932,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
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
          "id": "826340748afd9a2fc04ccf93d442476dbf5888bc",
          "message": "feat: improve test coverage to 80% and fix CI/CD issues",
          "timestamp": "2026-01-16T17:43:59Z",
          "url": "https://github.com/PoyrazK/thecloud/pull/24/commits/826340748afd9a2fc04ccf93d442476dbf5888bc"
        },
        "date": 1768661018995,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkInstanceServiceList",
            "value": 1.866,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641278204 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - ns/op",
            "value": 1.866,
            "unit": "ns/op",
            "extra": "641278204 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641278204 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641278204 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet",
            "value": 135.1,
            "unit": "ns/op\t     192 B/op\t       2 allocs/op",
            "extra": "8816574 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - ns/op",
            "value": 135.1,
            "unit": "ns/op",
            "extra": "8816574 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "8816574 times\n4 procs"
          },
          {
            "name": "BenchmarkVPCServiceGet - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8816574 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate",
            "value": 753.3,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "1592902 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - ns/op",
            "value": 753.3,
            "unit": "ns/op",
            "extra": "1592902 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "1592902 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1592902 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke",
            "value": 56025,
            "unit": "ns/op\t    1120 B/op\t      20 allocs/op",
            "extra": "22236 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - ns/op",
            "value": 56025,
            "unit": "ns/op",
            "extra": "22236 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - B/op",
            "value": 1120,
            "unit": "B/op",
            "extra": "22236 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceInvoke - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "22236 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel",
            "value": 531.2,
            "unit": "ns/op\t     688 B/op\t       8 allocs/op",
            "extra": "2250798 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - ns/op",
            "value": 531.2,
            "unit": "ns/op",
            "extra": "2250798 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - B/op",
            "value": 688,
            "unit": "B/op",
            "extra": "2250798 times\n4 procs"
          },
          {
            "name": "BenchmarkInstanceServiceCreateParallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "2250798 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel",
            "value": 407.5,
            "unit": "ns/op\t     368 B/op\t       5 allocs/op",
            "extra": "2745766 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - ns/op",
            "value": 407.5,
            "unit": "ns/op",
            "extra": "2745766 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "2745766 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLoginParallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2745766 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList",
            "value": 1.87,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "642900898 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - ns/op",
            "value": 1.87,
            "unit": "ns/op",
            "extra": "642900898 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "642900898 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "642900898 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel",
            "value": 109,
            "unit": "ns/op\t     208 B/op\t       1 allocs/op",
            "extra": "11494856 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - ns/op",
            "value": 109,
            "unit": "ns/op",
            "extra": "11494856 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11494856 times\n4 procs"
          },
          {
            "name": "BenchmarkDatabaseContentionParallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11494856 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList",
            "value": 5.918,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "203146274 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - ns/op",
            "value": 5.918,
            "unit": "ns/op",
            "extra": "203146274 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "203146274 times\n4 procs"
          },
          {
            "name": "BenchmarkCacheServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "203146274 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList",
            "value": 1.867,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "641803875 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - ns/op",
            "value": 1.867,
            "unit": "ns/op",
            "extra": "641803875 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "641803875 times\n4 procs"
          },
          {
            "name": "BenchmarkStorageServiceList - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "641803875 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList",
            "value": 42.36,
            "unit": "ns/op\t      64 B/op\t       1 allocs/op",
            "extra": "28227715 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - ns/op",
            "value": 42.36,
            "unit": "ns/op",
            "extra": "28227715 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - B/op",
            "value": 64,
            "unit": "B/op",
            "extra": "28227715 times\n4 procs"
          },
          {
            "name": "BenchmarkFunctionServiceList - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "28227715 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister",
            "value": 66390095,
            "unit": "ns/op\t    7280 B/op\t      42 allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - ns/op",
            "value": 66390095,
            "unit": "ns/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - B/op",
            "value": 7280,
            "unit": "B/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceRegister - allocs/op",
            "value": 42,
            "unit": "allocs/op",
            "extra": "16 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin",
            "value": 66376725,
            "unit": "ns/op\t    5944 B/op\t      22 allocs/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - ns/op",
            "value": 66376725,
            "unit": "ns/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - B/op",
            "value": 5944,
            "unit": "B/op",
            "extra": "18 times\n4 procs"
          },
          {
            "name": "BenchmarkAuthServiceLogin - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "18 times\n4 procs"
          }
        ]
      }
    ]
  }
}