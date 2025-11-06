# Benchmark results

```text
ninjacoder@81ddf6307408:/workspaces/snaggle$ go test -bench .
link /workspaces/snaggle/internal/testdata/hello_pie -> /workspaces/snaggle/.tmp/TestLinkDifferentFile_hello_pie_verbose3366495086/bin/hello_pie
link /workspaces/snaggle/internal/testdata/hello_pie -> /workspaces/snaggle/.tmp/TestLinkDifferentFile_hello_pie_relative_verbose1190415337/bin/hello_pie
goos: linux
goarch: amd64
pkg: github.com/MusicalNinjaDad/snaggle
cpu: 13th Gen Intel(R) Core(TM) i7-13700H
BenchmarkCommonBinaries/PIE_1_dependency-20                           67          24305275 ns/op
BenchmarkCommonBinaries/PIE_1_dependency_verbose-20                   28          41187516 ns/op

BenchmarkCommonBinaries/PIE_nested_dependencies-20                    56          22049336 ns/op
BenchmarkCommonBinaries/PIE_nested_dependencies_verbose-20            20          65470394 ns/op

BenchmarkCommonBinaries/Dynamic_library_(.so)-20                      62          19401239 ns/op
BenchmarkCommonBinaries/Dynamic_library_(.so)_verbose-20              28          47009684 ns/op

BenchmarkCommonBinaries/In_subdirectory-20                            54          21983069 ns/op
BenchmarkCommonBinaries/In_subdirectory_verbose-20                    45          35531424 ns/op

BenchmarkCommonBinaries/PIE_no_dependencies-20                        99          11941844 ns/op
BenchmarkCommonBinaries/PIE_no_dependencies_verbose-20               100          11663037 ns/op

BenchmarkCommonBinaries/Static_linked_executable-20                 3122            458260 ns/op
BenchmarkCommonBinaries/Static_linked_executable_verbose-20         3758            408359 ns/op

BenchmarkCommonBinaries/Directory-20                                  70          18701503 ns/op
BenchmarkCommonBinaries/Directory_verbose-20                          12          98306786 ns/op
PASS
ok      github.com/MusicalNinjaDad/snaggle      26.551s
```
