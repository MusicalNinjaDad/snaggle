# Benchmark results

## No goroutines (dc09dff9ef6a14096efb5a401f051f7eac4be74f)

```text
ninjacoder@5747a297e3a1:/workspaces/snaggle$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/MusicalNinjaDad/snaggle
cpu: 13th Gen Intel(R) Core(TM) i7-13700H
BenchmarkCommonBinaries/PIE_no_dependencies-20              1296            888198 ns/op
BenchmarkCommonBinaries/Static_linked_executable-20         4219            275257 ns/op
BenchmarkCommonBinaries/PIE_1_dependency-20                   87          13847470 ns/op
BenchmarkCommonBinaries/PIE_nested_dependencies-20            38          37086494 ns/op
PASS
ok      github.com/MusicalNinjaDad/snaggle      5.404s
ninjacoder@5747a297e3a1:/workspaces/snaggle$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/MusicalNinjaDad/snaggle
cpu: 13th Gen Intel(R) Core(TM) i7-13700H
BenchmarkCommonBinaries/PIE_no_dependencies-20              1321            862166 ns/op
BenchmarkCommonBinaries/Static_linked_executable-20         3900            273573 ns/op
BenchmarkCommonBinaries/PIE_1_dependency-20                   90          13016836 ns/op
BenchmarkCommonBinaries/PIE_nested_dependencies-20            40          32686659 ns/op
PASS
ok      github.com/MusicalNinjaDad/snaggle      5.140s
ninjacoder@5747a297e3a1:/workspaces/snaggle$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/MusicalNinjaDad/snaggle
cpu: 13th Gen Intel(R) Core(TM) i7-13700H
BenchmarkCommonBinaries/PIE_no_dependencies-20              1374            886515 ns/op
BenchmarkCommonBinaries/Static_linked_executable-20         4090            288873 ns/op
BenchmarkCommonBinaries/PIE_1_dependency-20                  104          12534866 ns/op
BenchmarkCommonBinaries/PIE_nested_dependencies-20            38          33616218 ns/op
PASS
ok      github.com/MusicalNinjaDad/snaggle      5.469s
```

## goroutines with errgroup (bd130b640045e4c8f644353aeb05dcb482242f88)

```text
ninjacoder@5747a297e3a1:/workspaces/snaggle$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/MusicalNinjaDad/snaggle
cpu: 13th Gen Intel(R) Core(TM) i7-13700H
BenchmarkCommonBinaries/PIE_no_dependencies-20              1389            860392 ns/op
BenchmarkCommonBinaries/Static_linked_executable-20         3597            324082 ns/op
BenchmarkCommonBinaries/PIE_1_dependency-20                   94          12484150 ns/op
BenchmarkCommonBinaries/PIE_nested_dependencies-20            79          15036758 ns/op
PASS
ok      github.com/MusicalNinjaDad/snaggle      5.182s
ninjacoder@5747a297e3a1:/workspaces/snaggle$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/MusicalNinjaDad/snaggle
cpu: 13th Gen Intel(R) Core(TM) i7-13700H
BenchmarkCommonBinaries/PIE_no_dependencies-20              1334            875738 ns/op
BenchmarkCommonBinaries/Static_linked_executable-20         3555            315325 ns/op
BenchmarkCommonBinaries/PIE_1_dependency-20                   98          12411629 ns/op
BenchmarkCommonBinaries/PIE_nested_dependencies-20            67          15240031 ns/op
PASS
ok      github.com/MusicalNinjaDad/snaggle      4.977s
ninjacoder@5747a297e3a1:/workspaces/snaggle$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/MusicalNinjaDad/snaggle
cpu: 13th Gen Intel(R) Core(TM) i7-13700H
BenchmarkCommonBinaries/PIE_no_dependencies-20              1441            874761 ns/op
BenchmarkCommonBinaries/Static_linked_executable-20         3718            313101 ns/op
BenchmarkCommonBinaries/PIE_1_dependency-20                   85          13472919 ns/op
BenchmarkCommonBinaries/PIE_nested_dependencies-20            78          15378742 ns/op
PASS
ok      github.com/MusicalNinjaDad/snaggle      5.234s
```
