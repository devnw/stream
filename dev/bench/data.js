window.BENCHMARK_DATA = {
  "lastUpdate": 1640216563629,
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
      },
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
          "id": "ec5d60ee3f89f0e0e45a4e15d38d601d7e918177",
          "message": "Updating readme with docs and benchmark link",
          "timestamp": "2021-12-22T18:41:54-05:00",
          "tree_id": "32fc31d9ecae4c47aa6dc74841fffc49a82e05b1",
          "url": "https://github.com/devnw/stream/commit/ec5d60ee3f89f0e0e45a4e15d38d601d7e918177"
        },
        "date": 1640216563048,
        "tool": "go",
        "benches": [
          {
            "name": "Benchmark_Pipe",
            "value": 565.4,
            "unit": "ns/op",
            "extra": "2114622 times\n2 procs"
          },
          {
            "name": "Benchmark_Intercept",
            "value": 549,
            "unit": "ns/op",
            "extra": "2155120 times\n2 procs"
          },
          {
            "name": "Benchmark_FanIn",
            "value": 1940,
            "unit": "ns/op",
            "extra": "611397 times\n2 procs"
          },
          {
            "name": "Benchmark_FanOut",
            "value": 1123,
            "unit": "ns/op",
            "extra": "1000000 times\n2 procs"
          }
        ]
      }
    ]
  }
}