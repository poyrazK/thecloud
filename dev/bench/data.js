window.BENCHMARK_DATA = {
  "lastUpdate": 1768131748084,
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
      }
    ]
  }
}