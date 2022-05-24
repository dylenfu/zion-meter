### zion-meter


#### compile
```bin bash
export ONROBOT=local
make compile
```

### how to simulate tps testing
```bin bash
make run group=60 user=20 last=3h
```

it denotes that there will be 120 group in this test case, and each group have 140 accumulate account.
