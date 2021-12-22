window.BENCHMARK_DATA = {
  "lastUpdate": 1640211427938,
  "repoUrl": "https://github.com/devnw/stream",
  "entries": {
    "Benchmark Results": [
      {
        "commit": {
          "author": {
            "email": "benji@devnw.com",
            "name": "Benji Vesterby",
            "username": "benjivesterby"
          },
          "committer": {
            "email": "benji@devnw.com",
            "name": "Benji Vesterby",
            "username": "benjivesterby"
          },
          "distinct": true,
          "id": "276d77ffdbe0617ab1be0dfe3eb6594789d1f9d5",
          "message": "Added benchmarks",
          "timestamp": "2021-12-22T17:16:27-05:00",
          "tree_id": "53e1c04b1581b48ef10a5bf29ae112d208d534c3",
          "url": "https://github.com/devnw/stream/commit/276d77ffdbe0617ab1be0dfe3eb6594789d1f9d5"
        },
        "date": 1640211427057,
        "tool": "go",
        "benches": [
          {
            "name": "Benchmark_Pipe",
            "value": 431.7,
            "unit": "ns/op",
            "extra": "2872239 times\n2 procs"
          },
          {
            "name": "Benchmark_Intercept",
            "value": 438,
            "unit": "ns/op",
            "extra": "2725785 times\n2 procs"
          },
          {
            "name": "Benchmark_FanIn",
            "value": 1583,
            "unit": "ns/op",
            "extra": "766567 times\n2 procs"
          },
          {
            "name": "Benchmark_FanOut",
            "value": 884.4,
            "unit": "ns/op",
            "extra": "1361472 times\n2 procs"
          }
        ]
      }
    ]
  }
}